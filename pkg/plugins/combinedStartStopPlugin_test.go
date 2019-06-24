package plugins

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
)

func TestNewCombinedPlugin(t *testing.T) {
	tests := []struct {
		name    string
		plugins []apis.StartStopPlugin
		want    *CombinedStartStopPlugin
		wantErr bool
	}{
		{"ok", []apis.StartStopPlugin{&DummyPlugin{}, &DummyPlugin{}}, &CombinedStartStopPlugin{"test", []apis.StartStopPlugin{&DummyPlugin{}, &DummyPlugin{}}, true}, false},
		{"too few plugins", []apis.StartStopPlugin{&DummyPlugin{}}, nil, true},
		{"no plugins", []apis.StartStopPlugin{}, nil, true},
		{"nil plugins", nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCombinedPlugin("test", tt.plugins, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCombinedPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCombinedPlugin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedStartStopPlugin_Start(t *testing.T) {
	tests := []struct {
		name    string
		plugins []apis.StartStopPlugin
		want    string
		wantErr bool
	}{
		{"ok", []apis.StartStopPlugin{&DummyPlugin{}, &DummyPlugin{}}, "test", false},
		{"one fails", []apis.StartStopPlugin{&DummyPlugin{}, &failingStartStopPlugin{}}, "test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CombinedStartStopPlugin{
				pluginName:     "test",
				plugins:        tt.plugins,
				singleRunnable: true,
			}
			got, err := c.Start(make(chan *apis.MonitoringMessage))
			if (err != nil) != tt.wantErr {
				t.Errorf("CombinedStartStopPlugin.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CombinedStartStopPlugin.Start() = %v, want %v", got, tt.want)
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
