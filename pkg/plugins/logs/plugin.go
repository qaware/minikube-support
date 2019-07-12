package logs

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"strings"
)

// plugin is the internal structure for the logs plugin.
type plugin struct {
	msgChannel chan *apis.MonitoringMessage
	buffer     *buffer
	writer     *writer
	logger     *logrus.Logger
	end        chan bool
	render     chan bool
}

const pluginName = "logs"

var formatter = &logrus.TextFormatter{
	ForceColors:   true,
	FullTimestamp: true,
}

// NewLogsPlugin initializes the logs plugin for the run view.
func NewLogsPlugin(logger *logrus.Logger) apis.StartStopPlugin {
	buffer := newBuffer()
	renderTrigger := make(chan bool, 10)
	hook := &plugin{
		end:    make(chan bool),
		render: renderTrigger,
		writer: &writer{buffer: buffer, renderTrigger: renderTrigger},
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
	go l.messageRenderer()
	return pluginName, nil
}

// messageRenderer produces the the current output for the logs plugin view
func (l *plugin) messageRenderer() {
	for {
		select {
		case <-l.render:
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
		case <-l.end:
			return
		}
	}
}

func (l *plugin) Stop() error {
	l.end <- true
	return nil
}
