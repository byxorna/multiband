package cmd

import (
	"io/fs"
	"os"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	docs_cli "codeberg.org/splitringresonator/multiband/pkg/cli/docs"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
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

		var width uint
		isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
		if isTerminal {
			w, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				width = uint(w) //nolint:gosec
			}

			if width > 120 {
				width = 120
			}
		}
		if width == 0 {
			width = 80
		}

		p := tea.NewProgram(docs_cli.NewModel(width, items))

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}
