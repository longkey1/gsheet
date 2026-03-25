package google

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsService wraps the Google Sheets API service
type SheetsService struct {
	*sheets.Service
}

// NewSheetsService creates a new Sheets service with the given authenticator
func NewSheetsService(ctx context.Context, auth Authenticator) (*SheetsService, error) {
	client, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %v", err)
	}

	var srv *sheets.Service
	if client != nil {
		srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	} else {
		// Use Application Default Credentials (for Service Account)
		srv, err = sheets.NewService(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %v", err)
	}

	return &SheetsService{srv}, nil
}
