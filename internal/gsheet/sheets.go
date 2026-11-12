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
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreatedTime  string `json:"createdTime"`
	ModifiedTime string `json:"modifiedTime"`
	WebViewLink  string `json:"webViewLink"`
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
func UpdateValues(svc *sheets.Service, spreadsheetID, rangeStr string, values [][]any) error {
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

// GetRichValues retrieves cell values with in-cell formatting rendered as Markdown.
// Cell-level text format is included because Sheets stores a run covering the
// whole text as cell-level format instead of as a run.
func GetRichValues(svc *sheets.Service, spreadsheetID, rangeStr string) (*CellData, error) {
	resp, err := svc.Spreadsheets.Get(spreadsheetID).
		Ranges(rangeStr).
		IncludeGridData(true).
		Fields("sheets(properties/title,data(startRow,startColumn,rowData/values(formattedValue,hyperlink,textFormatRuns,userEnteredFormat/textFormat)))").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Sheets) == 0 || len(resp.Sheets[0].Data) == 0 {
		return &CellData{Range: rangeStr}, nil
	}

	grid := resp.Sheets[0].Data[0]
	var values [][]string
	for _, row := range grid.RowData {
		var strRow []string
		for _, cell := range row.Values {
			var cellFormat *sheets.TextFormat
			if cell.UserEnteredFormat != nil {
				cellFormat = cell.UserEnteredFormat.TextFormat
			}
			strRow = append(strRow, RenderMarkdown(cellToSpans(cell.FormattedValue, cell.TextFormatRuns, cellFormat, cell.Hyperlink)))
		}
		// Match Values.Get, which omits trailing empty cells and rows.
		for len(strRow) > 0 && strRow[len(strRow)-1] == "" {
			strRow = strRow[:len(strRow)-1]
		}
		values = append(values, strRow)
	}
	for len(values) > 0 && len(values[len(values)-1]) == 0 {
		values = values[:len(values)-1]
	}

	return &CellData{
		Range:  gridRange(resp.Sheets[0].Properties.Title, int(grid.StartRow), int(grid.StartColumn), values, rangeStr),
		Values: values,
	}, nil
}

// gridRange builds the A1 range covered by values, falling back to the requested range when empty.
func gridRange(title string, startRow, startCol int, values [][]string, requested string) string {
	maxCols := 0
	for _, row := range values {
		maxCols = max(maxCols, len(row))
	}
	if len(values) == 0 || maxCols == 0 {
		return requested
	}
	return fmt.Sprintf("%s!%s:%s", quoteSheetTitle(title),
		colRowToA1(startCol, startRow), colRowToA1(startCol+maxCols-1, startRow+len(values)-1))
}

var simpleSheetTitle = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func quoteSheetTitle(title string) string {
	if simpleSheetTitle.MatchString(title) {
		return title
	}
	return "'" + strings.ReplaceAll(title, "'", "''") + "'"
}

// richCellFormatFields are the cell-level format fields that Markdown fully
// determines in --markdown mode. Other formatting such as color is kept.
const richCellFormatFields = "userEnteredFormat.textFormat.bold," +
	"userEnteredFormat.textFormat.italic," +
	"userEnteredFormat.textFormat.strikethrough," +
	"userEnteredFormat.textFormat.link"

// UpdateRichValues writes values parsed as Markdown, starting at the given cell.
// A cell whose whole text shares one style without a link is written with
// USER_ENTERED and the style is set as cell-level format, so numbers, dates and
// formulas keep being interpreted. Other styled cells are written as strings
// with text format runs, because runs only apply to string values.
func UpdateRichValues(svc *sheets.Service, spreadsheetID string, sheetID int64, rangeStr string, startCol, startRow int, values [][]any) error {
	plain := make([][]any, len(values))
	formatRows := make([]*sheets.RowData, len(values))
	var richRequests []*sheets.Request

	for r, row := range values {
		plain[r] = make([]any, len(row))
		formatRows[r] = &sheets.RowData{}
		for c, v := range row {
			spans := ParseMarkdown(fmt.Sprintf("%v", v))

			// Sheets does not keep a link set as cell-level format, so a linked
			// cell always goes through the run path.
			if len(spans) == 0 || (len(spans) == 1 && spans[0].Style.Link == "") {
				var style richStyle
				if len(spans) == 1 {
					plain[r][c] = spans[0].Text
					style = spans[0].Style
				} else {
					plain[r][c] = ""
				}
				formatRows[r].Values = append(formatRows[r].Values, &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{TextFormat: toTextFormat(style)},
				})
				continue
			}

			formatRows[r].Values = append(formatRows[r].Values, &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{TextFormat: &sheets.TextFormat{}},
			})
			text, runs := spansToCell(spans)
			// nil is skipped by Values.Update; the cell is written by the request below.
			plain[r][c] = nil
			richRequests = append(richRequests, &sheets.Request{
				UpdateCells: &sheets.UpdateCellsRequest{
					Start: &sheets.GridCoordinate{
						SheetId:     sheetID,
						RowIndex:    int64(startRow + r),
						ColumnIndex: int64(startCol + c),
					},
					Rows: []*sheets.RowData{{Values: []*sheets.CellData{{
						UserEnteredValue: &sheets.ExtendedValue{StringValue: &text},
						TextFormatRuns:   runs,
					}}}},
					Fields: "userEnteredValue,textFormatRuns",
				},
			})
		}
	}

	if err := UpdateValues(svc, spreadsheetID, rangeStr, plain); err != nil {
		return err
	}

	// Set cell-level styles and clear runs left from previous values before
	// applying the new runs.
	requests := append([]*sheets.Request{{
		UpdateCells: &sheets.UpdateCellsRequest{
			Start: &sheets.GridCoordinate{
				SheetId:     sheetID,
				RowIndex:    int64(startRow),
				ColumnIndex: int64(startCol),
			},
			Rows:   formatRows,
			Fields: "textFormatRuns," + richCellFormatFields,
		},
	}}, richRequests...)
	_, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{Requests: requests}).Do()
	if err != nil {
		return fmt.Errorf("unable to update text formatting: %v", err)
	}
	return nil
}

// RangeStart returns the 0-based column and row of the top-left cell of an A1 range like "B3:D5".
func RangeStart(cellRange string) (col, row int, err error) {
	start, _, _ := strings.Cut(cellRange, ":")
	col, row, err = a1ToColRow(start)
	if err != nil {
		return 0, 0, fmt.Errorf("range must start with a cell reference such as B3: %s", cellRange)
	}
	return col, row, nil
}

// FirstSheetID returns the sheetId of the leftmost sheet, which the API uses
// when a range has no sheet name.
func FirstSheetID(svc *sheets.Service, spreadsheetID string) (int64, error) {
	allSheets, err := GetSheets(svc, spreadsheetID)
	if err != nil {
		return 0, err
	}
	for _, s := range allSheets {
		if s.Index == 0 {
			return s.SheetID, nil
		}
	}
	return 0, fmt.Errorf("spreadsheet has no sheets")
}

// AppendValues appends rows to a spreadsheet
func AppendValues(svc *sheets.Service, spreadsheetID, rangeStr string, values [][]any) error {
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

// RenameSheet renames a worksheet tab.
func RenameSheet(svc *sheets.Service, spreadsheetID string, sheetID int64, newTitle string) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: sheetID,
						Title:   newTitle,
					},
					Fields: "title",
				},
			},
		},
	}
	_, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to rename sheet: %v", err)
	}
	return nil
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

// GetNote returns the note on a single cell identified by sheet name and A1 notation.
func GetNote(svc *sheets.Service, spreadsheetID, sheet, cell string) (string, error) {
	rangeStr := BuildRange(sheet, cell)
	resp, err := svc.Spreadsheets.Get(spreadsheetID).
		Ranges(rangeStr).
		IncludeGridData(true).
		Fields("sheets/data/rowData/values/note").
		Do()
	if err != nil {
		return "", fmt.Errorf("unable to get note: %v", err)
	}
	if len(resp.Sheets) == 0 ||
		len(resp.Sheets[0].Data) == 0 ||
		len(resp.Sheets[0].Data[0].RowData) == 0 ||
		len(resp.Sheets[0].Data[0].RowData[0].Values) == 0 {
		return "", nil
	}
	return resp.Sheets[0].Data[0].RowData[0].Values[0].Note, nil
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
func ParseValues(input string) [][]any {
	var result [][]any
	rows := strings.Split(input, ";")
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		cells := strings.Split(row, ",")
		var rowData []any
		for _, cell := range cells {
			rowData = append(rowData, strings.TrimSpace(cell))
		}
		result = append(result, rowData)
	}
	return result
}

// ParseValuesFromReader reads CSV-formatted data from a reader
func ParseValuesFromReader(r io.Reader) ([][]any, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	var result [][]any
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV input: %v", err)
		}
		var row []any
		for _, cell := range record {
			row = append(row, cell)
		}
		result = append(result, row)
	}
	return result, nil
}
