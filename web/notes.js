function LoadNotes() {
	App.rest({
		url: "/users/" + App.user.id + "/notes",
		success: function(payload) {
			App.notes = payload.notes;
			App.noteListRefresh();
		}
	});
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
			url: "/users/" + App.user.id + "/notes/" + note.id,
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
	if (App.currentNote == null) {
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
		method: note.id ? "PUT" : "POST",
		url: note.id
			? "/users/" + App.user.id + "/notes/" + note.id
			: "/users/" + App.user.id + "/notes",
		payload: note,
		success: function(payload) {
			App.currentNote = payload;
			App.currentNote.index = note.index;
			App.notes[note.index] = App.currentNote;
			App.noteRefresh();
			App.noteListRefresh();
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
		App.notes.push(App.currentNote);
		App.noteRefresh();
		App.noteListRefresh();
	});
}

function DeleteNote() {
	// TODO: confirmation modal
	App.confirm("Are you sure you want to delete this note?", false, function() {
		App.rest({
			method: "DELETE",
			url: "/users/" + App.user.id + "/notes/" + App.currentNote.id,
			success: function (payload) {
				App.notes.splice(App.currentNote.index, 1);
				App.currentNote = null;
				App.noteRefresh();
				App.noteListRefresh();
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
