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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <spreadsheet-id>",
	Short: "Update cell values in a sheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	cellRange, _ := cmd.Flags().GetString("range")
	valuesStr, _ := cmd.Flags().GetString("values")
	useStdin, _ := cmd.Flags().GetBool("stdin")

	if cellRange == "" {
		return fmt.Errorf("--range is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	var values [][]interface{}
	if useStdin {
		var err error
		values, err = gsheet.ParseValuesFromReader(os.Stdin)
		if err != nil {
			return err
		}
	} else if valuesStr != "" {
		values = gsheet.ParseValues(valuesStr)
	} else {
		return fmt.Errorf("either --values or --stdin is required")
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	rangeStr := gsheet.BuildRange(sheet, cellRange)

	if err := gsheet.UpdateValues(svc.Sheets.Service, spreadsheetID, rangeStr, values); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Values updated successfully.")
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("sheet", "", "Sheet name")
	updateCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C2)")
	updateCmd.Flags().String("values", "", `Values to set (semicolon-separated rows, comma-separated cells, e.g., "a,b,c;d,e,f")`)
	updateCmd.Flags().Bool("stdin", false, "Read CSV values from stdin")
}
