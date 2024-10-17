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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/pkg/errors"
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
	if update {
		updateModule(mod)
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&update, updateFlag, "u", false, "update proxy to the latest version")
	RootCmd.AddCommand(versionCmd)
}

func updateModule(mod string) {
	env := os.Environ()

	if _, err := runWhich("brew", env); err != nil {
		fmt.Println("Brew is not found. See https://brew.sh/ how to install  brew")
		return
	}

	fmt.Print("Loading updates... ")
	if _, err := runBrew([]string{"update"}, env); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Done")

	fmt.Print("Checking new " + mod + " versions... ")
	newVersion, err := checkNewVersion(mod, env)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Done")

	if len(newVersion) == 0 {
		fmt.Println("No new versions")
		return
	}
	fmt.Print("A new version " + newVersion + " of " + mod + " is available. Updating... ")
	if _, err := runBrew([]string{"upgrade", mod}, env); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Done")
}

func runBrew(params []string, env []string) (string, error) {
	return run("brew", params, env)
}

func runWhich(what string, env []string) (string, error) {
	str, err := run("which", []string{what}, env)
	if err != nil {
		return "", errors.Wrap(err, "RunWithError")
	}
	return strings.Trim(str, " \n"), nil
}

func run(command string, params []string, env []string) (string, error) {
	errBuf := &bytes.Buffer{}
	outBuf := &bytes.Buffer{}
	cmd := exec.Command(command, params...)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return "", errors.New(errBuf.String() + "\n" + err.Error())
	}
	status := strings.TrimSpace(outBuf.String())
	return status, nil
}

func checkNewVersion(c string, env []string) (string, error) {
	status, err := runBrew([]string{"upgrade", c, "-n"}, env)
	if err != nil {
		return "", err
	}
	if len(status) == 0 {
		return "", nil
	}
	lines := strings.Split(status, "\n")
	p := strings.Split(lines[1], " ")
	return strings.Join(p[1:], " "), nil
}
