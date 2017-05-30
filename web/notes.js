function LoadNotes() {
	var r = new XMLHttpRequest();
	r.open("GET", "/users/"+App.user.id+"/notes", true);
	r.onreadystatechange = function () {
		if (r.readyState != 4) return;
		if (r.status < 400) {
			App.log("Success: " + r.responseText);
			App.notes = JSON.parse(r.responseText).notes;
			App.noteListRefresh();
		} else {
			App.error("Getting note list failed: " + r.responseText);
		}
	};
	r.send();
}

function SelectNote(note) {
	SaveNote(function() {
		var r = new XMLHttpRequest();
		r.open("GET", "/users/" + App.user.id + "/notes/" + note.id, true);
		r.onreadystatechange = function () {
			if (r.readyState != 4) return;
			if (r.status < 400) {
				App.log("Success: " + r.responseText);
				App.currentNote = JSON.parse(r.responseText);
				App.currentNote.index = note.index;
				App.notes[note.index] = App.currentNote;
				App.noteRefresh();
				App.noteListRefresh();
			} else {
				App.error("Getting note list failed: " + r.responseText);
			}
		};
		r.send();
	});
}

function SaveNote(cb) {
	if (App.currentNote == null) {
		if (cb) cb();
		return;
	}
	var spinner = App.noteManager.querySelectorAll("#SaveButton .loader")[0];
	spinner.style.display = "block";
	var r = new XMLHttpRequest();
	if (App.currentNote.id)
		r.open("PUT", "/users/" + App.user.id + "/notes/" + App.currentNote.id, true);
	else
		r.open("POST", "/users/" + App.user.id + "/notes", true);
	r.setRequestHeader("Content-Type", "application/json");
	r.onreadystatechange = function () {
		if (r.readyState != 4) return;
		if (r.status < 400) {
			App.log("Success: " + r.responseText);
			idx = App.currentNote.index;
			App.currentNote = JSON.parse(r.responseText);
			App.currentNote.index = idx;
			App.notes[idx] = App.currentNote;
			App.noteRefresh();
			App.noteListRefresh();
			if (cb) cb();
		} else {
			App.error("Getting note list failed: " + r.responseText);
		}
		fade(spinner, 400);
	};
	App.currentNote.title = App.noteTitle.innerText;
	App.currentNote.body = "";
	App.currentNote.html = "";
	if (App.mode == "md") App.currentNote.body = App.noteBody.innerText;
	else App.currentNote.html = App.noteBody.innerHTML;
	App.currentNote._links = null;
	App.log("Saving as "+App.mode);
	r.send(JSON.stringify(App.currentNote));
}

function NewNote() {
	SaveNote(function() {
		App.currentNote = {
			index: App.notes.length,
			title: "Untitled note",
			body: "",
			html: ""
		};
		App.notes.push(App.currentNote);
		App.noteRefresh();
		App.noteListRefresh();
	});
}

function DeleteNote() {
	// TODO: confirmation modal

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
