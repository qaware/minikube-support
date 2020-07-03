package cmd

import (
	"fmt"
	"os/exec"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunOptions_Run(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name          string
		plugins       []apis.StartStopPlugin
		activePlugins []string
		lastMessages  []apis.MonitoringMessage
	}{
		{
			"all ok",
			[]apis.StartStopPlugin{&DummyPlugin{}},
			[]string{"dummy"},
			[]apis.MonitoringMessage{{Box: "dummy", Message: "Starting..."}},
		}, {
			"one start fails",
			[]apis.StartStopPlugin{&DummyPlugin{failStart: true},
				&DummyPlugin{name: "dummy1"}},
			[]string{"dummy1"},
			[]apis.MonitoringMessage{{Box: "dummy1", Message: "Starting..."}},
		}, {
			"one start fails",
			[]apis.StartStopPlugin{&DummyPlugin{failStop: true}},
			[]string{"dummy"},
			[]apis.MonitoringMessage{{Box: "dummy", Message: "Starting..."}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.TestProcessResponses = []testutils.TestProcessResponse{{Command: "sudo", Args: []string{"echo"}, ResponseStatus: 0}}
			options := &RunOptions{
				plugins:        tt.plugins,
				messageChannel: make(chan *apis.MonitoringMessage),
				lastMessages:   map[string]*apis.MonitoringMessage{},
				contextName:    func() string { return "" },
			}
			go options.Run(&cobra.Command{}, []string{})
			time.Sleep(200 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)

			assert.Equal(t, tt.activePlugins, options.activePlugins)
			messages := messagesValues(options.lastMessages)
			assert.Equal(t, tt.lastMessages, messages)
		})
	}
}

func TestRunOptions_startPlugins(t *testing.T) {
	tests := []struct {
		name          string
		plugins       []apis.StartStopPlugin
		activePlugins []string
	}{
		{"all ok", []apis.StartStopPlugin{&DummyPlugin{}}, []string{"dummy"}},
		{"one fails", []apis.StartStopPlugin{&DummyPlugin{failStart: true}, &DummyPlugin{name: "dummy1"}}, []string{"dummy1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &RunOptions{
				plugins:        tt.plugins,
				messageChannel: make(chan *apis.MonitoringMessage),
			}
			go options.startPlugins()
			i := 0
			for range options.messageChannel {
				i++
				if i == len(tt.activePlugins) {
					break
				}
			}
			assert.Equal(t, tt.activePlugins, options.activePlugins)
		})
	}
}

func Test_printHeader(t *testing.T) {
	terminalWidth = func() int { return 30 }
	var printed string
	terminalPrint = func(a ...interface{}) (n int, err error) {
		printed = fmt.Sprint(a...)
		return len(printed), nil
	}
	tests := []struct {
		name       string
		k8sContext string
		wantErr    bool
	}{
		{"", "context", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := printHeader(tt.k8sContext); (err != nil) != tt.wantErr {
				t.Errorf("printHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
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
