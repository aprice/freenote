'use strict';

var NoteList = {
	noteManager: null,
	noteList: null,
	notes: null,

	init: function () {
		this.noteManager = $id("NoteManager");
		this.noteList = $id("NoteList");

		var self = this;
		$id("NewNoteButton").addEventListener("click", function(evt) {
			self.newNote();
		});	
		$id("BackButton").addEventListener("click", function(evt) {
			self.selectNote(null);
		});

		window.addEventListener("summaryfetched", function(evt) {
			App.log("Note summaries fetched");
			self.loadNotes();
		});

		window.addEventListener("notefetched", function (event) {
			var note = event.detail;
			if (note.tempID) {
				var noteEl = $id("note-"+note.tempID);
				if (noteEl == null) return;
				noteEl.id = "note-"+note.id;
				noteEl.note = note;
			}
		});

		this.loadNotes();
	},

	loadNotes: function() {
		if (!User.loggedIn()) return;
		var self = this;
		DataStore.listNotes().then(function (notes) {
			self.notes = notes;
			self.refresh();
		}).catch(function (err) {
			console.log("Getting notes failed: " + err);
			if (status == 404) {
				self.notes = null;
				self.refresh();
				NoteEditor.selectNote(null);
			}
		});
	},

	refresh: function() {
		if (!User.loggedIn()) {
			this.notes = null;
			hide(this.noteManager);
			document.title = "freenote";
			return;
		}
		var self = this;
		var noteElements = this.noteList.children;
		var noteEl = this.noteList.hasChildNodes() ? this.noteList.children[0] : null;
		this.notes.sort(function(lhs,rhs) {
			if (lhs.modified < rhs.modified) return 1;
			else if (lhs.modified > rhs.modified) return -1;
			else if (lhs.id < rhs.id) return 1;
			else if (lhs.id > rhs.id) return -1;
			else return 0;
		});
		for (var i = 0; i < this.notes.length; i++) {
			var note = this.notes[i];
			if (noteEl == null) {
				// Nothing to compare to, append
				var newNote = this.buildNote(note);
				this.noteList.appendChild(newNote);
			} else if (note.id == noteEl.note.id && note.modified == noteEl.note.modified) {
				// Unmodified, skip!
				noteEl = noteEl.nextElementSibling;
			} else if (note.id == noteEl.note.id) {
				// Update in place
				$1(".noteTitle", noteEl).innerText = note.title;
				$1(".noteModified", noteEl).innerText = (new Date(n.modified)).toLocaleString();
				noteEl = noteEl.nextElementSibling;
			} else if (note.modified >= noteEl.note.modified) {
				// Newer than existing, insert in front
				var newNote = this.buildNote(note);
				noteEl.insertAdjacentElement("beforebegin", newNote);
			} else if (note.modified < noteEl.note.modified) {
				// Older than existing, remove existing
				var temp = noteEl;
				noteEl = noteEl.nextElementSibling;
				this.noteList.removeChild(temp);
				i--;
			}
		}
		while(noteEl != null) {
			var temp = noteEl.nextElementSibling;
			this.noteList.removeChild(noteEl);
			noteEl = temp;
		}
		show(this.noteManager);
	},

	buildNote: function(n) {
		var el = $el('<li id="note-'+n.id+'"><span class="noteTitle">'+n.title+'</span><span class="noteModified">'+(new Date(n.modified)).toLocaleString()+'</span></li>');
		el.note = n;
		var self = this;
		el.addEventListener("click", function(evt){self.selectNote(this.note)});
		return el;
	},

	selectNote: function(note) {
		var self = this;
		NoteEditor.saveNote().then(function(){
			// Deal w/unfetched notes
			if (NoteEditor.currentNote != null) {
				var noteEl = $id("note-"+NoteEditor.currentNote.id);
				if (noteEl != null) noteEl.classList.remove("selected");
			}
			NoteEditor.changeNote(note);
			$id("note-"+note.id).classList.add("selected");
		})
	},

	newNote: function() {
		var self = this;
		NoteEditor.saveNote().then(function () {
			var now = new Date().toISOString();
			var newNote = {
				id: getTempID(),
				title: "Untitled note",
				body: "",
				html: "",
				created: now,
				modified: now
			};
			self.notes.unshift(newNote);
			NoteEditor.changeNote(newNote);
			self.refresh();
		});
	},
};

var NoteEditor = {
	currentNote: null,
	mode: "md",
	noteEditor: null,
	noteTitle: null,
	noteBody: null,

	init: function () {
		this.noteEditor = $id("NoteEditor");
		this.noteTitle = $id("NoteTitle");
		this.noteBody = $id("NoteBody");
		
		var self = this;
		$id("SaveButton").addEventListener("click",function(evt) {
			self.saveNote();
		});
		$id("SourceButton").addEventListener("click", function(evt) {
			self.switchSourceView();
		});
		$id("DeleteButton").addEventListener("click", function(evt) {
			self.deleteNote();
		});
		$id("CopyIDButton").addEventListener("click", function(evt) {
			self.copyNoteID();
		});

		// Make links in HTML editing mode ctrl-clickable
		window.addEventListener("keydown", function (evt) {
			if (self.currentNote && self.mode == "html" && evt.ctrlKey) {
				$each("a", self.noteBody, function (v) {
					v.contentEditable = false;
				});
			}
		});
		window.addEventListener("keyup", function (evt) {
			if (self.currentNote && self.mode == "html" && !evt.ctrlKey) {
				$each("a", self.noteBody, function (v) {
					v.contentEditable = 'inherit';
				});
			}
		});

		$id("NoteTitle").addEventListener("change", function () {
			var title = self.noteTitle.innerText;
			self.noteTitle.innerHTML = title;
			self.currentNote.title = title;
			DataStore.saveNote(App.currentNote).then(function () {
				NoteList.loadNotes();
			})
		});

		// Save on lose visibility, refresh note list on regain visibility
		document.addEventListener("visibilitychange", function () {
			if (document.hidden) {
				//window.clearInterval(App.refreshInterval);
				//window.clearInterval(App.saveInterval);
				self.saveNote();
			} else {
				NoteList.loadNotes();
				//App.refreshInterval = window.setInterval(LoadNotes, App.refreshFrequency);
				//App.saveInterval = window.setInterval(SaveNote, App.saveFrequency);
			}
		});

		window.addEventListener("notefetched", function (event) {
			var note = event.detail;
			if ((self.currentNote.id == note.id || self.currentNote.id == note.tempID)
				&& (self.currentNote.modified < note.modified || !self.currentNote.fetched)) {
				App.log("Selected note fetched!");
				self.currentNote = note;
				if (self.isModified()
					&& (self.currentNote.title != note.title
						|| (self.mode == "md" && self.currentNote.body != note.body)
						|| (self.mode == "html" && self.currentNote.html != note.html))) {
					console.log("Conflict!!");
				}
				self.refresh();
			}
		});
	},

	changeNote: function(note) {
		this.currentNote = note;
		this.refresh();
	},

	refresh: function() {
		if (!User.loggedIn() || this.currentNote == null) {
			this.noteTitle.innerHTML = "";
			this.noteBody.innerHTML = "";
			NoteList.noteManager.classList.remove("noteSelected");
			NoteList.noteManager.classList.add("noneSelected");
			document.title = "freenote";
			return;
		}
		NoteList.noteManager.classList.remove("noneSelected");
		NoteList.noteManager.classList.add("noteSelected");
		document.title = this.currentNote.title + " - freenote";

		this.noteTitle.innerText = this.currentNote.title;
		if (this.currentNote.id) {
			$1("#NoteID", this.noteManager).innerText = uuidToB64(this.currentNote.id);
			show($1("#NoteID", this.noteManager));
		} else {
			hide($1("#NoteID", this.noteManager));
		}

		if (this.mode == "html") {
			$1("#SourceButton i.html-mode", this.noteManager).classList.add("selected");
			$1("#SourceButton i.md-mode", this.noteManager).classList.remove("selected");
			this.noteBody.innerHTML = this.currentNote.html;
			this.currentNote.html = this.noteBody.innerHTML;
		} else {
			$1("#SourceButton i.md-mode", this.noteManager).classList.add("selected");
			$1("#SourceButton i.html-mode", this.noteManager).classList.remove("selected");
			this.noteBody.innerText = this.currentNote.body;
		}
		this.noteBody.classList.remove("md-mode", "html-mode");
		this.noteBody.classList.add(this.mode+"-mode");
	},

	saveNote: function() {
		var self = this;
		return new Promise(function(resolve,reject) {
			if (self.currentNote == null || !self.isModified()) {
				resolve(null);
				return;
			}
			show($1("#SaveButton .loader", self.noteEditor));
			var note = self.currentNote;
			note.title = self.noteTitle.innerText;
			note.body = "";
			note.html = "";
			note.modified = new Date();
			if (self.mode == "md") note.body = self.noteBody.innerText;
			else note.html = self.noteBody.innerHTML;
			DataStore.saveNote(note).then(function (note) {
				self.currentNote = note;
				self.refresh();
				NoteList.loadNotes();
				hide($1("#SaveButton .loader", self.noteEditor));
				resolve(note);
			}).catch(reject);
		});
	},

	deleteNote: function() {
		note = note || this.currentNote;
		App.confirm("Are you sure you want to delete the note \""+note.title+"\"?", false, function () {
			DataStore.deleteNote(note).then(function () {
				this.selectNote(null);
				NoteList.loadNotes();
			});
		});
	},

	isModified: function() {
		if (!this.currentNote) return false;
		else return this.currentNote.title != this.noteTitle.innerText
			|| (this.mode == "md" && this.currentNote.body != this.noteBody.innerText)
			|| (this.mode == "html" && this.currentNote.html != this.noteBody.innerHTML);
	},

	switchSourceView: function() {
		if (this.currentNote == null) return;
		var self = this;
		this.saveNote().then(function() {
			if (self.mode == "md") {
				self.mode = "html";
			} else {
				self.mode = "md";
			}
			self.refresh();
		});
	},

	copyNoteID: function() {
		var textArea = document.createElement("textarea");
		textArea.value = uuidToB64(this.currentNote.id);
		document.body.appendChild(textArea);
		textArea.select();
		try {
			document.execCommand('copy');
		} catch (err) {
			console.log('copy command failed: ' + err);
		}
		document.body.removeChild(textArea);
	},
};
