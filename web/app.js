// core app state & helpers
var App = {
	debug: true,
	user: null,
	notes: null,
	currentNote: null,
	mode: "md",
	prevPage: null,
	nextPage: null,
	curPage: null,
	createLink: null,
	refreshInterval: null,
	refreshFrequency: 5 * 60 * 1000,
	saveInterval: null,
	saveFrequency: 1 * 60 * 1000,

	init: function () {
		this.userPanel = $id("User");
		this.loginPanel = $id("Login");
		this.messagePanel = $id("StatusMessage");
		this.noteManager = $id("NoteManager");
		this.noteList = $id("NoteList");
		this.noteTitle = $id("NoteTitle");
		this.noteBody = $id("NoteBody");
		this.confirmationPanel = $id("Confirmation");

		if (typeof(Version) != "undefined") {
			$id("AppVersion").innerText = Version;
		}
	},

	userRefresh: function () {
		if (this.user == null) {
			this.userPanel.style.display = "none";
			this.noteManager.style.display = "none";
			this.loginPanel.style.display = "block";
		} else {
			$id("WhoAmI").innerText = this.user.username;
			this.loginPanel.style.display = "none";
			this.userPanel.style.display = "block";
		}
	},

	noteListRefresh: function() {
		if (this.user == null) {
			this.notes = null;
			this.currentNote = null;
			this.noteManager.style.display = "none";
			return;
		}
		while(this.noteList.hasChildNodes()) {
			this.noteList.removeChild(this.noteList.firstChild);
		}
		if (this.notes == null || this.notes.length == 0) return;
		for (var i = 0; i < this.notes.length; i++) {
			this.notes[i].index = i;
			var note = this.notes[i];
			var el = $el('<li id="note-'+note.id+'"><span class="noteTitle">'+note.title+'</span><span class="noteModified">'+(new Date(note.modified)).toLocaleString()+'</span></li>');
			if (note == this.currentNote) {
				el.classList.add("selected");
				this.currentNote.index = i;
			}
			this.noteList.appendChild(el);
			el.note = note;
			el.addEventListener("click", function(evt){SelectNote(this.note)});
		}
		this.noteManager.style.display = "flex";
		$("#NoteListPager .toolButton", this.noteManager).forEach(function(v){
			v.classList.add("disabled");
		});
		if (this.prevPage) {
			$id("PrevPageButton").classList.remove("disabled");
		}
		if (this.nextPage) {
			$id("NextPageButton").classList.remove("disabled");
		}
	},

	noteRefresh: function() {
		if (this.user == null || this.currentNote == null) {
			this.noteManager.classList.add("noneSelected");
			this.noteManager.classList.remove("noteSelected");
			this.noteTitle.innerHTML = "";
			this.noteBody.innerHTML = "";
			document.title = "freenote";
			return;
		}
		this.noteManager.classList.add("noteSelected");
		this.noteManager.classList.remove("noneSelected");
		this.noteTitle.innerText = this.currentNote.title;
		document.title = this.currentNote.title + " - freenote";
		if (this.currentNote.id)
			$1("#NoteID", this.noteManager).innerText = uuidToB64(this.currentNote.id);

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

	isModified: function() {
		if (!this.currentNote) return false;
		else return this.currentNote.title != this.noteTitle.innerText
			|| (this.mode == "md" && this.currentNote.body != this.noteBody.innerText)
			|| (this.mode == "html" && this.currentNote.html != this.noteBody.innerHTML);
	},

	log: function (data) {
		if (this.debug) {
			console.log(data);
		}
	},

	message: function (message, isHtml = false) {
		var mp = this.messagePanel;
		if (isHtml) mp.innerHTML = message;
		else mp.innerText = message;
		mp.cssClass = "";
		mp.style.display = "block";
		setTimeout(function () { fade(mp, 800) }, 3000);
	},

	error: function (message, isHtml = false) {
		var mp = this.messagePanel;
		if (isHtml) mp.innerHTML = message;
		else mp.innerText = message;
		mp.cssClass = "error";
		mp.style.display = "block";
		setTimeout(function () { fade(mp, 800) }, 3000 );
	},

	confirm: function(message, isHtml = false, confirmCB = null, cancelCB = null) {
		var cp = this.confirmationPanel;
		if (isHtml) $1(".message", cp).innerHTML = message;
		else $1(".message", cp).innerText = message;
		cp.style.display = "block";
		$1(".confirmButton", cp).onclick = function() {
			fade(cp, 600);
			if (confirmCB) confirmCB();
		};
		$1(".cancelButton", cp).onclick = function () {
			fade(cp, 600);
			if (cancelCB) cancelCB();
		};
	},

	rest: function(options) {
		var r = new XMLHttpRequest();
		if (options.link) {
			options.method = options.link.method;
			options.url = options.link.href;
		}
		options.method = options.method || "GET";
		r.open(options.method, options.url, true);
		r.setRequestHeader("Accept", "application/json");
		r.onreadystatechange = function () {
			if (r.readyState != 4) return;
			if (r.status < 400) {
				App.log("Success: " + r.responseText);
				var payload = null;
				if (r.responseText && r.getResponseHeader("Content-Type") == "application/json") {
					payload = JSON.parse(r.responseText);
				}
				if (options.success) options.success(payload);
			} else {
				if (options.failed) options.failed(r.status);
				else if (!options.hideError) App.error("Request failed: " + r.responseText);
			}
			if (options.finally) options.finally();
		};
		if (options.payload) {
			if (options.payload['_links']) delete (options.payload['_links']);
			options.body = JSON.stringify(options.payload);
		}
		options.ctype = options.ctype || "application/json";
		if (options.body) r.setRequestHeader("Content-Type", options.ctype);
		App.log("Sending " + options.method + " " + options.url + ": " + JSON.stringify(options.body));
		r.send(options.body);
	},

	userPanel: null,
	loginPanel: null,
	messagePanel: null,
	confirmationPanel: null,
	noteManager: null,
	noteList: null,
	noteTitle: null,
	noteBody: null
};

window.addEventListener("load", function () {
	App.init();
});