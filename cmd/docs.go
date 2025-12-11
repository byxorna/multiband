package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	docs_cli "codeberg.org/splitringresonator/multiband/internal/cli/docs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var docsServeCmd = &cobra.Command{
	Use:     "serve",
	GroupID: "docs",
	Short:   "Serve embedded docs over HTTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := "127.0.0.1"
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Documentation served at http://%s:%d/\nctrl-c to exit\n", host, port)

		httpfs := http.FileServer(http.FS(docs.Docs)) // TODO: render markdown to html
		return http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), httpfs)
	},
}

var docsCmd = &cobra.Command{
	Use:     "docs",
	GroupID: "docs",
	Short:   "View built-in documentation",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		items := []string{}

		if err := fs.WalkDir(docs.Docs, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				name := d.Name()

				if name == "." || name == ".." {
					return nil
				}

				return nil // keep walking
			}

			if !d.IsDir() {
				if !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}

				items = append(items, path)
			}
			return nil
		}); err != nil {
			return err
		}

		var width, height uint
		isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
		if isTerminal {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				width = uint(w)  //nolint:gosec
				height = uint(h) //nolint:gosec
			}

			//if width > 120 {
			//	width = 120
			//}
		}
		//if width == 0 {
		//	width = 80
		//}

		height = 0 // TODO: debug why using reported term height makes pager layout hard to manage with header
		p := tea.NewProgram(docs_cli.NewModel(width, height))

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Local flag: only applies to `serve`.
	docsCmd.AddGroup(&cobra.Group{
		ID:    "docs",
		Title: "Documentation",
	})
	docsServeCmd.Flags().Int("port", 8080, "port to listen on")
	docsCmd.AddCommand(docsServeCmd)
}
