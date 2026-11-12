package gsheet

import (
	"slices"
	"strings"
	"unicode/utf16"

	"google.golang.org/api/sheets/v4"
)

// richStyle is the subset of a Sheets TextFormat that has a Markdown representation.
type richStyle struct {
	Bold          bool
	Italic        bool
	Strikethrough bool
	Link          string
}

// richSpan is a run of text sharing a single style.
type richSpan struct {
	Text  string
	Style richStyle
}

type mdTokenKind int

const (
	mdText mdTokenKind = iota
	mdBold
	mdItalic
	mdStrike
	mdLinkOpen
	mdLinkClose
)

// mdToken holds literal text for mdText, the raw marker for emphasis toggles,
// and the URL for mdLinkOpen.
type mdToken struct {
	kind mdTokenKind
	text string
}

// ParseMarkdown converts inline Markdown (links, bold, italic, strikethrough)
// into styled spans. Unmatched emphasis markers and brackets that do not form
// a link are kept as literal text.
func ParseMarkdown(s string) []richSpan {
	tokens := tokenizeMarkdown([]rune(s), true)
	pairEmphasis(tokens)

	var spans []richSpan
	var style richStyle
	for _, t := range tokens {
		switch t.kind {
		case mdText:
			spans = appendSpan(spans, t.text, style)
		case mdBold:
			style.Bold = !style.Bold
		case mdItalic:
			style.Italic = !style.Italic
		case mdStrike:
			style.Strikethrough = !style.Strikethrough
		case mdLinkOpen:
			style.Link = t.text
		case mdLinkClose:
			style.Link = ""
		}
	}
	return spans
}

func tokenizeMarkdown(s []rune, allowLinks bool) []mdToken {
	var tokens []mdToken
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, mdToken{kind: mdText, text: buf.String()})
			buf.Reset()
		}
	}

	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && isTextEscapable(s[i+1]):
			buf.WriteRune(s[i+1])
			i += 2
		case c == '*' && i+1 < len(s) && s[i+1] == '*':
			flush()
			tokens = append(tokens, mdToken{kind: mdBold, text: "**"})
			i += 2
		case c == '*':
			flush()
			tokens = append(tokens, mdToken{kind: mdItalic, text: "*"})
			i++
		case c == '~' && i+1 < len(s) && s[i+1] == '~':
			flush()
			tokens = append(tokens, mdToken{kind: mdStrike, text: "~~"})
			i += 2
		case c == '[' && allowLinks:
			textEnd, url, end, ok := scanLink(s, i)
			if !ok {
				buf.WriteRune(c)
				i++
				continue
			}
			flush()
			tokens = append(tokens, mdToken{kind: mdLinkOpen, text: url})
			tokens = append(tokens, tokenizeMarkdown(s[i+1:textEnd], false)...)
			tokens = append(tokens, mdToken{kind: mdLinkClose})
			i = end
		default:
			buf.WriteRune(c)
			i++
		}
	}
	flush()
	return tokens
}

// scanLink checks whether s[start:] begins with [text](url). It returns the
// index of the closing bracket, the unescaped URL and the index just past the
// closing parenthesis.
func scanLink(s []rune, start int) (textEnd int, url string, end int, ok bool) {
	j := start + 1
	for ; j < len(s); j++ {
		if s[j] == '\\' && j+1 < len(s) && isTextEscapable(s[j+1]) {
			j++
			continue
		}
		if s[j] == '[' {
			return 0, "", 0, false
		}
		if s[j] == ']' {
			break
		}
	}
	if j >= len(s) || j == start+1 || j+1 >= len(s) || s[j+1] != '(' {
		return 0, "", 0, false
	}

	var buf strings.Builder
	depth := 0
	for k := j + 2; k < len(s); k++ {
		c := s[k]
		switch {
		case c == '\\' && k+1 < len(s) && isURLEscapable(s[k+1]):
			buf.WriteRune(s[k+1])
			k++
		case c == '\n':
			return 0, "", 0, false
		case c == '(':
			depth++
			buf.WriteRune(c)
		case c == ')' && depth > 0:
			depth--
			buf.WriteRune(c)
		case c == ')':
			if buf.Len() == 0 {
				return 0, "", 0, false
			}
			return j, buf.String(), k + 1, true
		default:
			buf.WriteRune(c)
		}
	}
	return 0, "", 0, false
}

// pairEmphasis turns the last marker of each kind into literal text when the
// markers of that kind cannot all be paired.
func pairEmphasis(tokens []mdToken) {
	for _, kind := range []mdTokenKind{mdBold, mdItalic, mdStrike} {
		last, count := -1, 0
		for i, t := range tokens {
			if t.kind == kind {
				last = i
				count++
			}
		}
		if count%2 == 1 {
			tokens[last].kind = mdText
		}
	}
}

func appendSpan(spans []richSpan, text string, style richStyle) []richSpan {
	if text == "" {
		return spans
	}
	if n := len(spans); n > 0 && spans[n-1].Style == style {
		spans[n-1].Text += text
		return spans
	}
	return append(spans, richSpan{Text: text, Style: style})
}

func isTextEscapable(r rune) bool {
	return strings.ContainsRune(`\*~[]`, r)
}

func isURLEscapable(r rune) bool {
	return strings.ContainsRune(`\()`, r)
}

// mdMarker is an open Markdown construct while rendering. Links are compared
// by URL so that adjacent links to different targets are closed and reopened.
type mdMarker struct {
	kind mdTokenKind
	url  string
}

// RenderMarkdown converts styled spans back into inline Markdown.
// The output parses back to the same spans with ParseMarkdown.
func RenderMarkdown(spans []richSpan) string {
	var b strings.Builder
	var stack []mdMarker

	for _, sp := range spans {
		want := markersFor(sp.Style)

		keep := 0
		for keep < len(stack) && slices.Contains(want, stack[keep]) {
			keep++
		}
		for len(stack) > keep {
			b.WriteString(closeMarker(stack[len(stack)-1]))
			stack = stack[:len(stack)-1]
		}
		for _, m := range want {
			if !slices.Contains(stack, m) {
				b.WriteString(openMarker(m))
				stack = append(stack, m)
			}
		}
		b.WriteString(escapeText(sp.Text))
	}
	for len(stack) > 0 {
		b.WriteString(closeMarker(stack[len(stack)-1]))
		stack = stack[:len(stack)-1]
	}
	return b.String()
}

// markersFor lists markers outermost first. The link comes last so that
// emphasis spanning a link wraps it instead of being split around it.
func markersFor(st richStyle) []mdMarker {
	var ms []mdMarker
	if st.Bold {
		ms = append(ms, mdMarker{kind: mdBold})
	}
	if st.Italic {
		ms = append(ms, mdMarker{kind: mdItalic})
	}
	if st.Strikethrough {
		ms = append(ms, mdMarker{kind: mdStrike})
	}
	if st.Link != "" {
		ms = append(ms, mdMarker{kind: mdLinkOpen, url: st.Link})
	}
	return ms
}

func openMarker(m mdMarker) string {
	switch m.kind {
	case mdBold:
		return "**"
	case mdItalic:
		return "*"
	case mdStrike:
		return "~~"
	default:
		return "["
	}
}

func closeMarker(m mdMarker) string {
	if m.kind == mdLinkOpen {
		return "](" + escapeURL(m.url) + ")"
	}
	return openMarker(m)
}

func escapeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isTextEscapable(r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func escapeURL(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isURLEscapable(r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// spansToCell returns the plain cell text and its text format runs.
// Runs are nil when no span is styled.
func spansToCell(spans []richSpan) (string, []*sheets.TextFormatRun) {
	var text strings.Builder
	styled := false
	for _, sp := range spans {
		text.WriteString(sp.Text)
		if sp.Style != (richStyle{}) {
			styled = true
		}
	}
	if !styled {
		return text.String(), nil
	}

	// Sheets counts run positions in UTF-16 code units.
	var runs []*sheets.TextFormatRun
	offset := 0
	for _, sp := range spans {
		runs = append(runs, &sheets.TextFormatRun{
			StartIndex: int64(offset),
			Format:     toTextFormat(sp.Style),
		})
		offset += len(utf16.Encode([]rune(sp.Text)))
	}
	return text.String(), runs
}

func toTextFormat(st richStyle) *sheets.TextFormat {
	f := &sheets.TextFormat{
		Bold:          st.Bold,
		Italic:        st.Italic,
		Strikethrough: st.Strikethrough,
	}
	if st.Link != "" {
		f.Link = &sheets.Link{Uri: st.Link}
	}
	return f
}

// cellToSpans splits cell text into spans by its text format runs, on top of
// the cell-level format. cellLink applies to the whole text when no format
// carries a link; it covers cells such as =HYPERLINK() whose link is not
// expressed as a format.
func cellToSpans(text string, runs []*sheets.TextFormatRun, cellFormat *sheets.TextFormat, cellLink string) []richSpan {
	base := fromTextFormat(cellFormat, richStyle{})
	hasLink := base.Link != ""
	for _, r := range runs {
		if r.Format != nil && r.Format.Link != nil && r.Format.Link.Uri != "" {
			hasLink = true
			break
		}
	}
	if !hasLink {
		base.Link = cellLink
	}

	u := utf16.Encode([]rune(text))
	var spans []richSpan
	pos := 0
	style := base
	for _, r := range runs {
		start := min(max(int(r.StartIndex), pos), len(u))
		spans = appendSpan(spans, string(utf16.Decode(u[pos:start])), style)
		style = fromTextFormat(r.Format, base)
		pos = start
	}
	return appendSpan(spans, string(utf16.Decode(u[pos:])), style)
}

// fromTextFormat overlays f on base. A field missing from a run inherits the
// cell-level value, and the client cannot tell a missing field from false, so
// only true overrides.
func fromTextFormat(f *sheets.TextFormat, base richStyle) richStyle {
	st := base
	if f == nil {
		return st
	}
	st.Bold = st.Bold || f.Bold
	st.Italic = st.Italic || f.Italic
	st.Strikethrough = st.Strikethrough || f.Strikethrough
	if f.Link != nil && f.Link.Uri != "" {
		st.Link = f.Link.Uri
	}
	return st
}
