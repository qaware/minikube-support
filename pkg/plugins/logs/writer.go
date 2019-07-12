package logs

import (
	"fmt"
	"io"
)

// writer implements the io.Writer interface to use this as output for all logged
// messages when run the logs plugin.
type writer struct {
	buffer        *buffer
	upstream      io.Writer
	renderTrigger chan bool
}

// newWriter creates a consistent new log entry writer.
func newWriter(buffer *buffer, renderTrigger chan bool, upstream io.Writer) (*writer, error) {
	if buffer == nil {
		return nil, fmt.Errorf("can not initialize log entry writer: no buffer available")
	}
	if renderTrigger == nil {
		return nil, fmt.Errorf("can not initialize log entry writer: no renderTrigger channel available")
	}
	return &writer{
		buffer:        buffer,
		renderTrigger: renderTrigger,
		upstream:      upstream,
	}, nil
}

// Write writes the log entry to the ring buffer for the ui.
func (w *writer) Write(p []byte) (n int, err error) {
	entry := string(p)
	w.buffer.Write(entry)
	w.renderTrigger <- true
	if w.upstream != nil {
		return w.upstream.Write(p)
	}

	return len(p), nil
}
