package logs

import (
	"fmt"
	"io"
)

// writer implements the io.Writer interface to use this as output for all logged
// messages when run the logs plugin.
type writer struct {
	buffer   *buffer
	upstream io.Writer
}

// Write writes the log entry to the ring buffer for the ui.
func (w *writer) Write(p []byte) (n int, err error) {
	if w.buffer == nil {
		return 0, fmt.Errorf("can not write entry: no buffer available")
	}

	entry := string(p)
	w.buffer.Write(entry)
	if w.upstream != nil {
		return w.upstream.Write(p)
	}
	return len(p), nil
}
