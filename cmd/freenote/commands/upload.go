package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aprice/freenote/ids"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/rest"
)

var (
	overWriteID string
	title       string
	folder      string
	uploadFile  string
)

func init() {
	uploadCmd.Flags().StringVar(&overWriteID, "overwrite", "", "ID of note to overwrite")
	uploadCmd.Flags().StringVar(&title, "title", "", "note title (default guess)")
	uploadCmd.Flags().StringVar(&folder, "folder", "", "note folder")
	uploadCmd.Flags().StringVarP(&uploadFile, "file", "f", "", "file to upload (default stdin)")
	rootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a markdown file to Freenote",
	Long: `
freenote upload will upload a markdown note to the Freenote server. Notes will
be uploaded as new notes by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
			f   []byte
		)
		if uploadFile != "" {
			f, err = ioutil.ReadFile(uploadFile)
			if err != nil {
				fmt.Println("unable to read ", uploadFile, ": ", err)
				os.Exit(1)
			}
		} else {
			f, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("unable to read stdin: ", err)
				os.Exit(1)
			}
		}
		if title == "" {
			if bytes.HasPrefix(f, []byte("# ")) {
				if bytes.Contains(f, []byte{'\n'}) {
					title = strings.TrimSpace(string(f[2:bytes.IndexByte(f, '\n')]))
				} else {
					title = string(f[2:])
				}
			} else if uploadFile != "" {
				title = filepath.Base(uploadFile)
				if strings.Contains(title, ".") {
					title = title[1:strings.LastIndex(title, ".")]
				}
			}
		}
		c, err := initClient()
		if err != nil {
			fmt.Println("failed to connect: ", err)
			os.Exit(1)
		}
		ownerID := c.User.ID
		n := notes.Note{
			Title:    title,
			Body:     string(f),
			Created:  time.Now(),
			Modified: time.Now(),
			Owner:    ownerID,
			Folder:   folder,
		}
		resPL := new(struct {
			Links struct {
				Canonical rest.Link `json:"canonical"`
			} `json:"_links"`
		})
		if overWriteID != "" {
			noteID, err := ids.ParseID(overWriteID)
			if err != nil {
				fmt.Printf("could not parse ", overWriteID, " as ID: ", err)
				os.Exit(1)
			}
			n.ID = noteID
			c.Send("PUT",
				fmt.Sprintf("/users/%s/notes/%s", ownerID, noteID),
				&n, resPL)
			if err != nil {
				fmt.Println("note upload failed: ", err)
				os.Exit(1)
			}
		} else {
			c.Send("POST",
				fmt.Sprintf("/users/%s/notes/", ownerID),
				&n, resPL)
			if err != nil {
				fmt.Println("note upload failed: ", err)
				os.Exit(1)
			}
		}
		fmt.Println("Upload successful.")
		fmt.Println("Note URL: ", resPL.Links.Canonical.Href)
	},
}
