package gsheet

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// SpreadsheetInfo represents a spreadsheet in Google Drive
type SpreadsheetInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatedTime string `json:"createdTime"`
	ModifiedTime string `json:"modifiedTime"`
	WebViewLink string `json:"webViewLink"`
}

// CellData represents cell values
type CellData struct {
	Range  string     `json:"range"`
	Values [][]string `json:"values"`
}

// ListSpreadsheets returns the list of spreadsheets owned by or shared with the user
func ListSpreadsheets(svc *drive.Service, query string, mine bool, maxResults int64) ([]SpreadsheetInfo, error) {
	q := "mimeType='application/vnd.google-apps.spreadsheet' and trashed=false"
	if mine {
		q += " and 'me' in owners"
	}
	if query != "" {
		q += fmt.Sprintf(" and name contains '%s'", query)
	}

	call := svc.Files.List().
		Q(q).
		Fields("files(id, name, createdTime, modifiedTime, webViewLink)").
		OrderBy("modifiedTime desc")

	if maxResults > 0 {
		call = call.PageSize(maxResults)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list spreadsheets: %v", err)
	}

	var result []SpreadsheetInfo
	for _, f := range resp.Files {
		result = append(result, SpreadsheetInfo{
			ID:           f.Id,
			Name:         f.Name,
			CreatedTime:  f.CreatedTime,
			ModifiedTime: f.ModifiedTime,
			WebViewLink:  f.WebViewLink,
		})
	}

	return result, nil
}

// GetValues retrieves cell values from a spreadsheet
func GetValues(svc *sheets.Service, spreadsheetID, rangeStr string) (*CellData, error) {
	resp, err := svc.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	var values [][]string
	for _, row := range resp.Values {
		var strRow []string
		for _, cell := range row {
			strRow = append(strRow, fmt.Sprintf("%v", cell))
		}
		values = append(values, strRow)
	}

	return &CellData{
		Range:  resp.Range,
		Values: values,
	}, nil
}

// UpdateValues updates cell values in a spreadsheet
func UpdateValues(svc *sheets.Service, spreadsheetID, rangeStr string, values [][]interface{}) error {
	vr := &sheets.ValueRange{
		Values: values,
	}
	_, err := svc.Spreadsheets.Values.Update(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to update data in sheet: %v", err)
	}
	return nil
}

// AppendValues appends rows to a spreadsheet
func AppendValues(svc *sheets.Service, spreadsheetID, rangeStr string, values [][]interface{}) error {
	vr := &sheets.ValueRange{
		Values: values,
	}
	_, err := svc.Spreadsheets.Values.Append(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}
	return nil
}

// ClearValues clears cell values in a spreadsheet
func ClearValues(svc *sheets.Service, spreadsheetID, rangeStr string) error {
	_, err := svc.Spreadsheets.Values.Clear(spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return fmt.Errorf("unable to clear data in sheet: %v", err)
	}
	return nil
}

// BuildRange constructs a range string like "Sheet1!A1:C10"
func BuildRange(sheet, cellRange string) string {
	if sheet == "" && cellRange == "" {
		return ""
	}
	if sheet == "" {
		return cellRange
	}
	if cellRange == "" {
		return sheet
	}
	return fmt.Sprintf("%s!%s", sheet, cellRange)
}

// ParseValues parses a semicolon-separated, comma-separated value string into a 2D array
// Example: "a,b,c;d,e,f" -> [[a,b,c],[d,e,f]]
func ParseValues(input string) [][]interface{} {
	var result [][]interface{}
	rows := strings.Split(input, ";")
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		cells := strings.Split(row, ",")
		var rowData []interface{}
		for _, cell := range cells {
			rowData = append(rowData, strings.TrimSpace(cell))
		}
		result = append(result, rowData)
	}
	return result
}

// ParseValuesFromReader reads CSV-formatted data from a reader
func ParseValuesFromReader(r io.Reader) ([][]interface{}, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	var result [][]interface{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV input: %v", err)
		}
		var row []interface{}
		for _, cell := range record {
			row = append(row, cell)
		}
		result = append(result, row)
	}
	return result, nil
}
