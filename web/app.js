var $ = document.querySelectorAll.bind(document);
var $id = document.getElementById.bind(document);
function fade(el,ms,step) {
	if (!step) el.style.opacity = 1;
	if ((el.style.opacity -= .05) <= 0)
		el.style.display = "none";
	else
		setTimeout(function(){fade(el,ms,true)}, ms * .05);
}
function $el(html) {
	var tpl = document.createElement('template');
	tpl.innerHTML = html;
	return tpl.content.firstChild;
}

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
			this.noteTitle.innerHTML = "";
			this.noteBody.innerHTML = "";
			return;
		}
		this.noteTitle.innerText = this.currentNote.title;
		if (this.mode == "html") {
			this.noteManager.querySelectorAll("#SourceButton i.html-mode")[0].classList.add("selected");
			this.noteManager.querySelectorAll("#SourceButton i.md-mode")[0].classList.remove("selected");
			this.noteBody.innerHTML = this.currentNote.html;
		} else {
			this.noteManager.querySelectorAll("#SourceButton i.md-mode")[0].classList.add("selected");
			this.noteManager.querySelectorAll("#SourceButton i.html-mode")[0].classList.remove("selected");
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
		if (isHtml) this.messagePanel.innerHTML = message;
		else this.messagePanel.innerText = message;
		this.messagePanel.cssClass = "";
		this.messagePanel.style.display = "block";
		this.fadeMessage();
	},

	error: function (message, isHtml = false) {
		if (isHtml) this.messagePanel.innerHTML = message;
		else this.messagePanel.innerText = message;
		this.messagePanel.cssClass = "error";
		this.messagePanel.style.display = "block";
		this.fadeMessage();
	},

	fadeMessage: function() {
		var s = this.messagePanel.style;
		var fade = function() {(s.opacity -= .05) <= 0 ? s.display = "none" : setTimeout(fade, 40)};
		s.opacity = 1;
		setTimeout(function(){fade(this.messagePanel, 800), 3000});
	},

	userPanel: null,
	loginPanel: null,
	messagePanel: null,
	noteManager: null,
	noteList: null,
	noteTitle: null,
	noteBody: null
};

window.addEventListener("load", function () {
	App.init();
	App.message("Initialized");
});