/*
Copyright © 2025 longkey1

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

	"github.com/longkey1/gsheet/internal/gsheet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  *gsheet.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gsheet",
	Short: "Google Sheets cli client",
	// SilenceErrors allows us to handle errors gracefully without cobra printing them twice
	SilenceErrors: true,
	// SilenceUsage prevents usage from being printed on every error
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/gsheet/config.toml)")
	rootCmd.PersistentPreRunE = checkReadOnly
}

// checkReadOnly blocks commands annotated as write operations when read-only
// mode is enabled via the GSHEET_READ_ONLY env var or the read_only config
// option. This lets write commands stay disabled even when no config file
// is present (e.g. when only the env var is set for LLM usage).
func checkReadOnly(cmd *cobra.Command, args []string) error {
	if cmd.Annotations["write"] == "true" && viper.GetBool("read_only") {
		return fmt.Errorf("read-only mode is enabled (GSHEET_READ_ONLY or read_only in config): write commands are disabled")
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".config/gsheet"))
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	cobra.CheckErr(viper.BindEnv("read_only", "GSHEET_READ_ONLY"))

	// Config file is optional for some commands (e.g., version)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only fail if it's not a "file not found" error
			cobra.CheckErr(fmt.Errorf("unable to read config file: %w", err))
		}
		return
	}

	var err error
	config, err = gsheet.LoadConfig()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("unable to load config: %w", err))
	}
}

// GetConfig returns the loaded configuration
// This function will panic if called before config is loaded, but that's intentional
// as commands requiring config should only run after initConfig
func GetConfig() *gsheet.Config {
	if config == nil {
		cobra.CheckErr(fmt.Errorf("config file not found. Please create a config file at $HOME/.config/gsheet/config.toml"))
	}
	return config
}
