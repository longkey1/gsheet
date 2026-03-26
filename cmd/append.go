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
	"fmt"
	"os"

	"github.com/longkey1/gsheet/internal/gsheet"
	"github.com/spf13/cobra"
)

// appendCmd represents the append command
var appendCmd = &cobra.Command{
	Use:   "append <spreadsheet-id>",
	Short: "Append rows to a sheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runAppend,
}

func runAppend(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	valuesStr, _ := cmd.Flags().GetString("values")
	filePath, _ := cmd.Flags().GetString("file")
	useStdin, _ := cmd.Flags().GetBool("stdin")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	values, err := parseInputValues(filePath, useStdin, valuesStr)
	if err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	rangeStr := sheet
	if rangeStr == "" {
		rangeStr = "Sheet1"
	}

	if err := gsheet.AppendValues(svc.Sheets.Service, spreadsheetID, rangeStr, values); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Rows appended successfully.")
	return nil
}

func init() {
	rootCmd.AddCommand(appendCmd)
	appendCmd.Flags().String("sheet", "", "Sheet name (default: Sheet1)")
	appendCmd.Flags().String("values", "", `Values to append (semicolon-separated rows, comma-separated cells, e.g., "a,b,c;d,e,f")`)
	appendCmd.Flags().String("file", "", "Path to CSV file")
	appendCmd.Flags().Bool("stdin", false, "Read CSV values from stdin")
}
