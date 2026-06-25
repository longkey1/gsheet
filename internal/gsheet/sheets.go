package gsheet

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
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

// SheetInfo represents a worksheet tab within a spreadsheet
type SheetInfo struct {
	SheetID  int64  `json:"sheetId"`
	Title    string `json:"title"`
	Index    int64  `json:"index"`
	RowCount int64  `json:"rowCount"`
	ColCount int64  `json:"colCount"`
}

// CellLocation represents a found cell's position
type CellLocation struct {
	Sheet string `json:"sheet"`
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Cell  string `json:"cell"`
	Value string `json:"value"`
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

// GetSheets retrieves all worksheet tabs from a spreadsheet
func GetSheets(svc *sheets.Service, spreadsheetID string) ([]SheetInfo, error) {
	resp, err := svc.Spreadsheets.Get(spreadsheetID).
		Fields("sheets.properties").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get spreadsheet: %v", err)
	}

	var result []SheetInfo
	for _, s := range resp.Sheets {
		p := s.Properties
		result = append(result, SheetInfo{
			SheetID:  p.SheetId,
			Title:    p.Title,
			Index:    p.Index,
			RowCount: p.GridProperties.RowCount,
			ColCount: p.GridProperties.ColumnCount,
		})
	}
	return result, nil
}

// FilterSheets filters sheets by substring or regex match on title
func FilterSheets(allSheets []SheetInfo, query string, useRegex bool) ([]SheetInfo, error) {
	if query == "" {
		return allSheets, nil
	}

	if useRegex {
		re, err := regexp.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
		var result []SheetInfo
		for _, s := range allSheets {
			if re.MatchString(s.Title) {
				result = append(result, s)
			}
		}
		return result, nil
	}

	var result []SheetInfo
	for _, s := range allSheets {
		if strings.Contains(s.Title, query) {
			result = append(result, s)
		}
	}
	return result, nil
}

// FindCells searches for cells containing the query text in a worksheet
func FindCells(svc *sheets.Service, spreadsheetID, sheet, query string, useRegex bool) ([]CellLocation, error) {
	data, err := GetValues(svc, spreadsheetID, sheet)
	if err != nil {
		return nil, err
	}

	var re *regexp.Regexp
	if useRegex {
		re, err = regexp.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	var results []CellLocation
	for rowIdx, row := range data.Values {
		for colIdx, cell := range row {
			matched := false
			if useRegex {
				matched = re.MatchString(cell)
			} else {
				matched = strings.Contains(cell, query)
			}
			if matched {
				results = append(results, CellLocation{
					Sheet: sheet,
					Row:   rowIdx,
					Col:   colIdx,
					Cell:  colRowToA1(colIdx, rowIdx),
					Value: cell,
				})
			}
		}
	}
	return results, nil
}

// AddSheet adds a new empty worksheet tab to a spreadsheet.
func AddSheet(svc *sheets.Service, spreadsheetID, title string) (*SheetInfo, error) {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: title,
					},
				},
			},
		},
	}
	resp, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to add sheet: %v", err)
	}
	p := resp.Replies[0].AddSheet.Properties
	return &SheetInfo{
		SheetID:  p.SheetId,
		Title:    p.Title,
		Index:    p.Index,
		RowCount: p.GridProperties.RowCount,
		ColCount: p.GridProperties.ColumnCount,
	}, nil
}

// DuplicateSheet copies an existing worksheet tab within a spreadsheet.
func DuplicateSheet(svc *sheets.Service, spreadsheetID string, sourceSheetID int64, newTitle string) (*SheetInfo, error) {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DuplicateSheet: &sheets.DuplicateSheetRequest{
					SourceSheetId: sourceSheetID,
					NewSheetName:  newTitle,
				},
			},
		},
	}
	resp, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to duplicate sheet: %v", err)
	}
	p := resp.Replies[0].DuplicateSheet.Properties
	return &SheetInfo{
		SheetID:  p.SheetId,
		Title:    p.Title,
		Index:    p.Index,
		RowCount: p.GridProperties.RowCount,
		ColCount: p.GridProperties.ColumnCount,
	}, nil
}

// GetSheetID finds the numeric sheetId for a given sheet title.
func GetSheetID(svc *sheets.Service, spreadsheetID, sheetTitle string) (int64, error) {
	allSheets, err := GetSheets(svc, spreadsheetID)
	if err != nil {
		return 0, err
	}
	for _, s := range allSheets {
		if s.Title == sheetTitle {
			return s.SheetID, nil
		}
	}
	return 0, fmt.Errorf("sheet not found: %s", sheetTitle)
}

// InsertDimension inserts rows or columns into a sheet.
// dimension is "ROWS" or "COLUMNS". startIndex and endIndex are 0-based (endIndex is exclusive).
func InsertDimension(svc *sheets.Service, spreadsheetID string, sheetID int64, dimension string, startIndex, endIndex int64, inheritFromBefore bool) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				InsertDimension: &sheets.InsertDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  dimension,
						StartIndex: startIndex,
						EndIndex:   endIndex,
					},
					InheritFromBefore: inheritFromBefore,
				},
			},
		},
	}
	_, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to insert %s: %v", strings.ToLower(dimension), err)
	}
	return nil
}

// DeleteDimension deletes rows or columns from a sheet.
// dimension is "ROWS" or "COLUMNS". startIndex and endIndex are 0-based (endIndex is exclusive).
func DeleteDimension(svc *sheets.Service, spreadsheetID string, sheetID int64, dimension string, startIndex, endIndex int64) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  dimension,
						StartIndex: startIndex,
						EndIndex:   endIndex,
					},
				},
			},
		},
	}
	_, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to delete %s: %v", strings.ToLower(dimension), err)
	}
	return nil
}

// UpdateNote sets or clears the note on a single cell identified by A1 notation.
// Pass an empty string to clear the note.
func UpdateNote(svc *sheets.Service, spreadsheetID string, sheetID int64, cell, note string) error {
	col, row, err := a1ToColRow(cell)
	if err != nil {
		return err
	}
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateCells: &sheets.UpdateCellsRequest{
					Rows: []*sheets.RowData{
						{Values: []*sheets.CellData{{Note: note}}},
					},
					Fields: "note",
					Range: &sheets.GridRange{
						SheetId:          sheetID,
						StartRowIndex:    int64(row),
						EndRowIndex:      int64(row + 1),
						StartColumnIndex: int64(col),
						EndColumnIndex:   int64(col + 1),
					},
				},
			},
		},
	}
	_, err = svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to update note: %v", err)
	}
	return nil
}

// a1ToColRow converts an A1 notation cell reference to 0-based column and row indices.
func a1ToColRow(cell string) (col, row int, err error) {
	cell = strings.ToUpper(strings.TrimSpace(cell))
	split := strings.IndexAny(cell, "0123456789")
	if split <= 0 || split == len(cell) {
		return 0, 0, fmt.Errorf("invalid cell reference: %s", cell)
	}
	colPart := cell[:split]
	rowPart := cell[split:]

	for _, c := range colPart {
		if c < 'A' || c > 'Z' {
			return 0, 0, fmt.Errorf("invalid cell reference: %s", cell)
		}
		col = col*26 + int(c-'A'+1)
	}
	col--

	rowNum, err := strconv.Atoi(rowPart)
	if err != nil || rowNum < 1 {
		return 0, 0, fmt.Errorf("invalid cell reference: %s", cell)
	}
	row = rowNum - 1
	return col, row, nil
}

// colRowToA1 converts 0-based column and row to A1 notation
func colRowToA1(col, row int) string {
	colStr := ""
	c := col
	for {
		colStr = string(rune('A'+c%26)) + colStr
		c = c/26 - 1
		if c < 0 {
			break
		}
	}
	return fmt.Sprintf("%s%d", colStr, row+1)
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
