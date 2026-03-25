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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <spreadsheet-id>",
	Short: "Get cell values from a sheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	cellRange, _ := cmd.Flags().GetString("range")
	format, _ := cmd.Flags().GetString("format")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	rangeStr := gsheet.BuildRange(sheet, cellRange)
	if rangeStr == "" {
		rangeStr = "Sheet1"
	}

	data, err := gsheet.GetValues(svc.Sheets.Service, spreadsheetID, rangeStr)
	if err != nil {
		return err
	}

	return gsheet.FormatCellData(os.Stdout, data, gsheet.OutputFormat(format))
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().String("sheet", "", "Sheet name (default: Sheet1)")
	getCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C10)")
	getCmd.Flags().String("format", "text", "Output format (text or json)")
}
