package plugins

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
)

func TestCombinedStartStopPlugin_Start(t *testing.T) {
	tests := []struct {
		name        string
		combineFunc CombineFunc
		wantPlugins []string
		wantErr     bool
	}{
		{
			"ok",
			func() ([]apis.StartStopPlugin, error) {
				return []apis.StartStopPlugin{&DummyPlugin{}}, nil
			},
			[]string{"dummy"},
			false,
		},
		{
			"combine func nil",
			nil,
			[]string{},
			true,
		},
		{
			"one ok, other fails",
			func() ([]apis.StartStopPlugin, error) {
				return []apis.StartStopPlugin{&DummyPlugin{}, &failingStartStopPlugin{}}, nil
			},
			[]string{"dummy"},
			false,
		},
		{
			"combine func error",
			func() ([]apis.StartStopPlugin, error) {
				return nil, fmt.Errorf("fail")
			},
			[]string{"dummy"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCombinedPlugin("t", tt.combineFunc, true)
			_, err := c.Start(make(chan *apis.MonitoringMessage))
			if (err != nil) != tt.wantErr {
				t.Errorf("CombinedStartStopPlugin.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			var plugins []string
			for _, p := range c.(*CombinedStartStopPlugin).plugins {
				plugins = append(plugins, p.String())
			}
			if !reflect.DeepEqual(plugins, tt.wantPlugins) {
				t.Errorf("CombinedStartStopPlugin.Start() = %v, want %v", plugins, tt.wantPlugins)
			}
		})
	}
}

func TestCombinedStartStopPlugin_Stop(t *testing.T) {
	tests := []struct {
		name    string
		plugins []apis.StartStopPlugin
		wantErr bool
	}{
		{"ok", []apis.StartStopPlugin{&DummyPlugin{}, &failingStartStopPlugin{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CombinedStartStopPlugin{
				plugins: tt.plugins,
			}
			if err := c.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("CombinedStartStopPlugin.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type failingStartStopPlugin struct{}

func (failingStartStopPlugin) String() string {
	return "failingStartStopPlugin"
}

func (failingStartStopPlugin) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	return "", fmt.Errorf("fail")
}

func (failingStartStopPlugin) Stop() error {
	return fmt.Errorf("fail")
}

func (failingStartStopPlugin) IsSingleRunnable() bool {
	return false
}
