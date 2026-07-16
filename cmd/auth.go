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

	"github.com/longkey1/gsheet/internal/google"
	"github.com/longkey1/gsheet/internal/gsheet"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google Sheets API using OAuth",
	Long: `Authenticate with Google Sheets API using OAuth.
This command initiates the OAuth flow to obtain and save access tokens.
Only applicable when auth_type is set to "oauth" in config.`,
	RunE: runAuth,
}

func runAuth(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	if cfg.AuthType != gsheet.AuthTypeOAuth {
		return fmt.Errorf("auth command is only available for OAuth authentication (current: %s)", cfg.AuthType)
	}

	// Check if token already exists
	if _, err := os.Stat(cfg.GoogleUserCredentials); err == nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token file already exists: %s\n", cfg.GoogleUserCredentials)
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "Do you want to re-authenticate? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			return nil
		}
	}

	// Run OAuth flow
	auth := google.NewOAuthAuthenticator(
		cfg.GoogleApplicationCredentials,
		cfg.GoogleUserCredentials,
	)

	if err := auth.Authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Authentication successful!")
	return nil
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.SetOut(os.Stdout)
}
