package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/rest"
)

var (
	overWriteID string
	title       string
	folder      string

	client = new(http.Client)
)

func init() {
	uploadCmd.Flags().StringVar(&overWriteID, "overwrite", "", "ID of note to overwrite")
	uploadCmd.Flags().StringVar(&title, "title", "", "note title (default guess)")
	uploadCmd.Flags().StringVar(&folder, "folder", "", "note folder")
	rootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "Upload a markdown file to Freenote",
	Long: `
freenote upload will upload a markdown file on disk to the Freenote server.
Files will be uploaded as new notes by default.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("You must provide a file name to upload")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		f, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Println("unable to read ", args[0], ": ", err)
		}
		if title == "" {
			title = filepath.Base(args[0])
			if strings.Contains(title, ".") {
				title = title[1:strings.LastIndex(title, ".")]
			}
		}
		//TODO: Get own ID
		ownerID := uuid.Nil
		n := notes.Note{
			Title:    title,
			Body:     string(f),
			Created:  time.Now(),
			Modified: time.Now(),
			Owner:    ownerID,
			Folder:   folder,
		}
		if overWriteID != "" {
			noteID, err := uuid.FromString(overWriteID)
			n.ID = noteID
			pl, err := json.Marshal(&n)
			if err != nil {
				panic(err)
			}
			req, err := http.NewRequest("PUT",
				fmt.Sprintf("https://%s/users/%s/notes/%s", host, ownerID, noteID),
				bytes.NewReader(pl))
			if err != nil {
				panic(err)
			}
			res, err := client.Do(req)
			if err != nil {
				fmt.Println("note upload failed: ", err)
				os.Exit(1)
			} else if res.StatusCode >= 400 {
				fmt.Println("note upload failed: ", res.Status)
				os.Exit(1)
			}
			//TODO: Fixme
			defer res.Body.Close()
			fmt.Println("Upload successful.")
			resPL := new(struct {
				Links struct {
					Canonical rest.Link `json:"canonical"`
				} `json:"_links"`
			})
			err = json.NewDecoder(res.Body).Decode(resPL)
			if err != nil {
				fmt.Println("failed decoding response: ", err)
				os.Exit(1)
			}
			fmt.Println("Note URL: ", resPL.Links.Canonical)
		} else {

		}
	},
}
