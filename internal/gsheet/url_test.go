package gsheet

import (
	"reflect"
	"testing"
)

func TestParseSheetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    *SheetURL
		wantErr bool
	}{
		{
			name: "plain edit URL without gid",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit",
			want: &SheetURL{SpreadsheetID: "abc123"},
		},
		{
			name: "gid in fragment",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit#gid=42",
			want: &SheetURL{SpreadsheetID: "abc123", GID: 42, HasGID: true},
		},
		{
			name: "gid zero in fragment",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit#gid=0",
			want: &SheetURL{SpreadsheetID: "abc123", GID: 0, HasGID: true},
		},
		{
			name: "gid in query string",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit?gid=7",
			want: &SheetURL{SpreadsheetID: "abc123", GID: 7, HasGID: true},
		},
		{
			name: "fragment gid wins over query gid",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit?gid=1#gid=2",
			want: &SheetURL{SpreadsheetID: "abc123", GID: 2, HasGID: true},
		},
		{
			name: "range in fragment",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit#gid=0&range=A1:B2",
			want: &SheetURL{SpreadsheetID: "abc123", GID: 0, HasGID: true, Range: "A1:B2"},
		},
		{
			name: "range in query string",
			raw:  "https://docs.google.com/spreadsheets/d/abc123/edit?range=C3",
			want: &SheetURL{SpreadsheetID: "abc123", Range: "C3"},
		},
		{
			name: "spreadsheet ID with underscore and hyphen",
			raw:  "https://docs.google.com/spreadsheets/d/a_b-C9/edit#gid=5",
			want: &SheetURL{SpreadsheetID: "a_b-C9", GID: 5, HasGID: true},
		},
		{
			name: "bare spreadsheet path",
			raw:  "https://docs.google.com/spreadsheets/d/abc123",
			want: &SheetURL{SpreadsheetID: "abc123"},
		},
		{
			name:    "non numeric gid",
			raw:     "https://docs.google.com/spreadsheets/d/abc123/edit#gid=abc",
			wantErr: true,
		},
		{
			name:    "not a spreadsheet URL",
			raw:     "https://docs.google.com/document/d/abc123/edit",
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "unparsable URL",
			raw:     ":bad",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseSheetURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSheetURL(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSheetURL(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}
