package cmd

import (
	"syscall"
	"testing"
	"time"

	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/spf13/cobra"
)

func TestRunSingleOptions_Run(t *testing.T) {
	hook := test.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)
	tests := []struct {
		name          string
		plugin        apis.StartStopPlugin
		startupPrefix string
		stopPrefix    string
	}{
		{"ok", &DummyPlugin{run: func(messages chan *apis.MonitoringMessage) {
			messages <- &apis.MonitoringMessage{Box: "", Message: "message"}
		}}, "New dummy status:\nmessage", "Received signal "},
		{"fail start", &DummyPlugin{failStart: true}, "Can not start plugin dummy: fail", "Can not start plugin dummy: fail"},
		{"fail stop", &DummyPlugin{failStop: true, run: func(messages chan *apis.MonitoringMessage) {}}, "Plugin dummy successfully started.", "Unable to terminate plugin dummy:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer hook.Reset()
			i := NewRunSingleOptions(tt.plugin)

			go i.Run(&cobra.Command{}, []string{})
			time.Sleep(50 * time.Millisecond)
			testutils.CheckLogEntry(t, hook, tt.startupPrefix)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			time.Sleep(50 * time.Millisecond)
			testutils.CheckLogEntry(t, hook, tt.stopPrefix)
		})
	}
}
