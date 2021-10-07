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
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/runtime"
	"github.com/upvestco/httpsignature-proxy/service/signer"
)

const (
	privateKeyFileNameFlag = "private-key"
	privateKeyPasswordFlag = "private-key-password"
	keyIDFlag              = "key-id"
	serverBaseUrlFlag      = "server-base-url"
	portFlag               = "port"
	verboseModeFlag        = "verbose-mode"
)

var (
	privateKeyFileName string
	privateKeyPassword string
	serverBaseUrl      string
	keyID              string
	port               int
	verboseMode        bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the proxy on localhost for signing HTTP-requests",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initConfig(cmd)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the CLI arguments for start command

		startProxy()
	},
}

func init() {
	// Register the start command
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&privateKeyFileName, privateKeyFileNameFlag, "f", "", "filename of the private key file")
	_ = startCmd.MarkFlagRequired(privateKeyFileNameFlag)

	startCmd.Flags().StringVarP(&privateKeyPassword, privateKeyPasswordFlag, "P", "", "password of the private key")
	_ = startCmd.MarkFlagRequired(privateKeyPasswordFlag)

	startCmd.Flags().StringVarP(&serverBaseUrl, serverBaseUrlFlag, "s", "", "server base URL to pipe the requests to")
	_ = startCmd.MarkFlagRequired(serverBaseUrlFlag)

	startCmd.Flags().StringVarP(&keyID, keyIDFlag, "i", "", "id of the private key")
	_ = startCmd.MarkFlagRequired(keyIDFlag)

	startCmd.Flags().BoolVarP(&verboseMode, verboseModeFlag, "v", false, "enable verbose mode")

	startCmd.Flags().IntVarP(&port, portFlag, "p", 3000, "port to start server")
}

// startProxy starts the listener
func startProxy() {
	fmt.Printf("Starting to listen on port %d\n", port)

	fmt.Printf("- Using private key file %s for HTTP Signatures\n", privateKeyFileName)
	fmt.Printf("- Using keyID %s for HTTP Signatures\n", keyID)

	fmt.Printf("- Piping all requests to %s\n", serverBaseUrl)
	cfg := &config.Config{
		Port:               port,
		BaseUrl:            serverBaseUrl,
		PrivateKeyFileName: privateKeyFileName,
		Password:           privateKeyPassword,
		DefaultTimeout:     30 * time.Second,
		KeyID:              keyID,
		VerboseMode:        verboseMode,
	}

	lsBuilder, err := signer.NewLocalPrivateSchemeBuilder(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := runtime.NewRuntime(cfg, lsBuilder)
	r.Run()
}
