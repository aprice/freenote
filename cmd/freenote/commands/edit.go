package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/rest"
)

var (
	editor string
)

func init() {
	editCmd.Flags().StringVarP(&editor, "editor", "e", "", "text editor (default auto-detect from $EDITOR et al)")
	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit [noteID]",
	Short: "Edit a note",
	Long: `
freenote edit will download a note from the Freenote server for local editing.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || len(args[0]) == 0 {
			return errors.New("You must provide a note ID to edit")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		tmpDir := filepath.Join(os.TempDir(), "freenote")
		err := os.MkdirAll(tmpDir, 0600)
		if err != nil && err != os.ErrExist {
			fmt.Println("failed to create temp dir ", tmpDir, ":", err)
			os.Exit(1)
		}

		tmpFile := filepath.Join(tmpDir, args[0]+".md")
		f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			fmt.Println("failed to open ", tmpFile, ":", err)
			os.Exit(1)
		}
		defer os.Remove(tmpFile)

		note := new(notes.Note)
		c, err := initClient()
		if err != nil {
			fmt.Println("failed to connect: ", err)
			os.Exit(1)
		}
		ownerID := c.User.ID
		err = c.Get(fmt.Sprintf("/users/%s/notes/%s", ownerID, url.QueryEscape(args[0])), note)
		if err != nil {
			fmt.Println("get note failed: ", err)
			os.Exit(1)
		}
		body := note.Body
		_, err = fmt.Fprint(f, body)
		if err != nil {
			fmt.Println("failed to write file: ", err)
			os.Exit(1)
		}

		err = f.Close()
		if err != nil {
			fmt.Println("failed to write file: ", err)
			os.Exit(1)
		}

		if editor == "" {
			editor = getEditor()
			if editor == "" {
				fmt.Println("no editor set")
				os.Exit(1)
			}
		}

		command := exec.Command(editor, tmpFile) // nolint: gas
		command.Stderr = os.Stderr
		command.Stdout = os.Stdout
		command.Stdin = os.Stdin
		err = command.Run()
		if err != nil {
			fmt.Println("editor returned error editing ", tmpFile, ":", err)
			os.Exit(1)
		}

		bodyBytes, err := ioutil.ReadFile(tmpFile)
		if err != nil {
			fmt.Println("unable to read ", uploadFile, ":", err)
			os.Exit(1)
		}
		note.Body = string(bodyBytes)
		note.HTMLBody = ""
		note.Modified = time.Now()

		resPL := new(struct {
			Links struct {
				Canonical rest.Link `json:"canonical"`
			} `json:"_links"`
		})
		c.Send("PUT",
			fmt.Sprintf("/users/%s/notes/%s", ownerID, args[0]),
			&note, resPL)
		if err != nil {
			fmt.Println("note upload failed: ", err)
			os.Exit(1)
		}
		fmt.Println("Upload successful.")
		fmt.Println("Note URL: ", resPL.Links.Canonical.Href)
	},
}

var possibleEditors = []string{
	"editor",
	"sensible-editor",
	"vim",
	"emacs",
	"nano",
	"notepad",
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	for _, poss := range possibleEditors {
		if editor, err := exec.LookPath(poss); err == nil {
			return editor
		}
	}
	return ""
}
