package cmd

import (
	"codeberg.org/splitringresonator/multiband/pkg/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Surf the waves in style",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(tui.NewModel())
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}
