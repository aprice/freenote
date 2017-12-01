'use strict';

// IndexDBDataStore acts as a caching proxy between the UI and the REST API.
// Calls return a Promise, which will return the fastest result it can get -
// usually what's cached in the DB. It will also fire a "notefetched" or
// "summaryfetched" event on window when it fetches data from the API newer than
// what was in the DB.
var IndexedDBDataStore = {
	schemaVersion: 2,
	schemaName: "freenote",
	storeName: "notes",
	db: null,
	lastRefresh: new Date(2017, 1, 1, 0, 0, 0, 0),
	//syncFrequency: 1 * 60 * 1000,
	syncFrequency: 5 * 1000,
	syncInterval: null,
	hasOffline: true,

	init: function() {
		var self = this;
		var openReq = window.indexedDB.open(this.schemaName, this.schemaVersion);
		openReq.onerror = function (event) {
			console.log(event.target.error);
			DataStore = RESTDataStore;
		};
		openReq.onsuccess = function (event) {
			console.log("DB opened");
			self.db = event.target.result;
			// Update lastRefresh from max(notes.modified)
			self.getNoteStore("readonly").index("modified").openCursor(null, "prev").onsuccess = function(event) {
				var cursor = event.target.result;
				if (cursor) {
					self.lastRefresh = new Date(cursor.value.modified);
					console.log("Up to date as of "+self.lastRefresh);
				}
			}
			self.syncNotes();
			self.syncInterval = window.setInterval(function() {self.syncNotes();}, self.syncFrequency);
		};
		openReq.onupgradeneeded = function (event) {
			var db = event.target.result;
			var noteStore = db.createObjectStore(self.storeName, { keyPath: "id" });
			noteStore.createIndex("modified", "modified");
			noteStore.createIndex("path", "path");
			noteStore.createIndex("pending", "pending");
			noteStore.createIndex("delete", "delete");
		};
	},

	delete: function() {
		window.indexedDB.deleteDatabase(this.schemaName);		
	},

	getNoteStore: function(txMode) {
		return this.db.transaction(this.storeName, txMode).objectStore(this.storeName);
	},

	getNote: function(id) {
		var self = this;
		var request = this.getNoteStore("readonly").get(id);

		return new Promise(function(resolve,reject) {
			request.onsuccess = function () {
				if (isTempID(request.result.id)) {
					request.result.fetched = request.result.fetched || new Date();
					resolve(request.result);
				} else if (!request.result.fetched) {
					resolve(RESTDataStore.getNote(id));
				} else {
					RESTDataStore.getNote(id).then(function (note) {
						var event = new CustomEvent('notefetched', { detail: note });
						window.dispatchEvent(event);
					});
					resolve(request.result);
				}
			};

			request.onerror = function () {
				RESTDataStore.getNote(id).then(resolve,reject);
			};
		});
	},

	/*
	page = {
		folder: string
		start: int
		limit: int
	}
	*/
	listNotes: function(page) {
		page = page || {};
		var delay = page.start || 0;
		var idx = page.folder ? "path" : "modified";
		var kr = page.folder ? IDBKeyRange.only(page.folder) : null;
		var limit = page.limit || 0;
		var self = this;
		return new Promise(function(resolve,reject) {
			var notes = [];
			var scanner = self.getNoteStore("readonly").index(idx).openCursor(kr, "prev");
			scanner.onsuccess = function (event) {
				var cursor = event.target.result;
				if (cursor) {
					if (delay > 0) {
						delay--;
					} else {
						notes.push(cursor.value);
					}

					if (limit <= 0 || notes.length < limit) {
						cursor.continue();
					} else {
						console.log("Limit reached");
						resolve(notes);
					}
				} else {
					resolve(notes);
				}
			};

			scanner.onerror = function () {
				reject(request.error);
			};
		});
	},

	saveNote: function(note) {
		note.pending = 1;
		if (!(note.id)) {
			note.id = getTempID();
		}
		return this.requestPromise(this.getNoteStore("readwrite").put(note)).then(function () { return note });
	},

	deleteNote: function(note) {
		if (isTempID(note.id)) {
			return this.requestPromise(this.getNoteStore("readwrite").delete(note));
		} else {
			note.delete = 1;
			return this.requestPromise(this.getNoteStore("readwrite").put(note));
		}
	},

	syncNotes: function() {
		var u = User.user ? User.user.username : "not logged in";
		App.log("Online: " + navigator.onLine + ", user: "+u);
		if (!navigator.onLine || document.hidden || !User.user) {
			return;
		}

		var self = this;
		var store = this.getNoteStore("readonly");

		// Pending saves
		var kr = IDBKeyRange.only(1);
		store.index("pending").openCursor(kr).onsuccess = function (event) {
			var cursor = event.target.result;
			if (cursor) {
				var note = cursor.value;
				var tempID = false;
				if (isTempID(note.id)) {
					tempID = note.id;
					note.id = null;
				}
				RESTDataStore.saveNote(note).then(function(note) {
					note.fetched = new Date();
					self.getNoteStore("readwrite").put(note);
					if (tempID) {
						note.tempID = tempID;
						self.getNoteStore("readwrite").delete(tempID);
					}
					var event = new CustomEvent('notefetched', { detail: note });
					window.dispatchEvent(event);
				});
				cursor.continue();
			}
		}

		// Pending deletes
		var kr = IDBKeyRange.only(1);
		store.index("delete").openCursor(kr).onsuccess = function (event) {
			var cursor = event.target.result;
			if (cursor) {
				RESTDataStore.deleteNote(note).then(function(id) {
					self.getNoteStore("readwrite").delete(id);
				});
				cursor.continue();
			}
		}

		// Refresh notes
		var pg = appendQuery(User.user._links.notes.href, "modifiedSince=" + (self.lastRefresh.toISOString()));
		var notes = [];
		new Promise(function(resolve, reject) {
			var handler = function (payload) {
				if (payload.notes) {
					Array.prototype.push.apply(notes, payload.notes);
				}
				if (payload._links.next) {
					return restRequest({ link: payload._links.next }).then(handler);
				} else {
					return notes;
				}
			};
			resolve(restRequest({ url: pg }).then(handler));
		}).then(function() {
			var maxMod = self.lastRefresh;
			var tx = self.getNoteStore("readwrite");
			for (var i = 0; i < notes.length; i++) {
				var note = notes[i];
				if (note.modified > maxMod) {
					maxMod = note.modified;
				}
				(function (note) {
					var request = tx.get(note.id)
					request.onsuccess = function () {
						var payload = request.result;
						if (!payload || payload.modified < note.modified) {
							tx.put(note);
						}
					}
				})(note)
			}
			tx.onsuccess = function () {
				self.lastRefresh = new Date(maxMod);
				console.log("Up to date as of " + self.lastRefresh);
				var event = new CustomEvent('summaryfetched', { detail: notes });
				window.dispatchEvent(event);
			}
		}).catch(App.debug);
	},

	requestPromise: function(request) {
		return new Promise(function (resolve, reject) {
			request.onsuccess = function () {
				resolve(request.result);
			};

			request.onerror = function () {
				reject(request.error);
			};
		});
	}
};

var RESTDataStore = {
	hasOffline: false,

	init: function() {

	},

	delete: function() {

	},

	getNote: function(id) {
		return restRequest({url: User.user._links.notes.href + '/' + id});
	},

	/*
	page = {
		folder: string
		start: int
		limit: int
	}
	*/
	listNotes: function (page) {
		page = page || {};
		var url = User.user._links.notes.href;
		if (page.folder) {
			url = appendQuery(url, "folder=" + page.folder);
		}
		if (page.start) {
			url = appendQuery(url, "start=", page.start);
		}
		if (page.limit) {
			url = appendQuery(url, "length=", page.limit);
		}
		return restRequest({url: url});
	},

	saveNote: function(note) {
		var ln = note.id ? note._links.save : User.user._links.createnote;
		note._links = null;
		return restRequest({
			link: ln,
			payload: note
		}).then(function(payload) {
			payload.fetched = new Date();
			return payload;
		});
	},

	deleteNote: function(note) {
		return restRequest({link: note._links.delete}).then(function() {
			return note.id;
		});
	}
};

RESTDataStore.init();
var DataStore = IndexedDBDataStore;
if (window.indexedDB) {
	IndexedDBDataStore.init();
} else {
	console.log("IndexDB not supported, offline storage disabled");
	DataStore = RESTDataStore;
}
