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

// cellsCmd represents the cells command
var cellsCmd = &cobra.Command{
	Use:   "cells",
	Short: "Manage cells in a spreadsheet",
}

// cellsListCmd represents the cells list command
var cellsListCmd = &cobra.Command{
	Use:   "list <spreadsheet-id>",
	Short: "Search for cells by content in a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runCellsList,
}

// cellsGetCmd represents the cells get command
var cellsGetCmd = &cobra.Command{
	Use:   "get <spreadsheet-id>",
	Short: "Get cell values from a worksheet (entire sheet if --range is omitted)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCellsGet,
}

// cellsUpdateCmd represents the cells update command
var cellsUpdateCmd = &cobra.Command{
	Use:   "update <spreadsheet-id>",
	Short: "Update cell values in a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runCellsUpdate,
}

// cellsClearCmd represents the cells clear command
var cellsClearCmd = &cobra.Command{
	Use:   "clear <spreadsheet-id>",
	Short: "Clear cell values in a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runCellsClear,
}

// cellsImportCmd represents the cells import command
var cellsImportCmd = &cobra.Command{
	Use:   "import <spreadsheet-id>",
	Short: "Import CSV data into a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runCellsImport,
}

func runCellsList(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	query, _ := cmd.Flags().GetString("query")
	useRegex, _ := cmd.Flags().GetBool("regex")
	format, _ := cmd.Flags().GetString("format")

	if sheet == "" {
		return fmt.Errorf("--sheet is required")
	}
	if query == "" {
		return fmt.Errorf("--query is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	locations, err := gsheet.FindCells(svc.Sheets.Service, spreadsheetID, sheet, query, useRegex)
	if err != nil {
		return err
	}

	return gsheet.FormatCellLocations(os.Stdout, locations, gsheet.OutputFormat(format))
}

func runCellsGet(cmd *cobra.Command, args []string) error {
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

func runCellsUpdate(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	cellRange, _ := cmd.Flags().GetString("range")
	valuesStr, _ := cmd.Flags().GetString("values")
	filePath, _ := cmd.Flags().GetString("file")
	useStdin, _ := cmd.Flags().GetBool("stdin")

	if cellRange == "" {
		return fmt.Errorf("--range is required")
	}

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

	rangeStr := gsheet.BuildRange(sheet, cellRange)

	if err := gsheet.UpdateValues(svc.Sheets.Service, spreadsheetID, rangeStr, values); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Values updated successfully.")
	return nil
}

func runCellsClear(cmd *cobra.Command, args []string) error {
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

func runCellsImport(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	cellRange, _ := cmd.Flags().GetString("range")
	filePath, _ := cmd.Flags().GetString("file")
	useStdin, _ := cmd.Flags().GetBool("stdin")

	if sheet == "" {
		return fmt.Errorf("--sheet is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	values, err := parseInputValues(filePath, useStdin, "")
	if err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	rangeStr := gsheet.BuildRange(sheet, cellRange)

	if err := gsheet.UpdateValues(svc.Sheets.Service, spreadsheetID, rangeStr, values); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Values imported successfully.")
	return nil
}

// parseInputValues reads values from --file, --stdin, or --values flag.
func parseInputValues(filePath string, useStdin bool, valuesStr string) ([][]interface{}, error) {
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("unable to open file: %w", err)
		}
		defer f.Close()
		return gsheet.ParseValuesFromReader(f)
	}
	if useStdin {
		return gsheet.ParseValuesFromReader(os.Stdin)
	}
	if valuesStr != "" {
		return gsheet.ParseValues(valuesStr), nil
	}
	return nil, fmt.Errorf("either --file, --stdin, or --values is required")
}

func init() {
	rootCmd.AddCommand(cellsCmd)

	cellsCmd.AddCommand(cellsListCmd)
	cellsListCmd.Flags().String("sheet", "", "Sheet name to search in (required)")
	cellsListCmd.Flags().StringP("query", "q", "", "Search text (substring match)")
	cellsListCmd.Flags().Bool("regex", false, "Treat query as regular expression")
	cellsListCmd.Flags().String("format", "text", "Output format (text or json)")

	cellsCmd.AddCommand(cellsGetCmd)
	cellsGetCmd.Flags().String("sheet", "", "Sheet name (default: Sheet1)")
	cellsGetCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C10). Omit to fetch the entire sheet.")
	cellsGetCmd.Flags().String("format", "text", "Output format (text, json, or csv)")

	cellsCmd.AddCommand(cellsUpdateCmd)
	cellsUpdateCmd.Flags().String("sheet", "", "Sheet name")
	cellsUpdateCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C2)")
	cellsUpdateCmd.Flags().String("values", "", `Values to set (semicolon-separated rows, comma-separated cells, e.g., "a,b,c;d,e,f")`)
	cellsUpdateCmd.Flags().String("file", "", "Path to CSV file")
	cellsUpdateCmd.Flags().Bool("stdin", false, "Read CSV values from stdin")

	cellsCmd.AddCommand(cellsClearCmd)
	cellsClearCmd.Flags().String("sheet", "", "Sheet name")
	cellsClearCmd.Flags().String("range", "", "Cell range in A1 notation (e.g., A1:C10)")

	cellsCmd.AddCommand(cellsImportCmd)
	cellsImportCmd.Flags().String("sheet", "", "Sheet name (required)")
	cellsImportCmd.Flags().String("range", "A1", "Starting cell in A1 notation")
	cellsImportCmd.Flags().String("file", "", "Path to CSV file")
	cellsImportCmd.Flags().Bool("stdin", false, "Read CSV from stdin")
}
