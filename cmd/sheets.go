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

// sheetsCmd represents the sheets command
var sheetsCmd = &cobra.Command{
	Use:   "sheets",
	Short: "Manage worksheet tabs in a spreadsheet",
}

// sheetsListCmd represents the sheets list command
var sheetsListCmd = &cobra.Command{
	Use:   "list <spreadsheet-id>",
	Short: "List or search worksheet tabs in a spreadsheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsList,
}

func runSheetsList(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	query, _ := cmd.Flags().GetString("query")
	useRegex, _ := cmd.Flags().GetBool("regex")
	format, _ := cmd.Flags().GetString("format")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	allSheets, err := gsheet.GetSheets(svc.Sheets.Service, spreadsheetID)
	if err != nil {
		return err
	}

	filtered, err := gsheet.FilterSheets(allSheets, query, useRegex)
	if err != nil {
		return err
	}

	return gsheet.FormatSheetList(os.Stdout, filtered, gsheet.OutputFormat(format))
}

func init() {
	rootCmd.AddCommand(sheetsCmd)
	sheetsCmd.AddCommand(sheetsListCmd)
	sheetsListCmd.Flags().StringP("query", "q", "", "Filter sheets by title (substring match)")
	sheetsListCmd.Flags().Bool("regex", false, "Treat query as regular expression")
	sheetsListCmd.Flags().String("format", "text", "Output format (text or json)")
}
