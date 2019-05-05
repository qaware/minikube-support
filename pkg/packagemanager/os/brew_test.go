// +build darwin

package os

import (
	"testing"
)

func Test_runBrewCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"list", "list", false},
		{"invalid command", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runBrewCommand(tt.command); (err != nil) != tt.wantErr {
				t.Errorf("runBrewCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
