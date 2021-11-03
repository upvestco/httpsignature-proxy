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
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/upvestco/httpsignature-proxy/config"
)

const (
	envPrefixName = "HTTP_PROXY"
)

var cfgFile string
var errConfigInitFail = errors.New("failed to initialize config")

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "httpsignature-proxy",
	Short: "HTTP Proxy to add HTTP Signatures to your requests.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	// Global flag for config file
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.httpsignature-proxy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".httpsignature-proxy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".httpsignature-proxy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintf(os.Stderr, "Errors in reading config file: %v", err)
	}
	bindFlags(cmd)

	format := "key-configs.config-%d"
	for i := 1; ; i++ {
		key := fmt.Sprintf(format, i)
		v := viper.Sub(key)
		if v == nil {
			break
		}
		keyConfig, err := mapToConfig(v.AllSettings())
		if err != nil {
			log.Fatalf("failed to initialize key config: %s", key)
		}
		keyConfigs = append(keyConfigs, keyConfig)
	}
}

func mapToConfig(m map[string]interface{}) (config.KeyConfig, error) {
	return config.KeyConfig{
		ClientID: m["client-id"].(string),
		BaseConfig: config.BaseConfig{
			PrivateKeyFileName: m["private-key"].(string),
			Password:           m["private-key-password"].(string),
			BaseUrl:            m["server-base-url"].(string),
			KeyID:              m["key-id"].(string),
		},
	}, nil
}

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if err := viper.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefixName, envVarSuffix)); err != nil {
				log.Fatal(errConfigInitFail)
			}
		}
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				log.Fatal(errConfigInitFail)
			}
		}
	})
}
