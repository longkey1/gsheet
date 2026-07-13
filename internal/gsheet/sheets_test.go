package gsheet

import (
	"reflect"
	"strings"
	"testing"
)

func TestA1ToColRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cell    string
		wantCol int
		wantRow int
		wantErr bool
	}{
		{name: "first cell", cell: "A1", wantCol: 0, wantRow: 0},
		{name: "single letter column", cell: "C4", wantCol: 2, wantRow: 3},
		{name: "last single letter column", cell: "Z10", wantCol: 25, wantRow: 9},
		{name: "double letter column", cell: "AA1", wantCol: 26, wantRow: 0},
		{name: "double letter column AZ", cell: "AZ2", wantCol: 51, wantRow: 1},
		{name: "double letter column BA", cell: "BA1", wantCol: 52, wantRow: 0},
		{name: "large row number", cell: "B1000", wantCol: 1, wantRow: 999},
		{name: "lowercase is accepted", cell: "b3", wantCol: 1, wantRow: 2},
		{name: "surrounding whitespace is trimmed", cell: "  C4  ", wantCol: 2, wantRow: 3},
		{name: "row only", cell: "123", wantErr: true},
		{name: "column only", cell: "ABC", wantErr: true},
		{name: "empty string", cell: "", wantErr: true},
		{name: "row zero", cell: "A0", wantErr: true},
		{name: "negative row", cell: "A-1", wantErr: true},
		{name: "digit before column", cell: "1A", wantErr: true},
		{name: "trailing letters after row", cell: "A1B", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			col, row, err := a1ToColRow(tt.cell)
			if (err != nil) != tt.wantErr {
				t.Fatalf("a1ToColRow(%q) error = %v, wantErr %v", tt.cell, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if col != tt.wantCol || row != tt.wantRow {
				t.Errorf("a1ToColRow(%q) = (%d, %d), want (%d, %d)", tt.cell, col, row, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestColRowToA1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		col  int
		row  int
		want string
	}{
		{name: "first cell", col: 0, row: 0, want: "A1"},
		{name: "single letter column", col: 2, row: 3, want: "C4"},
		{name: "last single letter column", col: 25, row: 0, want: "Z1"},
		{name: "first double letter column", col: 26, row: 0, want: "AA1"},
		{name: "double letter column AZ", col: 51, row: 1, want: "AZ2"},
		{name: "double letter column BA", col: 52, row: 0, want: "BA1"},
		{name: "last double letter column", col: 701, row: 0, want: "ZZ1"},
		{name: "first triple letter column", col: 702, row: 0, want: "AAA1"},
		{name: "large row", col: 1, row: 999, want: "B1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := colRowToA1(tt.col, tt.row); got != tt.want {
				t.Errorf("colRowToA1(%d, %d) = %q, want %q", tt.col, tt.row, got, tt.want)
			}
		})
	}
}

// A1 conversion should round-trip in both directions.
func TestA1RoundTrip(t *testing.T) {
	t.Parallel()

	for col := 0; col < 800; col += 7 {
		for row := 0; row < 100; row += 11 {
			cell := colRowToA1(col, row)
			gotCol, gotRow, err := a1ToColRow(cell)
			if err != nil {
				t.Fatalf("a1ToColRow(%q) error = %v", cell, err)
			}
			if gotCol != col || gotRow != row {
				t.Errorf("round trip (%d, %d) -> %q -> (%d, %d)", col, row, cell, gotCol, gotRow)
			}
		}
	}
}

func TestBuildRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sheet     string
		cellRange string
		want      string
	}{
		{name: "sheet and range", sheet: "Sheet1", cellRange: "A1:C10", want: "Sheet1!A1:C10"},
		{name: "sheet only", sheet: "Sheet1", cellRange: "", want: "Sheet1"},
		{name: "range only", sheet: "", cellRange: "A1:C10", want: "A1:C10"},
		{name: "both empty", sheet: "", cellRange: "", want: ""},
		{name: "sheet with spaces", sheet: "My Sheet", cellRange: "B2", want: "My Sheet!B2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := BuildRange(tt.sheet, tt.cellRange); got != tt.want {
				t.Errorf("BuildRange(%q, %q) = %q, want %q", tt.sheet, tt.cellRange, got, tt.want)
			}
		})
	}
}

func TestParseValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  [][]any
	}{
		{
			name:  "multiple rows and cells",
			input: "a,b,c;d,e,f",
			want:  [][]any{{"a", "b", "c"}, {"d", "e", "f"}},
		},
		{
			name:  "single cell",
			input: "a",
			want:  [][]any{{"a"}},
		},
		{
			name:  "single row",
			input: "a,b",
			want:  [][]any{{"a", "b"}},
		},
		{
			name:  "cells are trimmed",
			input: " a , b ; c ,d",
			want:  [][]any{{"a", "b"}, {"c", "d"}},
		},
		{
			name:  "empty rows are skipped",
			input: "a,b;;c,d;",
			want:  [][]any{{"a", "b"}, {"c", "d"}},
		},
		{
			name:  "empty cells are kept",
			input: "a,,c",
			want:  [][]any{{"a", "", "c"}},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "only separators",
			input: ";;",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParseValues(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseValues(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseValuesFromReader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    [][]any
		wantErr bool
	}{
		{
			name:  "simple csv",
			input: "a,b\nc,d\n",
			want:  [][]any{{"a", "b"}, {"c", "d"}},
		},
		{
			name:  "quoted fields with comma and newline",
			input: "\"x,y\",\"line1\nline2\"\n",
			want:  [][]any{{"x,y", "line1\nline2"}},
		},
		{
			name:  "no trailing newline",
			input: "a,b",
			want:  [][]any{{"a", "b"}},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:    "inconsistent field count",
			input:   "a,b\nc\n",
			wantErr: true,
		},
		{
			name:    "bare quote in field",
			input:   "a\"b,c\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseValuesFromReader(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseValuesFromReader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseValuesFromReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterSheets(t *testing.T) {
	t.Parallel()

	all := []SheetInfo{
		{SheetID: 0, Title: "Sheet1"},
		{SheetID: 1, Title: "Sheet2"},
		{SheetID: 2, Title: "Summary 2025"},
		{SheetID: 3, Title: "data"},
	}

	tests := []struct {
		name     string
		query    string
		useRegex bool
		want     []SheetInfo
		wantErr  bool
	}{
		{
			name:  "empty query returns all",
			query: "",
			want:  all,
		},
		{
			name:  "substring match",
			query: "Sheet",
			want:  []SheetInfo{all[0], all[1]},
		},
		{
			name:  "substring match is case sensitive",
			query: "sheet",
			want:  nil,
		},
		{
			name:  "substring no match",
			query: "nope",
			want:  nil,
		},
		{
			name:     "regex match",
			query:    `^Sheet\d$`,
			useRegex: true,
			want:     []SheetInfo{all[0], all[1]},
		},
		{
			name:     "regex match digits",
			query:    `20\d{2}`,
			useRegex: true,
			want:     []SheetInfo{all[2]},
		},
		{
			name:     "invalid regex",
			query:    "[",
			useRegex: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FilterSheets(all, tt.query, tt.useRegex)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FilterSheets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterSheets() = %v, want %v", got, tt.want)
			}
		})
	}
}
