package cmd

import (
	"os/exec"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/spf13/cobra"

	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"

	"github.com/stretchr/testify/assert"

	"github.com/qaware/minikube-support/pkg/apis"
)

func TestRunOptions_Run(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	newGui = func(mode gocui.OutputMode, supportOverlaps bool) (*gocui.Gui, error) {
		return gocui.NewGui(gocui.OutputSimulator, supportOverlaps)
	}
	defer func() {
		sh.ExecCommand = exec.Command
		newGui = gocui.NewGui
	}()

	tests := []struct {
		name          string
		plugins       []apis.StartStopPlugin
		activePlugins []string
		lastMessages  []apis.MonitoringMessage
	}{
		{
			"all ok",
			[]apis.StartStopPlugin{
				&DummyPlugin{},
			},
			[]string{"dummy"},
			[]apis.MonitoringMessage{{Box: "dummy", Message: "Starting..."}},
		},
		{
			"one start fails",
			[]apis.StartStopPlugin{
				&DummyPlugin{failStart: true},
				&DummyPlugin{name: "dummy1"},
			},
			[]string{"dummy1"},
			[]apis.MonitoringMessage{{Box: "dummy1", Message: "Starting..."}},
		},
		{
			"one stop fails",
			[]apis.StartStopPlugin{
				&DummyPlugin{failStop: true},
			},
			[]string{"dummy"},
			[]apis.MonitoringMessage{{Box: "dummy", Message: "Starting..."}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.SetTestProcessResponses([]testutils.TestProcessResponse{
				{Command: "sudo", Args: []string{"echo", ""}, ResponseStatus: 0},
				{Command: "which", Args: []string{"sudo"}, ResponseStatus: 0},
			})
			startedChannel, cntPlugins := injectStartedChannel(tt.plugins)
			options := &RunOptions{
				plugins:          tt.plugins,
				messageChannel:   make(chan *apis.MonitoringMessage),
				lastMessages:     map[string]*apis.MonitoringMessage{},
				contextName:      func() string { return "" },
				lastMessagesLock: sync.RWMutex{},
			}
			terminated := make(chan bool)
			go func() {
				options.Run(&cobra.Command{}, []string{})
				terminated <- true
			}()
			waitForStarted(startedChannel, cntPlugins)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)

			select {
			case <-terminated:
				options.lastMessagesLock.RLock()
				messages := messagesValues(options.lastMessages)
				options.lastMessagesLock.RUnlock()
				assert.Equal(t, tt.lastMessages, messages)

			case <-time.After(1 * time.Second):
				assert.Fail(t, "terminated message not received")
			}
		})
	}
}

func TestRunOptions_startPlugins(t *testing.T) {
	tests := []struct {
		name          string
		plugins       []apis.StartStopPlugin
		activePlugins []string
	}{
		{
			"all ok",
			[]apis.StartStopPlugin{
				&DummyPlugin{},
			},
			[]string{"dummy"},
		},
		{
			"one fails",
			[]apis.StartStopPlugin{
				&DummyPlugin{failStart: true},
				&DummyPlugin{name: "dummy1"},
			},
			[]string{"dummy1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &RunOptions{
				plugins:        tt.plugins,
				messageChannel: make(chan *apis.MonitoringMessage),
			}
			go func() {
				options.startPlugins()
				close(options.messageChannel)
			}()
			var actualActivePlugins []string
			for message := range options.messageChannel {
				actualActivePlugins = append(actualActivePlugins, message.Box)
			}
			assert.Equal(t, tt.activePlugins, actualActivePlugins)
		})
	}
}

func Test_createHeader(t *testing.T) {
	tests := []struct {
		name       string
		k8sContext string
		wantErr    bool
	}{
		{"", "context", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printed := createHeader(tt.k8sContext, 30)
			assert.Equal(t, "Kubernetes Kontext: context "+time.Now().Format(time.UnixDate), printed)
		})
	}
}

func messagesValues(m map[string]*apis.MonitoringMessage) []apis.MonitoringMessage {
	var result []apis.MonitoringMessage
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

func Test_calcBoxSize(t *testing.T) {
	tests := []struct {
		name      string
		available int
		numBoxes  int
		want      []int
	}{
		{"1", 100, 1, []int{100}},
		{"2", 100, 2, []int{50, 50}},
		{"3", 100, 3, []int{34, 33, 33}},
		{"4", 100, 4, []int{25, 25, 25, 25}},
		{"5", 100, 5, []int{20, 20, 20, 20, 20}},
		{"6", 100, 6, []int{17, 17, 17, 17, 16, 16}},
		{"7", 100, 7, []int{15, 15, 14, 14, 14, 14, 14}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcBoxSize(tt.available, tt.numBoxes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calcBoxSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}

// injectStartedChannel is a helper function which injects into all DummyPlugins a channel that indicates that the
// plugin was started. It will return the injected channel and the number of plugins in which the channel was injected.
func injectStartedChannel(plugins []apis.StartStopPlugin) (chan bool, int) {
	cnt := 0
	started := make(chan bool)
	for _, o := range plugins {
		p, ok := o.(*DummyPlugin)
		if ok {
			p.started = started
			cnt++
		}
	}
	return started, cnt
}

// waitForStarted is a small helper which waits until the same number of messages on the channel were received as the
// defined in countPlugins. It is built to work together with the injectStartChannel() function
func waitForStarted(startedChannel chan bool, countPlugins int) {
	for range startedChannel {
		countPlugins--
		if countPlugins == 0 {
			return
		}
	}
}
