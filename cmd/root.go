/*
Copyright © 2025 Gabe Conradi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"embed"
	"os"

	"codeberg.org/splitringresonator/multiband/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "multiband",
	Short:   "Experimental communications platform",
	Example: ``, //TODO
	Version: version.Verbose(),
}

func Root() *cobra.Command {
	return rootCmd
}

func Execute(docsFS embed.FS) {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(&cobra.Group{
		ID:    "docs",
		Title: "Documentation Commands:",
	}, &cobra.Group{
		ID:    "tools",
		Title: "Tools:",
	})
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format")
	rootCmd.PersistentFlags().BoolP("anon", "A", false, "Generate single use identity for this session")
}
