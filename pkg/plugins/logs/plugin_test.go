package logs

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
)

func Test_logHook(t *testing.T) {
	logger := logrus.New()
	l := NewLogsPlugin(logger)
	msgChannel := make(chan *apis.MonitoringMessage)
	_, _ = l.Start(msgChannel)

	logger.Info("Test1")
	msg := <-msgChannel
	assert.True(t, strings.Contains(msg.Message, "Test1"))

	logger.Info("Test2")
	msg = <-msgChannel
	assert.True(t, strings.Contains(msg.Message, "Test1") && strings.Contains(msg.Message, "Test2"))

	_ = l.Stop()
}
