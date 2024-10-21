/*
Copyright © 2021 Upvest GmbH <support@upvest.co>

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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

// Filled in at link-time by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the application version",
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func printVersion() {
	_, mod := filepath.Split(os.Args[0])
	cyan := color.FgCyan.Render
	info := []string{
		fmt.Sprintf(mod+" %s", cyan(version)),
		fmt.Sprintf("built with %s from commit %s at %s by %s", cyan(runtime.Version()), cyan(commit), cyan(date), cyan(builtBy)),
	}
	output := strings.Join(info, "\n")
	fmt.Println(output)
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
