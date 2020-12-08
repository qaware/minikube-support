package cmd

import (
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/qaware/minikube-support/pkg/testutils"

	"github.com/spf13/cobra"

	"github.com/qaware/minikube-support/pkg/apis"
)

func TestRunSingleOptions_Run(t *testing.T) {
	hook := test.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)
	tests := []struct {
		name          string
		plugin        *DummyPlugin
		startupPrefix string
		stopPrefix    string
	}{
		{
			"ok",
			&DummyPlugin{
				started: make(chan bool),
				run: func(messages chan *apis.MonitoringMessage) {
					messages <- &apis.MonitoringMessage{Box: "", Message: "message"}
				},
			},
			"New dummy status:\nmessage",
			"Received signal ",
		},
		{
			"fail start",
			&DummyPlugin{
				failStart: true,
				started:   make(chan bool),
			},
			"Can not start plugin dummy: fail",
			"Can not start plugin dummy: fail",
		},
		{
			"fail stop",
			&DummyPlugin{
				failStop: true,
				started:  make(chan bool),
				run:      func(messages chan *apis.MonitoringMessage) {},
			},
			"New dummy status:\nStarting...",
			"Unable to terminate plugin dummy:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer hook.Reset()
			i := NewRunSingleOptions(tt.plugin)

			terminated := make(chan bool)
			go func() {
				i.Run(&cobra.Command{}, []string{})
				terminated <- true
			}()
			<-tt.plugin.started
			time.Sleep(50 * time.Millisecond)
			testutils.CheckLogEntry(t, hook, tt.startupPrefix)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			<-terminated
			testutils.CheckLogEntry(t, hook, tt.stopPrefix)
		})
	}
}
