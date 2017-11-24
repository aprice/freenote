package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/rest"
)

var (
	backupHTML bool
	backupDir  string
)

func init() {
	backupCmd.Flags().BoolVar(&backupHTML, "html", false, "download HTML instead of markdown")
	backupCmd.Flags().StringVarP(&backupDir, "dir", "d", ".", "directory to write to")
	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Download all notes",
	Long: `
freenote backup will download all notes on the Freenote server and save them to
local disk. Notes will be downloaded as markdown by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		var err error
		path := filepath.Clean(backupDir)
		ext := ".md"
		if backupHTML {
			ext = ".html"
		}

		var payload rest.DecoratedNotes

		c, err := initClient()
		if err != nil {
			fmt.Println("failed to connect: ", err)
			os.Exit(1)
		}
		ownerID := c.User.ID

		url := fmt.Sprintf("/users/%s/notes/", ownerID)
		for {
			err = c.Get(url, &payload)
			if err != nil {
				fmt.Println("get note list failed: ", err)
				os.Exit(1)
			}
			for _, note := range payload.Notes {
				wg.Add(1)
				go func(n rest.DecoratedNote) {
					downloadFile := filepath.Join(path, n.ID.String()+ext)
					f, err := os.OpenFile(downloadFile, os.O_CREATE, 0600)
					if err != nil {
						fmt.Println("failed to open ", downloadFile, ": ", err)
						os.Exit(1)
					}
					defer f.Close()
					note := new(notes.Note)
					err = c.Get(n.Links["canonical"].Href, note)
					if err != nil {
						fmt.Println("get note failed: ", err)
						os.Exit(1)
					}
					var body string
					if backupHTML {
						body = note.HTMLBody
					} else {
						body = note.Body
					}
					_, err = fmt.Fprint(f, body)
					if err != nil {
						fmt.Println("failed to write file: ", err)
						os.Exit(1)
					}
					f.Close()
				}(note)
			}

			if next, ok := payload.Links["next"]; !ok {
				break
			} else {
				url = next.Href
			}
		}
	},
}
