package utils

import (
	"time"
)

// Ticker always executes the runner function after the interval has expired.
// Every message on the done channel will stop the execution.
// Please note that this function will block until it receives the done message.
func Ticker(runner func(), done chan bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			runner()
		}
	}
}
