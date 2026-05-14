package gsheet

import (
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"strconv"

	"google.golang.org/api/sheets/v4"
)

// SheetURL holds the parsed components of a Google Sheets URL.
type SheetURL struct {
	SpreadsheetID string
	GID           int64
	HasGID        bool
	Range         string
}

var spreadsheetIDPattern = regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9_-]+)`)

// ParseSheetURL extracts the spreadsheet ID, gid, and range from a Google Sheets URL.
// The gid and range are read from the URL fragment first (Google's default), then the query string.
func ParseSheetURL(raw string) (*SheetURL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	m := spreadsheetIDPattern.FindStringSubmatch(u.Path)
	if len(m) < 2 {
		return nil, fmt.Errorf("could not find spreadsheet ID in URL: %s", raw)
	}
	result := &SheetURL{SpreadsheetID: m[1]}

	params := u.Query()
	if u.Fragment != "" {
		if fragParams, err := url.ParseQuery(u.Fragment); err == nil {
			maps.Copy(params, fragParams)
		}
	}

	if gidStr := params.Get("gid"); gidStr != "" {
		gid, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid gid in URL: %s", gidStr)
		}
		result.GID = gid
		result.HasGID = true
	}
	if rangeStr := params.Get("range"); rangeStr != "" {
		result.Range = rangeStr
	}

	return result, nil
}

// GetSheetTitleByID returns the sheet title for a given gid (numeric sheetId).
func GetSheetTitleByID(svc *sheets.Service, spreadsheetID string, sheetID int64) (string, error) {
	allSheets, err := GetSheets(svc, spreadsheetID)
	if err != nil {
		return "", err
	}
	for _, s := range allSheets {
		if s.SheetID == sheetID {
			return s.Title, nil
		}
	}
	return "", fmt.Errorf("sheet not found: gid=%d", sheetID)
}
