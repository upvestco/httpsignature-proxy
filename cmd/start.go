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
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/runtime"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/tunnels"
	"github.com/upvestco/httpsignature-proxy/service/ui"

	"github.com/upvestco/httpsignature-proxy/config"
)

const (
	privateKeyFileNameFlag = "private-key"
	privateKeyPasswordFlag = "private-key-password"
	keyIDFlag              = "key-id"
	clientIDFlag           = "client-id"
	serverBaseUrlFlag      = "server-base-url"
	portFlag               = "port"
	verboseModeFlag        = "verbose-mode"
	updateFlag             = "update"
	listenFlag             = "listen"
	eventsFlag             = "events"
	showWebhookHeader      = "show-webhook-headers"
	uiFlag                 = "ui"
)

var (
	keyConfigs         []config.KeyConfig
	privateKeyFileName string
	privateKeyPassword string
	serverBaseUrl      string
	keyID              string
	clientID           string
	port               int
	verboseMode        bool
	update             bool
	listen             bool
	logHeaders         bool
	uiIsActive         bool
	events             []string
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
	startCmd.Flags().StringVarP(&privateKeyPassword, privateKeyPasswordFlag, "P", "", "password of the private key")
	startCmd.Flags().StringVarP(&serverBaseUrl, serverBaseUrlFlag, "s", "", "server base URL to pipe the requests to")
	startCmd.Flags().StringVarP(&keyID, keyIDFlag, "i", "", "id of the private key")
	startCmd.Flags().StringVarP(&clientID, clientIDFlag, "c", "", "client id for the private key")
	startCmd.Flags().BoolVarP(&verboseMode, verboseModeFlag, "v", false, "enable verbose mode")
	startCmd.Flags().IntVarP(&port, portFlag, "p", 3000, "port to start server")
	startCmd.Flags().BoolVarP(&listen, listenFlag, "l", false, "enable webhook events listening")
	startCmd.Flags().StringSliceVarP(&events, eventsFlag, "e", []string{}, "subscribe for event types")
	startCmd.Flags().BoolVar(&logHeaders, showWebhookHeader, false, "show webhook request headers.")
	startCmd.Flags().BoolVar(&uiIsActive, uiFlag, false, "enable UI mode")
}

func startProxy() {
	cfg, signerConfigs := initializeSignerConfig()
	if !listen && uiIsActive {
		fmt.Printf("Warning: UI mode could be enabled with the webhook events listening only. Ignore --ui flag")
	}
	if uiIsActive && listen {
		startWithUI(cfg, signerConfigs)
	} else {
		startDefault(cfg, signerConfigs)
	}
}

func startWithUI(cfg *config.Config, signerConfigs map[string]runtime.SignerConfig) {
	var tnls *tunnels.Tunnels
	var userCredentialsCh chan tunnels.UserCredentials
	wg := sync.WaitGroup{}
	wg.Add(1)
	ll := ui.CreateLogger(cfg.VerboseMode)

	ui.Create(func() {
		ui.Close()
		wg.Done()
	})
	if listen {
		userCredentialsCh = make(chan tunnels.UserCredentials)
		proxyAddress := fmt.Sprintf("http://localhost:%d", port)
		tnls = tunnels.CreateTunnels(ll, events, proxyAddress,
			func(credentials tunnels.UserCredentials) tunnels.ApiClient {
				return tunnels.NewClient(proxyAddress, credentials, cfg.DefaultTimeout)
			}, logHeaders)
	}

	proxy := runtime.NewProxy(cfg, signerConfigs, userCredentialsCh, ll)
	if err := proxy.Run(); err == nil {
		ll.PrintF("Starting to listen on port %d\n", port)
	} else {
		panic("Fail to start http proxy: " + err.Error())
	}
	if tnls != nil {
		go tnls.Start(userCredentialsCh)
	}
	wg.Wait()
	if tnls != nil {
		close(userCredentialsCh)
		tnls.Stop()
	}
}

func startDefault(cfg *config.Config, signerConfigs map[string]runtime.SignerConfig) {
	ll := logger.New(cfg.VerboseMode)
	var userCredentialsCh chan tunnels.UserCredentials
	var tnls *tunnels.Tunnels
	if listen {
		userCredentialsCh = make(chan tunnels.UserCredentials)
		proxyAddress := fmt.Sprintf("http://localhost:%d", port)
		tnls = tunnels.CreateTunnels(ll, events, proxyAddress,
			func(credentials tunnels.UserCredentials) tunnels.ApiClient {
				return tunnels.NewClient(proxyAddress, credentials, cfg.DefaultTimeout)
			}, logHeaders)
	}

	proxy := runtime.NewProxy(cfg, signerConfigs, userCredentialsCh, ll)
	if err := proxy.Run(); err == nil {
		ll.PrintF("Starting to listen on port %d\n", port)
	} else {
		panic("Fail to start http proxy: " + err.Error())
	}
	if tnls != nil {
		go tnls.Start(userCredentialsCh)
	}
	ll.PrintLn("Press CTRL-C to exit")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
	if tnls != nil {
		close(userCredentialsCh)
		tnls.Stop()
	}
}

//// startProxy starts the listener
//func startProxy() {
//	fmt.Printf("Starting to listen on port %d\n", port)
//
//	flagConfig := config.KeyConfig{
//		ClientID: clientID,
//		BaseConfig: config.BaseConfig{
//			BaseUrl:            serverBaseUrl,
//			KeyID:              keyID,
//			PrivateKeyFileName: privateKeyFileName,
//			Password:           privateKeyPassword,
//		},
//	}
//
//	if !flagConfig.IsEmpty() {
//		if err := flagConfig.BaseConfig.Validate(); err != nil {
//			fatalConfigError(flagConfig, err)
//		}
//		if flagConfig.ClientID == "" {
//			flagConfig.ClientID = config.DefaultClientKey
//		}
//		keyConfigs = append(keyConfigs, flagConfig)
//	}
//
//	cfg := &config.Config{
//		Port:           port,
//		DefaultTimeout: 30 * time.Second,
//		VerboseMode:    verboseMode,
//		KeyConfigs:     keyConfigs,
//	}
//
//	signerConfigs := make(map[string]runtime.SignerConfig)
//	for i := range cfg.KeyConfigs {
//		if err := cfg.KeyConfigs[i].Validate(); err != nil {
//			fatalConfigError(cfg.KeyConfigs[i], err)
//		}
//		builder, err := signer.NewLocalPrivateSchemeBuilder(&cfg.KeyConfigs[i].BaseConfig)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		clientID := cfg.KeyConfigs[i].ClientID
//		if _, ok := signerConfigs[clientID]; ok {
//			fmt.Printf("ClientID duplicated in configuration\n")
//			log.Fatalf("Stopped due missconfiguration")
//		}
//
//		signerConfigs[clientID] = runtime.SignerConfig{
//			SignBuilder: builder,
//			KeyConfig:   cfg.KeyConfigs[i].BaseConfig,
//		}
//	}
//
//	fmt.Printf("Private keys initialised: \n")
//	for i := range keyConfigs {
//		fmt.Printf("  Key %d for clientID %s:\n", i+1, keyConfigs[i].ClientID)
//		fmt.Printf("  - Using private key file %s for HTTP Signatures\n", keyConfigs[i].PrivateKeyFileName)
//		fmt.Printf("  - Using keyID %s for HTTP Signatures\n", keyConfigs[i].KeyID)
//		fmt.Printf("  - Piping all requests to %s\n", keyConfigs[i].BaseUrl)
//	}
//
//}
//func fatalConfigError(keyConfig config.KeyConfig, err error) {
//    fmt.Printf("Invalid confiruration:\n - keyID: %s;\n - clientID: %s;\n - privateKey: %s;\n - baseUrl: %s\n",
//        keyConfig.KeyID, keyConfig.ClientID, keyConfig.PrivateKeyFileName, keyConfig.BaseUrl)
//    fmt.Printf("Error: %s\n", err.Error())
//    log.Fatalf("invalid configuration: %s\n", err.Error())
//}

func initializeSignerConfig() (*config.Config, map[string]runtime.SignerConfig) {
	flagConfig := config.KeyConfig{
		ClientID: clientID,
		BaseConfig: config.BaseConfig{
			BaseUrl:            serverBaseUrl,
			KeyID:              keyID,
			PrivateKeyFileName: privateKeyFileName,
			Password:           privateKeyPassword,
		},
	}

	if !flagConfig.IsEmpty() {
		if err := flagConfig.BaseConfig.Validate(); err != nil {
			fatalConfigError(flagConfig, err)
		}
		if flagConfig.ClientID == "" {
			flagConfig.ClientID = config.DefaultClientKey
		}
		keyConfigs = append(keyConfigs, flagConfig)
	}

	cfg := &config.Config{
		Port:           port,
		DefaultTimeout: 30 * time.Second,
		PullDelay:      time.Second,
		VerboseMode:    verboseMode,
		KeyConfigs:     keyConfigs,
		LogHeaders:     logHeaders,
	}

	signerConfigs := make(map[string]runtime.SignerConfig)
	for i := range cfg.KeyConfigs {
		if err := cfg.KeyConfigs[i].Validate(); err != nil {
			fatalConfigError(cfg.KeyConfigs[i], err)
		}
		builder, err := signer.NewLocalPrivateSchemeBuilder(&cfg.KeyConfigs[i].BaseConfig)
		if err != nil {
			log.Fatal(err)
		}

		clientID := cfg.KeyConfigs[i].ClientID
		if _, ok := signerConfigs[clientID]; ok {
			fmt.Printf("ClientID duplicated in configuration\n")
			log.Fatalf("Stopped due missconfiguration")
		}

		signerConfigs[clientID] = runtime.SignerConfig{
			SignBuilder: builder,
			KeyConfig:   cfg.KeyConfigs[i].BaseConfig,
		}
	}

	fmt.Println("Private keys initialised:")
	for i := range keyConfigs {
		fmt.Printf("  Key %d for clientID %s:\n", i+1, keyConfigs[i].ClientID)
		fmt.Printf("  - Using private key file %s for HTTP Signatures\n", keyConfigs[i].PrivateKeyFileName)
		fmt.Printf("  - Using keyID %s for HTTP Signatures\n", keyConfigs[i].KeyID)
		fmt.Printf("  - Piping all requests to %s\n", keyConfigs[i].BaseUrl)
	}

	return cfg, signerConfigs
}

func fatalConfigError(keyConfig config.KeyConfig, err error) {
	fmt.Printf("Invalid confiruration:\n - keyID: %s;\n - clientID: %s;\n - privateKey: %s;\n - baseUrl: %s\n",
		keyConfig.KeyID, keyConfig.ClientID, keyConfig.PrivateKeyFileName, keyConfig.BaseUrl)
	fmt.Printf("Error: %s\n", err.Error())
	log.Fatalf("invalid configuration: %s\n", err.Error())
}
