function LoadNotes(pg,hideError) {
	if (!App.user) return;
	if (typeof(pg) == "boolean") {
		hideError = pg;
		pg = null;
	}
	pg = pg || App.curPage || App.user._links.notes.href;
	App.rest({
		url: pg,
		success: function(payload) {
			App.notes = payload.notes;
			App.prevPage = null;
			if (payload._links.previous) {
				App.prevPage = payload._links.previous.href;
			}
			App.nextPage = null;
			if (payload._links.next) {
				App.nextPage = payload._links.next.href;
			}
			App.curPage = payload._links.canonical.href;
			App.createLink = payload._links.create;
			App.noteListRefresh();
		},
		failed: function(status) {
			if (status == 404) {
				App.notes = null;
				App.nextPage = null;
				App.prevPage = null;
				App.curPage = null;
				App.currentNote = null;
				App.noteListRefresh();
			}
		}
	});
}

function NextPage() {
	if (App.nextPage) {
		LoadNotes(App.nextPage);
	}
}

function PrevPage() {
	if (App.prevPage) {
		LoadNotes(App.prevPage);
	}
}

function SelectNote(note) {
	SaveNote(function() {
		if (note == null) {
			App.currentNote = null;
			App.noteRefresh();
			App.noteListRefresh();
			return;
		}
		App.rest({
			url: note._links.canonical.href,
			success: function(payload) {
				App.currentNote = payload;
				App.currentNote.index = note.index;
				App.notes[note.index] = App.currentNote;
				App.noteRefresh();
				App.noteListRefresh();
			}
		});
	});
}

function SaveNote(cb) {
	if (App.currentNote == null || !App.isModified()) {
		if (cb) cb();
		return;
	}
	var spinner = $1("#SaveButton .loader", App.noteManager);
	spinner.style.display = "block";
	var note = App.currentNote;
	note.title = App.noteTitle.innerText;
	note.body = "";
	note.html = "";
	note.modified = new Date();
	if (App.mode == "md") note.body = App.noteBody.innerText;
	else note.html = App.noteBody.innerHTML;
	App.rest({
		link: note.id ? note._links.save : App.createLink,
		payload: note,
		success: function(payload) {
			App.currentNote = payload;
			App.currentNote.index = note.index;
			App.notes[note.index] = App.currentNote;
			if (cb) cb();
		},
		finally: function() {
			fade(spinner, 300);
		}
	});
}

function NewNote() {
	SaveNote(function() {
		App.currentNote = {
			index: App.notes.length,
			title: "Untitled note",
			body: "",
			html: "",
			created: new Date(),
			modified: new Date()
		};
		App.notes.unshift(App.currentNote);
		App.noteRefresh();
		App.noteListRefresh();
	});
}

function DeleteNote() {
	App.confirm("Are you sure you want to delete this note?", false, function() {
		App.rest({
			link: App.currentNote._links.delete,
			success: function (payload) {
				App.notes.splice(App.currentNote.index, 1);
				SelectNote(null);
				LoadNotes();
			}
		});
	});
}

function SwitchSourceView() {
	if (App.currentNote == null) return;
	SaveNote(function(){
		if (App.mode == "md") {
			App.mode = "html";
		} else {
			App.mode = "md";
		}
		App.noteRefresh();
	});
}

function CopyNoteID() {
	var textArea = document.createElement("textarea");
	/* Some theories this is necessary but leaving out until proven
	textArea.style.position = 'fixed';
	textArea.style.top = 0;
	textArea.style.left = 0;
	textArea.style.width = '2em';
	textArea.style.height = '2em';
	textArea.style.padding = 0;
	textArea.style.border = 'none';
	textArea.style.outline = 'none';
	textArea.style.boxShadow = 'none';
	textArea.style.background = 'transparent';
	*/
	textArea.value = uuidToB64(App.currentNote.id);
	document.body.appendChild(textArea);
	textArea.select();
	try {
		document.execCommand('copy');
	} catch (err) {
		console.log('copy command failed: '+err);
	}
	document.body.removeChild(textArea);
}

window.addEventListener("load", function () {
	// Make links in HTML editing mode ctrl-clickable
	window.addEventListener("keydown", function (evt) {
		if (App.currentNote && App.mode == "html" && evt.ctrlKey) {
			$each("#NoteBody a", function (v) {
				v.contentEditable = false;
			});
		}
	});
	window.addEventListener("keyup", function (evt) {
		if (App.currentNote && App.mode == "html" && !evt.ctrlKey) {
			$each("#NoteBody a", function (v) {
				v.contentEditable = 'inherit';
			});
		}
	});

	// Save on lose visibility, refresh note list on regain visibility
	document.addEventListener("visibilitychange", function () {
		if (document.hidden) {
			window.clearInterval(App.refreshInterval);
			window.clearInterval(App.saveInterval);
			SaveNote();
		} else {
			LoadNotes();
			App.refreshInterval = window.setInterval(LoadNotes, App.refreshFrequency);
			App.saveInterval = window.setInterval(SaveNote, App.saveFrequency);
		}
	});

	// Refresh note list on regular interval
	App.refreshInterval = window.setInterval(LoadNotes, App.refreshFrequency);

	// Save note on regular interval
	App.saveInterval = window.setInterval(SaveNote, App.saveFrequency);
});
