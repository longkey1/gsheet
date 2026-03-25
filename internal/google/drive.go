package google

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveService wraps the Google Drive API service
type DriveService struct {
	*drive.Service
}

// NewDriveService creates a new Drive service with the given authenticator
func NewDriveService(ctx context.Context, auth Authenticator) (*DriveService, error) {
	client, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %v", err)
	}

	var srv *drive.Service
	if client != nil {
		srv, err = drive.NewService(ctx, option.WithHTTPClient(client))
	} else {
		srv, err = drive.NewService(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %v", err)
	}

	return &DriveService{srv}, nil
}
