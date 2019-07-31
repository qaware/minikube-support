package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"testing"
)

func TestRootCommandOptions_preRun(t *testing.T) {
	tests := []struct {
		name       string
		preRunInit []PreRunInit
		wantErr    bool
	}{
		{
			"nil",
			nil,
			false,
		}, {
			"no function",
			[]PreRunInit{},
			false,
		}, {
			"one ok",
			[]PreRunInit{func(_ *RootCommandOptions) error { return nil }},
			false,
		}, {
			"one ok, one failing",
			[]PreRunInit{
				func(_ *RootCommandOptions) error { return nil },
				func(_ *RootCommandOptions) error { return fmt.Errorf("dummy") },
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RootCommandOptions{
				preRunInit: tt.preRunInit,
			}
			if err := o.preRun(&cobra.Command{}, []string{}); (err != nil) != tt.wantErr {
				t.Errorf("preRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
