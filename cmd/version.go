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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"codeberg.org/splitringresonator/multiband/internal/version"
	"github.com/spf13/cobra"
)

var (
	noteworthyDependencies = []string{
		"github.com/Sudo-Ivan/reticulum-go",
		"github.com/gomarkdown/markdown",
		"github.com/cockroachdb/pebble",
	}
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		bundle := struct {
			Title                  string `json:"title"`
			Program                string `json:"program"`
			BuiltRFC3339           string `json:"built_rfc3339"`
			Commit                 string `json:"commit"`
			*debug.BuildInfo       `json:"build_info"`
			Architecture           string   `json:"architecture"`
			Runtime                string   `json:"runtime"`
			NoteworthyDependencies []string `json:"noteworthy_dependencies"`
		}{
			Title:        rootCmd.Short,
			Program:      os.Args[0],
			BuiltRFC3339: version.BuiltAt().Format(time.RFC3339),
			Commit:       version.Commit,
			Architecture: runtime.GOARCH,
			Runtime:      runtime.Version(),
		}

		if bi, ok := debug.ReadBuildInfo(); ok {
			bundle.BuildInfo = bi
			deps := []string{}
			for _, d := range bi.Deps {
				var ofNote bool
				for _, dep := range noteworthyDependencies {
					if strings.HasPrefix(d.Path, dep) {
						ofNote = true
						break
					}
				}
				if ofNote {
					var x string
					x = fmt.Sprintf("%s@%s", d.Path, d.Version)
					if d.Replace != nil {
						x = fmt.Sprintf("%s (via %s@%s)", x, d.Replace.Replace.Path, d.Replace.Version)
					}
					deps = append(deps, x)
				}
			}
			bundle.NoteworthyDependencies = deps
		} else {
			fmt.Fprintf(os.Stderr, "unable to read debug build info\n")
		}

		switch outFormat, _ := cmd.Flags().GetString("output"); outFormat {
		case "json":
			out, err := json.Marshal(bundle)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Print(string(out))
		default:

			fmt.Printf("Program: %s\n", bundle.Program)
			fmt.Printf("Built: %s\nCommit: %s\n", bundle.BuiltRFC3339, bundle.Commit)
			if bundle.BuildInfo != nil {
				bi := bundle.BuildInfo
				fmt.Printf("Package: %s\nVersion: %s\nChecksum: %s\nRuntime: %s\nArchitecture: %s\n", bi.Path, bi.Main.Version, bi.Main.Sum, bi.GoVersion, runtime.GOARCH)
			}

			if len(bundle.NoteworthyDependencies) > 0 {
				fmt.Printf("Dependencies:\n  %s\n", strings.Join(bundle.NoteworthyDependencies, "\n  "))
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
