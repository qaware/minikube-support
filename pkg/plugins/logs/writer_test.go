package logs

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newWriter(t *testing.T) {
	tests := []struct {
		name          string
		buffer        *buffer
		renderTrigger chan bool
		upstream      io.Writer
		wantWriter    bool
		wantUpstream  bool
		wantErr       bool
	}{
		{"ok", newBuffer(), make(chan bool), &bytes.Buffer{}, true, true, false},
		{"ok no upstream", newBuffer(), make(chan bool), nil, true, false, false},
		{"no buffer", nil, make(chan bool), nil, false, false, true},
		{"no render trigger", newBuffer(), nil, nil, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newWriter(tt.buffer, tt.renderTrigger, tt.upstream)
			if (err != nil) != tt.wantErr {
				t.Errorf("newWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Falsef(t, (got != nil) != tt.wantWriter, "writer!=nil should be %s; got %s", tt.wantWriter, got != nil)
			if got != nil {
				assert.Falsef(t, (got.upstream != nil) != tt.wantUpstream, "upstream!=nil should be %s; got %s", tt.wantUpstream, got.upstream != nil)
			}
		})
	}
}

func Test_writer_Write(t *testing.T) {
	tests := []struct {
		name     string
		buffer   *buffer
		upstream io.Writer
		message  string
		wantErr  bool
	}{
		{"ok no upstream", newBuffer(), nil, "message", false},
		{"ok upstream", newBuffer(), &dummyWriter{}, "message", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderTrigger := make(chan bool, 2)

			w := &writer{
				buffer:        tt.buffer,
				upstream:      tt.upstream,
				renderTrigger: renderTrigger,
			}
			_, err := w.Write([]byte(tt.message))
			if (err != nil) != tt.wantErr {
				t.Errorf("writer.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.buffer != nil {
				assert.Equal(t, tt.message, tt.buffer.entries[0])
			}
			if tt.upstream != nil {
				assert.Equal(t, tt.message, tt.upstream.(*dummyWriter).message)
			}
			assert.Equal(t, true, <-renderTrigger)
		})
	}
}

type dummyWriter struct {
	message string
}

func (w *dummyWriter) Write(p []byte) (n int, err error) {
	w.message = string(p)
	return 0, nil
}
