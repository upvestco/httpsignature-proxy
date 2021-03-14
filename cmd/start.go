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

	"github.com/spf13/cobra"
)

const (
	privateKeyFileNameFlag = "privateKey"
	privateKeyPasswordFlag = "privateKeyPassword"
	serverBaseUrlFlag      = "serverBaseUrl"
)

var (
	privateKeyFileName string
	privateKeyPassword string
	serverBaseUrl      string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the proxy on localhost for signing HTTP-requests",
	Run: func(cmd *cobra.Command, args []string) {
		startProxy(3000)
	},
}

func init() {
	// Setup the CLI arguments for start command
	startCmd.Flags().StringVarP(&privateKeyFileName, privateKeyFileNameFlag, "k", "", "filename of the private key file")
	_ = startCmd.MarkFlagRequired(privateKeyFileNameFlag)

	startCmd.Flags().StringVarP(&privateKeyPassword, privateKeyPasswordFlag, "p", "", "password of the private key")
	_ = startCmd.MarkFlagRequired(privateKeyPasswordFlag)

	startCmd.Flags().StringVarP(&serverBaseUrl, serverBaseUrlFlag, "s", "", "server base URL to pipe the requests to")
	_ = startCmd.MarkFlagRequired(serverBaseUrlFlag)

	// Register the start command
	RootCmd.AddCommand(startCmd)
}

// startProxy starts the listener
func startProxy(port int) {
	fmt.Printf("Starting to listen on port %d\n", port)

	fmt.Printf("- Using private key file %s for HTTP Signatures\n", privateKeyFileName)
	fmt.Printf("- Piping all requests to %s\n", serverBaseUrl)

	// TODO implement the HTTP-server to listen on given port and sign all the requests

}
