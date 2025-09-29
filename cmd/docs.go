package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glow/v2/ui"
	"github.com/spf13/cobra"
)

var (
	docsFS embed.FS
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "View built-in documentation with Glow",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		localDocsPath := filepath.Join(pwd, "docs")
		fmt.Printf("TODO: this command only views local files at %s currently; FIXME to display embedded docs\n", localDocsPath)

		if _, err := ui.NewProgram(ui.Config{
			Path:         localDocsPath,
			ShowAllFiles: false,
			EnableMouse:  true,
		}, "").Run(); err != nil {
			return err
		}

		return nil
	},
}
