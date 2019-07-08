package plugins

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"reflect"
	"testing"
)

func (p *DummyPlugin) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	return p.String(), nil
}

func (p *DummyPlugin) Stop() error {
	return nil
}

func (c *DummyPlugin) IsSingleRunnable() bool {
	return false
}

func Test_startStopPluginRegistry_AddPlugin(t *testing.T) {
	tests := []struct {
		name        string
		plugin      apis.StartStopPlugin
		plugins     map[string]apis.StartStopPlugin
		shouldPanic bool
	}{
		{"ok", &DummyPlugin{}, map[string]apis.StartStopPlugin{}, false},
		{"nil plugin", nil, map[string]apis.StartStopPlugin{}, true},
		{"twice", &DummyPlugin{}, map[string]apis.StartStopPlugin{"dummy": &DummyPlugin{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.shouldPanic {
					t.Errorf("Recovered panic is %v; shouldPanic=%v", r, tt.shouldPanic)
				}
			}()

			r := &startStopPluginRegistry{
				plugins: tt.plugins,
			}
			r.AddPlugin(tt.plugin)
		})
	}
}

func Test_startStopPluginRegistry_FindPlugin(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		plugins    map[string]apis.StartStopPlugin
		want       apis.StartStopPlugin
		wantErr    bool
	}{
		{"ok", "dummy", map[string]apis.StartStopPlugin{"dummy": &DummyPlugin{}}, &DummyPlugin{}, false},
		{"not found", "dummy", map[string]apis.StartStopPlugin{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &startStopPluginRegistry{
				plugins: tt.plugins,
			}
			got, err := r.FindPlugin(tt.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("startStopPluginRegistry.FindPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("startStopPluginRegistry.FindPlugin() = %v, want %v", got, tt.want)
			}
		})
	}
}
