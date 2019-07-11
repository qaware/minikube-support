package logs

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/utils"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type logHook struct {
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

func NewLogHook(logger *logrus.Logger) apis.StartStopPlugin {
	buffer := newBuffer()
	hook := &logHook{
		end:    make(chan bool),
		writer: &writer{buffer: buffer},
		buffer: buffer,
		logger: logger,
	}

	return hook
}

func (*logHook) String() string {
	return pluginName
}

func (*logHook) IsSingleRunnable() bool {
	return false
}

func (l *logHook) Start(messageChannel chan *apis.MonitoringMessage) (boxName string, err error) {
	l.logger.SetOutput(l.writer)
	l.logger.SetFormatter(formatter)
	l.msgChannel = messageChannel
	go utils.Ticker(l.messageTicker, l.end, 1*time.Second)
	return pluginName, nil
}

func (l *logHook) messageTicker() {
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

func (l *logHook) Stop() error {
	l.end <- true
	return nil
}
