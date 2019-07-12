package logs

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/utils"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

// plugin is the internal structure for the logs plugin.
type plugin struct {
	msgChannel chan *apis.MonitoringMessage
	buffer     *buffer
	writer     *writer
	logger     *logrus.Logger
	end        chan bool
}

const pluginName = "logs"

var formatter = &logrus.TextFormatter{
	ForceColors:   true,
	FullTimestamp: true,
}

// NewLogsPlugin initializes the logs plugin for the run view.
func NewLogsPlugin(logger *logrus.Logger) apis.StartStopPlugin {
	buffer := newBuffer()
	hook := &plugin{
		end:    make(chan bool),
		writer: &writer{buffer: buffer},
		buffer: buffer,
		logger: logger,
	}

	return hook
}

func (*plugin) String() string {
	return pluginName
}

func (*plugin) IsSingleRunnable() bool {
	return false
}

func (l *plugin) Start(messageChannel chan *apis.MonitoringMessage) (boxName string, err error) {
	l.logger.SetOutput(l.writer)
	l.logger.SetFormatter(formatter)
	l.msgChannel = messageChannel
	go utils.Ticker(l.messageTicker, l.end, 500*time.Microsecond)
	return pluginName, nil
}

// messageTicker produces the the current output for the logs plugin view
func (l *plugin) messageTicker() {
	payloads := l.buffer.GetEntries()

	message := ""
	outputPadding := "        "
	for i := len(payloads) - 1; i >= 0; i-- {
		if payloads[i] == nil {
			continue
		}

		payload := payloads[i].(string)
		payload = strings.Trim(payload, "\r\n")
		message += payload + outputPadding + "\n"
	}
	l.msgChannel <- &apis.MonitoringMessage{Box: pluginName, Message: message}
}

func (l *plugin) Stop() error {
	l.end <- true
	return nil
}
