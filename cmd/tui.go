package cmd

import (
	"io"
	"log"
	"os"

	"codeberg.org/splitringresonator/multiband/internal/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	GroupID: "tools",
	Short:   "Surf the waves in style",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var opts []tea.ProgramOption

		if !isatty.IsTerminal(os.Stdout.Fd()) {
			opts = []tea.ProgramOption{tea.WithoutRenderer()}
		} else {
			log.SetOutput(io.Discard)
		}

		p := tea.NewProgram(tui.NewModel(), opts...)
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
}
