package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkReadOnly reads from the global viper instance, so these subtests
// reset it and must not run in parallel.
func TestCheckReadOnly(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		readOnly   bool
		wantErr    bool
	}{
		{
			name:       "write command blocked when read only",
			annotation: "true",
			readOnly:   true,
			wantErr:    true,
		},
		{
			name:       "write command allowed when not read only",
			annotation: "true",
			readOnly:   false,
		},
		{
			name:     "read command allowed when read only",
			readOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			viper.Set("read_only", tt.readOnly)

			cmd := &cobra.Command{Use: "test"}
			if tt.annotation != "" {
				cmd.Annotations = map[string]string{"write": tt.annotation}
			}

			err := checkReadOnly(cmd, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkReadOnly() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
