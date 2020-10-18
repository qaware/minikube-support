package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, tt.wantWriter, writer.String())
		})
	}
}

func TestFormatAsTable(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		header  string
		want    string
		wantErr bool
	}{
		{"no entries", []string{}, "a\tb\n", "a |b\n", false},
		{"one entry", []string{"a\tb\n"}, "a\tb\n", "a |b\na |b\n", false},
		{"two entry", []string{"b\ta\n", "a\tb\n"}, "a\tb\n", "a |b\na |b\nb |a\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatAsTable(tt.entries, tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatAsTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
