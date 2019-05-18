package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
