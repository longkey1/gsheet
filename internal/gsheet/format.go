package gsheet

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
)

// FormatSpreadsheetList outputs spreadsheet list in the specified format
func FormatSpreadsheetList(w io.Writer, spreadsheets []SpreadsheetInfo, format OutputFormat) error {
	if format == OutputFormatJSON {
		data, err := json.MarshalIndent(spreadsheets, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("ID", "NAME", "MODIFIED")

	for _, s := range spreadsheets {
		table.Append(s.ID, s.Name, s.ModifiedTime)
	}

	table.Render()
	return nil
}

// FormatCellData outputs cell data in the specified format
func FormatCellData(w io.Writer, data *CellData, format OutputFormat) error {
	if format == OutputFormatJSON {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(jsonData))
		return nil
	}

	if len(data.Values) == 0 {
		fmt.Fprintln(w, "No data found.")
		return nil
	}

	table := tablewriter.NewWriter(w)

	// Use first row as header
	var headers []any
	for _, cell := range data.Values[0] {
		headers = append(headers, cell)
	}
	table.Header(headers...)

	// Remaining rows as data
	for _, row := range data.Values[1:] {
		var rowData []any
		for _, cell := range row {
			rowData = append(rowData, cell)
		}
		table.Append(rowData)
	}

	table.Render()
	return nil
}
