package testutils

import (
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// Check the latest log entry for the given prefix.
func CheckLogEntry(t *testing.T, hook *test.Hook, prefix string) {
	entry := hook.LastEntry()
	if entry == nil {
		t.Errorf("Entry is nil")
		return
	}
	assert.True(t, strings.HasPrefix(entry.Message, prefix), "Should have prefix: '%s'; got [%s] '%s'", prefix, entry.Level, entry.Message)
}
