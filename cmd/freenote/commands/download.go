package commands

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/aprice/freenote/notes"
)

var (
	htmlExport   bool
	downloadFile string
)

func init() {
	downloadCmd.Flags().BoolVar(&htmlExport, "html", false, "download HTML instead of markdown")
	downloadCmd.Flags().StringVarP(&downloadFile, "file", "f", "", "file to write to (default stdout)")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download [noteID]",
	Short: "Download a note from Freenote",
	Long: `
freenote download will download a note on the Freenote server and save it to
local disk. Notes will be downloaded as markdown by default.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("You must provide a note ID to download")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		f := os.Stdout
		if downloadFile != "" {
			f, err = os.OpenFile(downloadFile, os.O_CREATE, 0600)
			if err != nil {
				fmt.Println("failed to open ", downloadFile, ": ", err)
				os.Exit(1)
			}
		}
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
		var body string
		if htmlExport {
			body = note.HTMLBody
		} else {
			body = note.Body
		}
		_, err = fmt.Fprint(f, body)
		if err != nil {
			fmt.Println("failed to write file: ", err)
			os.Exit(1)
		}
	},
}
