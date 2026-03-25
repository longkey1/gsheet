package gsheet

import (
	"context"

	"github.com/longkey1/gsheet/internal/google"
)

// Service represents the gsheet application service
type Service struct {
	Sheets *google.SheetsService
	Drive  *google.DriveService
}

// NewService creates a new gsheet service based on the configuration
func NewService(ctx context.Context, config *Config) (*Service, error) {
	auth := newAuthenticator(config)

	sheetsSvc, err := google.NewSheetsService(ctx, auth)
	if err != nil {
		return nil, err
	}

	driveSvc, err := google.NewDriveService(ctx, auth)
	if err != nil {
		return nil, err
	}

	return &Service{
		Sheets: sheetsSvc,
		Drive:  driveSvc,
	}, nil
}

func newAuthenticator(config *Config) google.Authenticator {
	switch config.AuthType {
	case AuthTypeServiceAccount:
		return google.NewServiceAccountAuthenticator(config.GoogleApplicationCredentials)
	case AuthTypeOAuth:
		fallthrough
	default:
		return google.NewOAuthAuthenticator(
			config.GoogleApplicationCredentials,
			config.GoogleUserCredentials,
		)
	}
}
