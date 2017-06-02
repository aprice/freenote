// general-purpose utilities
var $id = document.getElementById.bind(document);
function $(qry,el=null) {
	el = el || document;
	return el.querySelectorAll(qry);
}
function $1(qry,el=null) {
	return $(qry,el)[0];
}
function fade(el,ms,step) {
	if (!step) el.style.opacity = 1;
	if ((el.style.opacity -= .05) <= 0) {
		el.style.display = "none";
		el.style.opacity = 1;
	} else {
		setTimeout(function(){fade(el,ms,true)}, ms * .05);
	}
}
function $el(html) {
	var tpl = document.createElement('template');
	tpl.innerHTML = html;
	return tpl.content.firstChild;
}

// core app state & helpers
var App = {
	debug: true,
	user: null,
	notes: null,
	currentNote: null,
	mode: "md",

	init: function () {
		this.userPanel = $id("User");
		this.loginPanel = $id("Login");
		this.messagePanel = $id("StatusMessage");
		this.noteManager = $id("NoteManager");
		this.noteList = $id("NoteList");
		this.noteTitle = $id("NoteTitle");
		this.noteBody = $id("NoteBody");
		this.confirmationPanel = $id("Confirmation");
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
			var el = $el('<li id="note-'+note.id+'">'+note.title+'</li>');
			if (note == this.currentNote) {
				el.classList.add("selected");
				this.currentNote.index = i;
			}
			this.noteList.appendChild(el);
			el.note = note;
			el.addEventListener("click", function(evt){SelectNote(this.note)});
		}
		this.noteManager.style.display = "flex";
	},

	noteRefresh: function() {
		if (this.user == null || this.currentNote == null) {
			this.noteManager.classList.add("noneSelected");
			this.noteManager.classList.remove("noteSelected");
			this.noteTitle.innerHTML = "";
			this.noteBody.innerHTML = "";
			return;
		}
		this.noteManager.classList.add("noteSelected");
		this.noteManager.classList.remove("noneSelected");
		this.noteTitle.innerText = this.currentNote.title;
		if (this.mode == "html") {
			$1("#SourceButton i.html-mode", this.noteManager).classList.add("selected");
			$1("#SourceButton i.md-mode", this.noteManager).classList.remove("selected");
			this.noteBody.innerHTML = this.currentNote.html;
		} else {
			$1("#SourceButton i.md-mode", this.noteManager).classList.add("selected");
			$1("#SourceButton i.html-mode", this.noteManager).classList.remove("selected");
			this.noteBody.innerText = this.currentNote.body;
		}
		this.noteBody.classList.remove("md-mode", "html-mode");
		this.noteBody.classList.add(App.mode+"-mode");
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
		setTimeout(function () { fade(mp, 800), 3000 });
	},

	error: function (message, isHtml = false) {
		var mp = this.messagePanel;
		if (isHtml) mp.innerHTML = message;
		else mp.innerText = message;
		mp.cssClass = "error";
		mp.style.display = "block";
		setTimeout(function () { fade(mp, 800), 3000 });
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
				App.error("Request failed: " + r.responseText);
				if (options.failed) options.failed();
			}
			if (options.finally) options.finally();
		};
		if (options.payload) options.body = JSON.stringify(options.payload);
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
	App.message("Initialized");
});