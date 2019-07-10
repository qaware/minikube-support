package logs

import (
	"fmt"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

type logHook struct {
	entries    *buffer
	mu         sync.RWMutex
	msgChannel chan *apis.MonitoringMessage
	end        chan bool
	end1       chan bool

	entryChan chan *logrus.Entry
}

const pluginName = "logs"

var formatter = logrus.TextFormatter{
	ForceColors:   false,
	FullTimestamp: true,
}

func NewLogHook() apis.StartStopPlugin {
	hook := &logHook{end: make(chan bool), end1: make(chan bool), entryChan: make(chan *logrus.Entry), entries: newBuffer()}
	logrus.AddHook(hook)
	go hook.handleEntries()
	return hook
}

func (*logHook) String() string {
	return pluginName
}

func (*logHook) IsSingleRunnable() bool {
	return false
}

func (l *logHook) Start(messageChannel chan *apis.MonitoringMessage) (boxName string, err error) {
	l.msgChannel = messageChannel
	go utils.Ticker(l.messageTicker, l.end1, 1*time.Second)
	return pluginName, nil
}

func (l *logHook) messageTicker() {
	payloads := l.entries.GetEntries()

	message := ""
	for i := len(payloads) - 1; i >= 0; i-- {
		if payloads[i] == nil {
			continue
		}
		message += payloads[i].(string) + "\n"
	}
	l.msgChannel <- &apis.MonitoringMessage{Box: pluginName, Message: message}
}

func (l *logHook) Stop() error {
	l.end <- true
	l.end1 <- true
	return nil
}

func (*logHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *logHook) Fire(entry *logrus.Entry) error {
	l.entryChan <- entry
	return nil
}

func (l *logHook) handleEntries() {
	for {
		select {
		case entry := <-l.entryChan:
			entry.Buffer = nil
			bytes, e := formatter.Format(entry)
			if e != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Unable to write log message: %s", e)
				continue
			}
			message := strings.Trim(string(bytes), "\r\n")
			l.entries.Write(message)

		case <-l.end:
			return
		}
	}
}
