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
	App.currentNote = note;
	App.noteRefresh();
	App.noteListRefresh();
}

function SaveNote() {
	App.currentNote.title = App.noteTitle.innerText;
	App.currentNote.body = App.noteBody.innerText;
	var r = new XMLHttpRequest();
	r.open("PUT", "/users/" + App.user.id + "/notes/" + App.currentNote.id, true);
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
		} else {
			App.error("Getting note list failed: " + r.responseText);
		}
	};
	App.currentNote._links = null;
	r.send(JSON.stringify(App.currentNote));
}
