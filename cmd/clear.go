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

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear <spreadsheet-id>",
	Short: "Clear cell values in a sheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runClear,
}

func runClear(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	cellRange, _ := cmd.Flags().GetString("range")

	if cellRange == "" {
		return fmt.Errorf("--range is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	rangeStr := gsheet.BuildRange(sheet, cellRange)

	if err := gsheet.ClearValues(svc.Sheets.Service, spreadsheetID, rangeStr); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Values cleared successfully.")
	return nil
}

func init() {
	rootCmd.AddCommand(clearCmd)
	clearCmd.Flags().String("sheet", "", "Sheet name")
	clearCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C10)")
}
