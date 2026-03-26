package gsheet

import (
	"encoding/csv"
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
	OutputFormatCSV  OutputFormat = "csv"
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

	if format == OutputFormatCSV {
		if len(data.Values) == 0 {
			return nil
		}
		writer := csv.NewWriter(w)
		for _, row := range data.Values {
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("unable to write CSV: %w", err)
			}
		}
		writer.Flush()
		return writer.Error()
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

// FormatSheetList outputs sheet tab list in the specified format
func FormatSheetList(w io.Writer, sheetInfos []SheetInfo, format OutputFormat) error {
	if format == OutputFormatJSON {
		data, err := json.MarshalIndent(sheetInfos, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	if len(sheetInfos) == 0 {
		fmt.Fprintln(w, "No sheets found.")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("SHEET ID", "TITLE", "INDEX", "ROWS", "COLS")

	for _, s := range sheetInfos {
		table.Append(s.SheetID, s.Title, s.Index, s.RowCount, s.ColCount)
	}

	table.Render()
	return nil
}

// FormatCellLocations outputs found cell locations in the specified format
func FormatCellLocations(w io.Writer, locations []CellLocation, format OutputFormat) error {
	if format == OutputFormatJSON {
		data, err := json.MarshalIndent(locations, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	if len(locations) == 0 {
		fmt.Fprintln(w, "No matching cells found.")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("SHEET", "CELL", "VALUE")

	for _, loc := range locations {
		table.Append(loc.Sheet, loc.Cell, loc.Value)
	}

	table.Render()
	return nil
}
