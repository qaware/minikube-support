package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	done := make(chan bool)
	count := 0
	go Ticker(func() {
		count++
	}, done, 100*time.Millisecond)
	time.Sleep(250 * time.Millisecond)
	done <- true
	assert.True(t, count >= 2)
}
