package notes

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// WelcomeNote creates a new welcome note for the given user.
func WelcomeNote(userID uuid.UUID) Note {
	return Note{
		ID:       uuid.NewV4(),
		Owner:    userID,
		Title:    welcomeNoteTitle,
		Body:     welcomeNoteBody,
		Created:  time.Now(),
		Modified: time.Now(),
	}
}

const welcomeNoteTitle = "Welcome to Freenote"

const welcomeNoteBody = `Welcome to Freenote!

You can put your notes here, if you want. If you're using the web app, you'll find a list of your notes on the left, and you can select one to view it on the right. Tap or click the title or body to edit. On mobile, selecting a note will take over the screen; tap the back arrow next to the title to go back to the list.

The plus at the top of the note list creates a new note. The arrows at the bottom switch between pages of notes in the list.

The toolbar at the top of the note editor lets you save, switch between markdown and WYSIWYG, or delete the note.`
