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
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"

	"codeberg.org/splitringresonator/multiband/internal/version"
	"github.com/spf13/cobra"
)

var (
	noteworthyDependencies = []string{"github.com/Sudo-Ivan/reticulum-go"}
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("(((| Multiband |)))\n")
		fmt.Printf("Compiled: %s\nCommit: %s\n", version.BuiltAt().Format("2006-01-02T15:04:05Z"), version.Commit)

		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("Package: %s\nVersion: %s\nChecksum: %s\nRuntime: %s\nArchitecture: %s\n", bi.Path, bi.Main.Version, bi.Main.Sum, bi.GoVersion, runtime.GOARCH)
			deps := []string{}
			for _, d := range bi.Deps {
				if slices.Contains(noteworthyDependencies, d.Path) {
					var x string
					x = fmt.Sprintf("%s@%s", d.Path, d.Version)
					if d.Replace != nil {
						x = fmt.Sprintf("%s (via %s@%s)", x, d.Replace.Replace.Path, d.Replace.Version)
					}
					deps = append(deps, x)
				}
			}
			if len(deps) > 0 {
				fmt.Printf("Dependencies:\n  %s\n", strings.Join(deps, "\n  "))
			}
		} else {
			fmt.Fprintf(os.Stderr, "unable to read debug build info\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
