package cmd

import (
	"io/fs"
	"os"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	docs_cli "codeberg.org/splitringresonator/multiband/internal/cli/docs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "View built-in documentation",
	Args:  cobra.NoArgs,
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
