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

// sheetsAddCmd represents the sheets add command
var sheetsAddCmd = &cobra.Command{
	Use:   "add <spreadsheet-id>",
	Short: "Add a new worksheet tab to a spreadsheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsAdd,
}

// sheetsRenameCmd represents the sheets rename command
var sheetsRenameCmd = &cobra.Command{
	Use:   "rename <spreadsheet-id>",
	Short: "Rename a worksheet tab in a spreadsheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsRename,
}

// sheetsRowsCmd represents the sheets rows command
var sheetsRowsCmd = &cobra.Command{
	Use:   "rows",
	Short: "Manage rows in a worksheet",
}

// sheetsRowsInsertCmd represents the sheets rows insert command
var sheetsRowsInsertCmd = &cobra.Command{
	Use:   "insert <spreadsheet-id>",
	Short: "Insert rows into a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsRowsInsert,
}

// sheetsRowsDeleteCmd represents the sheets rows delete command
var sheetsRowsDeleteCmd = &cobra.Command{
	Use:   "delete <spreadsheet-id>",
	Short: "Delete rows from a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsRowsDelete,
}

// sheetsColsCmd represents the sheets cols command
var sheetsColsCmd = &cobra.Command{
	Use:   "cols",
	Short: "Manage columns in a worksheet",
}

// sheetsColsInsertCmd represents the sheets cols insert command
var sheetsColsInsertCmd = &cobra.Command{
	Use:   "insert <spreadsheet-id>",
	Short: "Insert columns into a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsColsInsert,
}

// sheetsColsDeleteCmd represents the sheets cols delete command
var sheetsColsDeleteCmd = &cobra.Command{
	Use:   "delete <spreadsheet-id>",
	Short: "Delete columns from a worksheet",
	Args:  cobra.ExactArgs(1),
	RunE:  runSheetsColsDelete,
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

func runSheetsRename(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	title, _ := cmd.Flags().GetString("title")

	if sheet == "" {
		return fmt.Errorf("--sheet is required")
	}
	if title == "" {
		return fmt.Errorf("--title is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	sheetID, err := gsheet.GetSheetID(svc.Sheets.Service, spreadsheetID, sheet)
	if err != nil {
		return err
	}

	if err := gsheet.RenameSheet(svc.Sheets.Service, spreadsheetID, sheetID, title); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Sheet \"%s\" renamed to \"%s\" successfully.\n", sheet, title)
	return nil
}

func runSheetsAdd(cmd *cobra.Command, args []string) error {
	spreadsheetID := args[0]
	title, _ := cmd.Flags().GetString("title")
	from, _ := cmd.Flags().GetString("from")

	if title == "" {
		return fmt.Errorf("--title is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	var info *gsheet.SheetInfo
	if from != "" {
		sourceSheetID, err := gsheet.GetSheetID(svc.Sheets.Service, spreadsheetID, from)
		if err != nil {
			return err
		}
		info, err = gsheet.DuplicateSheet(svc.Sheets.Service, spreadsheetID, sourceSheetID, title)
		if err != nil {
			return err
		}
	} else {
		info, err = gsheet.AddSheet(svc.Sheets.Service, spreadsheetID, title)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stdout, "Sheet \"%s\" added successfully (sheetId: %d).\n", info.Title, info.SheetID)
	return nil
}

func runDimensionInsert(cmd *cobra.Command, args []string, dimension string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	start, _ := cmd.Flags().GetInt("start")
	count, _ := cmd.Flags().GetInt("count")
	after, _ := cmd.Flags().GetBool("after")

	if sheet == "" {
		return fmt.Errorf("--sheet is required")
	}
	if start < 1 {
		return fmt.Errorf("--start must be >= 1")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	sheetID, err := gsheet.GetSheetID(svc.Sheets.Service, spreadsheetID, sheet)
	if err != nil {
		return err
	}

	startIndex := int64(start - 1)
	endIndex := startIndex + int64(count)

	if err := gsheet.InsertDimension(svc.Sheets.Service, spreadsheetID, sheetID, dimension, startIndex, endIndex, after); err != nil {
		return err
	}

	label := "row(s)"
	if dimension == "COLUMNS" {
		label = "column(s)"
	}
	fmt.Fprintf(os.Stdout, "%d %s inserted successfully.\n", count, label)
	return nil
}

func runDimensionDelete(cmd *cobra.Command, args []string, dimension string) error {
	spreadsheetID := args[0]
	sheet, _ := cmd.Flags().GetString("sheet")
	start, _ := cmd.Flags().GetInt("start")
	count, _ := cmd.Flags().GetInt("count")

	if sheet == "" {
		return fmt.Errorf("--sheet is required")
	}
	if start < 1 {
		return fmt.Errorf("--start must be >= 1")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gsheet.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	sheetID, err := gsheet.GetSheetID(svc.Sheets.Service, spreadsheetID, sheet)
	if err != nil {
		return err
	}

	startIndex := int64(start - 1)
	endIndex := startIndex + int64(count)

	if err := gsheet.DeleteDimension(svc.Sheets.Service, spreadsheetID, sheetID, dimension, startIndex, endIndex); err != nil {
		return err
	}

	label := "row(s)"
	if dimension == "COLUMNS" {
		label = "column(s)"
	}
	fmt.Fprintf(os.Stdout, "%d %s deleted successfully.\n", count, label)
	return nil
}

func runSheetsRowsInsert(cmd *cobra.Command, args []string) error {
	return runDimensionInsert(cmd, args, "ROWS")
}

func runSheetsRowsDelete(cmd *cobra.Command, args []string) error {
	return runDimensionDelete(cmd, args, "ROWS")
}

func runSheetsColsInsert(cmd *cobra.Command, args []string) error {
	return runDimensionInsert(cmd, args, "COLUMNS")
}

func runSheetsColsDelete(cmd *cobra.Command, args []string) error {
	return runDimensionDelete(cmd, args, "COLUMNS")
}

func init() {
	rootCmd.AddCommand(sheetsCmd)
	sheetsCmd.AddCommand(sheetsListCmd)
	sheetsCmd.AddCommand(sheetsAddCmd)
	sheetsAddCmd.Flags().String("title", "", "Title for the new worksheet tab (required)")
	sheetsAddCmd.Flags().String("from", "", "Copy from an existing sheet by title")

	sheetsCmd.AddCommand(sheetsRenameCmd)
	sheetsRenameCmd.Flags().String("sheet", "", "Current sheet name (required)")
	sheetsRenameCmd.Flags().String("title", "", "New sheet name (required)")

	sheetsListCmd.Flags().StringP("query", "q", "", "Filter sheets by title (substring match)")
	sheetsListCmd.Flags().Bool("regex", false, "Treat query as regular expression")
	sheetsListCmd.Flags().String("format", "text", "Output format (text or json)")

	sheetsCmd.AddCommand(sheetsRowsCmd)
	sheetsRowsCmd.AddCommand(sheetsRowsInsertCmd)
	sheetsRowsInsertCmd.Flags().String("sheet", "", "Sheet name (required)")
	sheetsRowsInsertCmd.Flags().Int("start", 0, "Row number to insert at (1-based, required)")
	sheetsRowsInsertCmd.Flags().Int("count", 1, "Number of rows to insert")
	sheetsRowsInsertCmd.Flags().Bool("after", false, "Inherit formatting from the row before")

	sheetsRowsCmd.AddCommand(sheetsRowsDeleteCmd)
	sheetsRowsDeleteCmd.Flags().String("sheet", "", "Sheet name (required)")
	sheetsRowsDeleteCmd.Flags().Int("start", 0, "Starting row number (1-based, required)")
	sheetsRowsDeleteCmd.Flags().Int("count", 1, "Number of rows to delete")

	sheetsCmd.AddCommand(sheetsColsCmd)
	sheetsColsCmd.AddCommand(sheetsColsInsertCmd)
	sheetsColsInsertCmd.Flags().String("sheet", "", "Sheet name (required)")
	sheetsColsInsertCmd.Flags().Int("start", 0, "Column number to insert at (1-based, required)")
	sheetsColsInsertCmd.Flags().Int("count", 1, "Number of columns to insert")
	sheetsColsInsertCmd.Flags().Bool("after", false, "Inherit formatting from the column before")

	sheetsColsCmd.AddCommand(sheetsColsDeleteCmd)
	sheetsColsDeleteCmd.Flags().String("sheet", "", "Sheet name (required)")
	sheetsColsDeleteCmd.Flags().Int("start", 0, "Starting column number (1-based, required)")
	sheetsColsDeleteCmd.Flags().Int("count", 1, "Number of columns to delete")
}
