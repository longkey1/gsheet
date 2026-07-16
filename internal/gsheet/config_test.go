package gsheet

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// LoadConfig reads from the global viper instance, so these subtests
// reset it and must not run in parallel.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Config
	}{
		{
			name: "full oauth config",
			content: `auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
`,
			want: &Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
			},
		},
		{
			name: "service account config",
			content: `auth_type = "service_account"
application_credentials = "/path/to/sa.json"
`,
			want: &Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/to/sa.json",
			},
		},
		{
			name: "missing auth_type defaults to oauth",
			content: `application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
`,
			want: &Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
			},
		},
		{
			name:    "empty config defaults to oauth",
			content: "",
			want:    &Config{AuthType: AuthTypeOAuth},
		},
		{
			name: "read_only enabled",
			content: `application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
read_only = true
`,
			want: &Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
				ReadOnly:                     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			got, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "oauth with both credentials",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
			},
		},
		{
			name: "service account with application credentials only",
			config: Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/to/sa.json",
			},
		},
		{
			name: "missing application credentials",
			config: Config{
				AuthType:              AuthTypeOAuth,
				GoogleUserCredentials: "/path/to/token.json",
			},
			wantErr: true,
		},
		{
			name: "oauth without user credentials",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
