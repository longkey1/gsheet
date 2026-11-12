package gsheet

import (
	"reflect"
	"testing"

	"google.golang.org/api/sheets/v4"
)

func TestParseMarkdown(t *testing.T) {
	t.Parallel()

	const jira = "https://foo.atlassian.net/browse/PROJ-1"

	tests := []struct {
		name  string
		input string
		want  []richSpan
	}{
		{name: "plain", input: "hello", want: []richSpan{{Text: "hello"}}},
		{name: "empty", input: "", want: nil},
		{
			name:  "link in text",
			input: "対応: [PROJ-1](" + jira + ") 済",
			want: []richSpan{
				{Text: "対応: "},
				{Text: "PROJ-1", Style: richStyle{Link: jira}},
				{Text: " 済"},
			},
		},
		{
			name:  "adjacent links separated by comma",
			input: "[A](https://a), [B](https://b)",
			want: []richSpan{
				{Text: "A", Style: richStyle{Link: "https://a"}},
				{Text: ", "},
				{Text: "B", Style: richStyle{Link: "https://b"}},
			},
		},
		{
			name:  "bold italic strikethrough",
			input: "**b** *i* ~~s~~",
			want: []richSpan{
				{Text: "b", Style: richStyle{Bold: true}},
				{Text: " "},
				{Text: "i", Style: richStyle{Italic: true}},
				{Text: " "},
				{Text: "s", Style: richStyle{Strikethrough: true}},
			},
		},
		{
			name:  "bold italic combined",
			input: "***x***",
			want:  []richSpan{{Text: "x", Style: richStyle{Bold: true, Italic: true}}},
		},
		{
			name:  "bold inside link text",
			input: "[a**b**](https://x)",
			want: []richSpan{
				{Text: "a", Style: richStyle{Link: "https://x"}},
				{Text: "b", Style: richStyle{Link: "https://x", Bold: true}},
			},
		},
		{
			name:  "url with balanced parentheses",
			input: "[w](https://en.wikipedia.org/wiki/Go_(language))",
			want:  []richSpan{{Text: "w", Style: richStyle{Link: "https://en.wikipedia.org/wiki/Go_(language)"}}},
		},
		{name: "brackets without url are literal", input: "[WIP] task", want: []richSpan{{Text: "[WIP] task"}}},
		{name: "empty link text is literal", input: "[](https://x)", want: []richSpan{{Text: "[](https://x)"}}},
		{name: "unclosed bold is literal", input: "a ** b", want: []richSpan{{Text: "a ** b"}}},
		{name: "unclosed italic is literal", input: "5*3", want: []richSpan{{Text: "5*3"}}},
		{name: "escaped markers", input: `\*a\* \[b\] \~\~c \\`, want: []richSpan{{Text: `*a* [b] ~~c \`}}},
		{name: "backslash before ordinary char is literal", input: `C:\path`, want: []richSpan{{Text: `C:\path`}}},
		{name: "single tilde is literal", input: "10~20", want: []richSpan{{Text: "10~20"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseMarkdown(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMarkdown(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		spans []richSpan
		want  string
	}{
		{name: "plain with special chars", spans: []richSpan{{Text: `5*3 [x] ~ \`}}, want: `5\*3 \[x\] \~ \\`},
		{
			name: "bold wraps link",
			spans: []richSpan{
				{Text: "a", Style: richStyle{Bold: true}},
				{Text: "b", Style: richStyle{Bold: true, Link: "https://x"}},
				{Text: "c", Style: richStyle{Bold: true}},
			},
			want: "**a[b](https://x)c**",
		},
		{
			name: "adjacent links with different urls",
			spans: []richSpan{
				{Text: "A", Style: richStyle{Link: "https://a"}},
				{Text: "B", Style: richStyle{Link: "https://b"}},
			},
			want: "[A](https://a)[B](https://b)",
		},
		{name: "url escaping", spans: []richSpan{{Text: "x", Style: richStyle{Link: `https://x/a)b`}}}, want: `[x](https://x/a\)b)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := RenderMarkdown(tt.spans); got != tt.want {
				t.Errorf("RenderMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMarkdownRoundTrip(t *testing.T) {
	t.Parallel()

	spansList := [][]richSpan{
		{{Text: "plain *with* [chars] ~~ \\"}},
		{{Text: "対応: "}, {Text: "PROJ-1", Style: richStyle{Link: "https://x/PROJ-1"}}, {Text: " 😀"}},
		{{Text: "i", Style: richStyle{Italic: true}}, {Text: "b", Style: richStyle{Bold: true}}},
		{{Text: "b", Style: richStyle{Bold: true}}, {Text: "i", Style: richStyle{Italic: true}}},
		{{Text: "s", Style: richStyle{Strikethrough: true, Bold: true}}, {Text: "l", Style: richStyle{Link: "https://x/(y)"}}},
		{{Text: "a", Style: richStyle{Link: "https://x", Italic: true}}, {Text: "b", Style: richStyle{Link: "https://x"}}},
	}

	for _, spans := range spansList {
		md := RenderMarkdown(spans)
		if got := ParseMarkdown(md); !reflect.DeepEqual(got, spans) {
			t.Errorf("round trip via %q = %#v, want %#v", md, got, spans)
		}
	}
}

func TestSpansToCell(t *testing.T) {
	t.Parallel()

	t.Run("no style yields no runs", func(t *testing.T) {
		t.Parallel()
		text, runs := spansToCell([]richSpan{{Text: "abc"}})
		if text != "abc" || runs != nil {
			t.Errorf("spansToCell() = %q, %v, want %q, nil", text, runs, "abc")
		}
	})

	t.Run("indexes are utf16", func(t *testing.T) {
		t.Parallel()
		// "😀" is two UTF-16 code units, "あ" is one.
		text, runs := spansToCell([]richSpan{
			{Text: "😀あ"},
			{Text: "K-1", Style: richStyle{Link: "https://x"}},
			{Text: "!"},
		})
		if text != "😀あK-1!" {
			t.Errorf("text = %q", text)
		}
		var starts []int64
		for _, r := range runs {
			starts = append(starts, r.StartIndex)
		}
		if want := []int64{0, 3, 6}; !reflect.DeepEqual(starts, want) {
			t.Errorf("starts = %v, want %v", starts, want)
		}
		if runs[1].Format.Link == nil || runs[1].Format.Link.Uri != "https://x" {
			t.Errorf("link run format = %#v", runs[1].Format)
		}
		if runs[2].Format.Link != nil {
			t.Errorf("run after link should have no link: %#v", runs[2].Format)
		}
	})
}

func TestRangeStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input            string
		wantCol, wantRow int
		wantErr          bool
	}{
		{input: "B3", wantCol: 1, wantRow: 2},
		{input: "B3:D5", wantCol: 1, wantRow: 2},
		{input: "B:B", wantErr: true},
		{input: "3:3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			col, row, err := RangeStart(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RangeStart(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && (col != tt.wantCol || row != tt.wantRow) {
				t.Errorf("RangeStart(%q) = (%d, %d), want (%d, %d)", tt.input, col, row, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestCellToSpans(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		text       string
		runs       []*sheets.TextFormatRun
		cellFormat *sheets.TextFormat
		cellLink   string
		want       []richSpan
	}{
		{
			name:       "cell-level format applies to whole text",
			text:       "abc",
			cellFormat: &sheets.TextFormat{Bold: true, Link: &sheets.Link{Uri: "https://x"}},
			cellLink:   "https://x",
			want:       []richSpan{{Text: "abc", Style: richStyle{Bold: true, Link: "https://x"}}},
		},
		{
			name:       "runs inherit cell-level format",
			text:       "ab",
			cellFormat: &sheets.TextFormat{Bold: true},
			runs: []*sheets.TextFormatRun{
				{StartIndex: 0, Format: &sheets.TextFormat{Italic: true}},
				{StartIndex: 1, Format: &sheets.TextFormat{}},
			},
			want: []richSpan{
				{Text: "a", Style: richStyle{Bold: true, Italic: true}},
				{Text: "b", Style: richStyle{Bold: true}},
			},
		},
		{name: "no runs", text: "abc", want: []richSpan{{Text: "abc"}}},
		{name: "empty", text: "", want: nil},
		{
			name: "utf16 run positions",
			text: "😀あK-1!",
			runs: []*sheets.TextFormatRun{
				{StartIndex: 0, Format: &sheets.TextFormat{}},
				{StartIndex: 3, Format: &sheets.TextFormat{Link: &sheets.Link{Uri: "https://x"}}},
				{StartIndex: 6, Format: &sheets.TextFormat{}},
			},
			want: []richSpan{
				{Text: "😀あ"},
				{Text: "K-1", Style: richStyle{Link: "https://x"}},
				{Text: "!"},
			},
		},
		{
			name: "first run not at zero",
			text: "abcd",
			runs: []*sheets.TextFormatRun{{StartIndex: 2, Format: &sheets.TextFormat{Bold: true}}},
			want: []richSpan{{Text: "ab"}, {Text: "cd", Style: richStyle{Bold: true}}},
		},
		{
			name:     "cell link applies to whole text",
			text:     "site",
			cellLink: "https://x",
			want:     []richSpan{{Text: "site", Style: richStyle{Link: "https://x"}}},
		},
		{
			name: "cell link ignored when runs carry links",
			text: "ab",
			runs: []*sheets.TextFormatRun{
				{StartIndex: 0, Format: &sheets.TextFormat{Link: &sheets.Link{Uri: "https://x"}}},
				{StartIndex: 1, Format: &sheets.TextFormat{}},
			},
			cellLink: "https://x",
			want:     []richSpan{{Text: "a", Style: richStyle{Link: "https://x"}}, {Text: "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cellToSpans(tt.text, tt.runs, tt.cellFormat, tt.cellLink)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cellToSpans() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
