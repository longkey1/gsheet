package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeFile creates a file with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseInputValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent *string // non-nil writes a CSV file and passes its path
		valuesStr   string
		want        [][]any
		wantErr     bool
	}{
		{
			name:        "csv file",
			fileContent: new("a,b\nc,d\n"),
			want:        [][]any{{"a", "b"}, {"c", "d"}},
		},
		{
			name:        "csv file with quoted fields",
			fileContent: new("\"x,y\",z\n"),
			want:        [][]any{{"x,y", "z"}},
		},
		{
			name:        "file takes precedence over values",
			fileContent: new("from,file\n"),
			valuesStr:   "from,values",
			want:        [][]any{{"from", "file"}},
		},
		{
			name:        "invalid csv file",
			fileContent: new("a,b\nc\n"),
			wantErr:     true,
		},
		{
			name:      "values string",
			valuesStr: "a,b;c,d",
			want:      [][]any{{"a", "b"}, {"c", "d"}},
		},
		{
			name:    "no input at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tt.fileContent != nil {
				filePath = filepath.Join(t.TempDir(), "input.csv")
				writeFile(t, filePath, *tt.fileContent)
			}

			got, err := parseInputValues(filePath, false, tt.valuesStr, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseInputValues() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInputValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInputValuesMissingFile(t *testing.T) {
	t.Parallel()

	if _, err := parseInputValues(filepath.Join(t.TempDir(), "nosuch.csv"), false, "", false); err == nil {
		t.Error("parseInputValues() error = nil, want error")
	}
}

// The stdin tests replace os.Stdin, so they must not run in parallel.
func TestParseInputValuesStdin(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		stdinAsSingleCell bool
		want              [][]any
	}{
		{
			name:              "stdin as single cell keeps content verbatim",
			input:             "line1\nline2\n",
			stdinAsSingleCell: true,
			want:              [][]any{{"line1\nline2\n"}},
		},
		{
			name:  "stdin as csv",
			input: "a,b\nc,d\n",
			want:  [][]any{{"a", "b"}, {"c", "d"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.WriteString(tt.input); err != nil {
				t.Fatal(err)
			}
			_ = w.Close()

			orig := os.Stdin
			os.Stdin = r
			t.Cleanup(func() {
				os.Stdin = orig
				_ = r.Close()
			})

			got, err := parseInputValues("", true, "", tt.stdinAsSingleCell)
			if err != nil {
				t.Fatalf("parseInputValues() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInputValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
