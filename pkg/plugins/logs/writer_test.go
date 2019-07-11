package logs

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

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
		{"no buffer", nil, nil, "message", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &writer{
				buffer:   tt.buffer,
				upstream: tt.upstream,
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
