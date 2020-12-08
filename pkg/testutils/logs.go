package testutils

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// CheckLogEntry checks the latest logged entry if the message begin with the given prefix.
// If the prefix is "" then it expects that no log messages were written.
func CheckLogEntry(t *testing.T, hook *test.Hook, prefix string) {
	entry := hook.LastEntry()
	if len(prefix) == 0 && entry == nil {
		return
	} else if len(prefix) == 0 && entry != nil {
		t.Errorf("Expected no logged message. But found an entry %v", entry)
		return
	} else if entry == nil {
		t.Errorf("Entry is nil")
		return
	}
	assert.True(t, strings.HasPrefix(entry.Message, prefix), "Should have prefix: '%s'; got [%s] '%s'", prefix, entry.Level, entry.Message)
}

// CheckLogEntries checks the latest logged entry if the message begin with the given prefix.
// If the prefix is "" then it expects that no log messages were written.
func CheckLogEntries(t *testing.T, hook *test.Hook, message string) {
	entries := hook.AllEntries()
	if len(message) == 0 && len(entries) == 0 {
		return
	} else if len(message) == 0 && len(entries) != 0 {
		t.Errorf("Expected no logged message. But found an entry %v", entries)
		return
	}
	var messages []string
	for _, e := range entries {
		messages = append(messages, e.Message)
	}

	assert.Contains(t, messages, message)
}
