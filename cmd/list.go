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
	"context"
	"os"

	"github.com/longkey1/gsheet/internal/gsheet"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your spreadsheets",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	query, _ := cmd.Flags().GetString("query")
	mine, _ := cmd.Flags().GetBool("mine")
	maxResults, _ := cmd.Flags().GetInt64("max-results")
	format, _ := cmd.Flags().GetString("format")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	spreadsheets, err := gsheet.ListSpreadsheets(svc.Drive.Service, query, mine, maxResults)
	if err != nil {
		return err
	}

	return gsheet.FormatSpreadsheetList(os.Stdout, spreadsheets, gsheet.OutputFormat(format))
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("query", "q", "", "Search spreadsheets by name")
	listCmd.Flags().Bool("mine", false, "Show only spreadsheets owned by me")
	listCmd.Flags().Int64P("max-results", "n", 20, "Maximum number of results")
	listCmd.Flags().String("format", "text", "Output format (text or json)")
}
