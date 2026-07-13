package gsheet

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFormatCellDataJSON(t *testing.T) {
	t.Parallel()

	data := &CellData{
		Range:  "Sheet1!A1:B2",
		Values: [][]string{{"a", "b"}, {"c", "d"}},
	}

	var buf bytes.Buffer
	if err := FormatCellData(&buf, data, OutputFormatJSON); err != nil {
		t.Fatalf("FormatCellData() error = %v", err)
	}

	var got CellData
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if !reflect.DeepEqual(&got, data) {
		t.Errorf("FormatCellData() JSON round trip = %+v, want %+v", got, *data)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("FormatCellData() JSON output must end with a newline")
	}
}

func TestFormatCellDataCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data *CellData
		want string
	}{
		{
			name: "simple rows",
			data: &CellData{Values: [][]string{{"a", "b"}, {"c", "d"}}},
			want: "a,b\nc,d\n",
		},
		{
			name: "values needing quoting",
			data: &CellData{Values: [][]string{{"x,y", `he said "hi"`}}},
			want: "\"x,y\",\"he said \"\"hi\"\"\"\n",
		},
		{
			name: "empty values produce no output",
			data: &CellData{Values: nil},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := FormatCellData(&buf, tt.data, OutputFormatCSV); err != nil {
				t.Fatalf("FormatCellData() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("FormatCellData() CSV = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCellDataText(t *testing.T) {
	t.Parallel()

	t.Run("empty values", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatCellData(&buf, &CellData{}, OutputFormatText); err != nil {
			t.Fatalf("FormatCellData() error = %v", err)
		}
		if got, want := buf.String(), "No data found.\n"; got != want {
			t.Errorf("FormatCellData() = %q, want %q", got, want)
		}
	})

	t.Run("first row becomes header", func(t *testing.T) {
		t.Parallel()

		data := &CellData{Values: [][]string{{"name", "age"}, {"alice", "30"}}}
		var buf bytes.Buffer
		if err := FormatCellData(&buf, data, OutputFormatText); err != nil {
			t.Fatalf("FormatCellData() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"NAME", "AGE", "alice", "30"} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatCellData() output %q must contain %q", got, want)
			}
		}
	})
}

func TestFormatSpreadsheetList(t *testing.T) {
	t.Parallel()

	spreadsheets := []SpreadsheetInfo{
		{ID: "id1", Name: "Budget", CreatedTime: "2026-01-01", ModifiedTime: "2026-02-01", WebViewLink: "https://example.com/1"},
		{ID: "id2", Name: "Roster", CreatedTime: "2026-03-01", ModifiedTime: "2026-04-01", WebViewLink: "https://example.com/2"},
	}

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatSpreadsheetList(&buf, spreadsheets, OutputFormatJSON); err != nil {
			t.Fatalf("FormatSpreadsheetList() error = %v", err)
		}
		var got []SpreadsheetInfo
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
		}
		if !reflect.DeepEqual(got, spreadsheets) {
			t.Errorf("FormatSpreadsheetList() JSON round trip = %+v, want %+v", got, spreadsheets)
		}
	})

	t.Run("text table", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatSpreadsheetList(&buf, spreadsheets, OutputFormatText); err != nil {
			t.Fatalf("FormatSpreadsheetList() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"ID", "NAME", "MODIFIED", "id1", "Budget", "id2", "Roster"} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatSpreadsheetList() output %q must contain %q", got, want)
			}
		}
	})
}

func TestFormatSheetList(t *testing.T) {
	t.Parallel()

	sheetInfos := []SheetInfo{
		{SheetID: 0, Title: "Sheet1", Index: 0, RowCount: 1000, ColCount: 26},
		{SheetID: 123, Title: "Data", Index: 1, RowCount: 50, ColCount: 5},
	}

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatSheetList(&buf, sheetInfos, OutputFormatJSON); err != nil {
			t.Fatalf("FormatSheetList() error = %v", err)
		}
		var got []SheetInfo
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
		}
		if !reflect.DeepEqual(got, sheetInfos) {
			t.Errorf("FormatSheetList() JSON round trip = %+v, want %+v", got, sheetInfos)
		}
	})

	t.Run("text table", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatSheetList(&buf, sheetInfos, OutputFormatText); err != nil {
			t.Fatalf("FormatSheetList() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"SHEET ID", "TITLE", "Sheet1", "Data", "123", "1000"} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatSheetList() output %q must contain %q", got, want)
			}
		}
	})

	t.Run("empty text", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatSheetList(&buf, nil, OutputFormatText); err != nil {
			t.Fatalf("FormatSheetList() error = %v", err)
		}
		if got, want := buf.String(), "No sheets found.\n"; got != want {
			t.Errorf("FormatSheetList() = %q, want %q", got, want)
		}
	})
}

func TestFormatCellLocations(t *testing.T) {
	t.Parallel()

	locations := []CellLocation{
		{Sheet: "Sheet1", Row: 0, Col: 1, Cell: "B1", Value: "hello"},
		{Sheet: "Sheet1", Row: 2, Col: 0, Cell: "A3", Value: "world"},
	}

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatCellLocations(&buf, locations, OutputFormatJSON); err != nil {
			t.Fatalf("FormatCellLocations() error = %v", err)
		}
		var got []CellLocation
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
		}
		if !reflect.DeepEqual(got, locations) {
			t.Errorf("FormatCellLocations() JSON round trip = %+v, want %+v", got, locations)
		}
	})

	t.Run("text table", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatCellLocations(&buf, locations, OutputFormatText); err != nil {
			t.Fatalf("FormatCellLocations() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"SHEET", "CELL", "VALUE", "B1", "hello", "A3", "world"} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatCellLocations() output %q must contain %q", got, want)
			}
		}
	})

	t.Run("empty text", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatCellLocations(&buf, nil, OutputFormatText); err != nil {
			t.Fatalf("FormatCellLocations() error = %v", err)
		}
		if got, want := buf.String(), "No matching cells found.\n"; got != want {
			t.Errorf("FormatCellLocations() = %q, want %q", got, want)
		}
	})
}
