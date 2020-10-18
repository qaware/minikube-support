package utils

import (
	"bytes"
	"testing"
)

func TestWriteSorted(t *testing.T) {
	tests := []struct {
		name       string
		entries    []string
		wantWriter string
		wantErr    bool
	}{
		{"empty", []string{}, "", false},
		{"one", []string{"abc"}, "abc", false},
		{"empty", []string{"xyz ", "abc "}, "abc xyz ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			err := WriteSorted(tt.entries, writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteSorted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("WriteSorted() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
