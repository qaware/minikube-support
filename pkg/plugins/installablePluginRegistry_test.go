package plugins

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
)

type DummyPlugin struct {
	executedFunction string
	phase            apis.Phase
	name             string
}

func (p *DummyPlugin) String() string {
	if p.name == "" {
		return "dummy"
	}
	return p.name
}

func (p *DummyPlugin) Install() {
	p.executedFunction = "install"
}

func (p *DummyPlugin) Update() {
	p.executedFunction = "update"
}

func (p *DummyPlugin) Uninstall(purge bool) {
	p.executedFunction = fmt.Sprintf("uninstall %v", purge)
}

func (p *DummyPlugin) Phase() apis.Phase {
	return p.phase
}

func Test_installablePluginRegistry_AddPlugin(t *testing.T) {
	tests := []struct {
		name        string
		plugin      apis.InstallablePlugin
		plugins     map[string]apis.InstallablePlugin
		shouldPanic bool
	}{
		{"ok", &DummyPlugin{}, map[string]apis.InstallablePlugin{}, false},
		{"nil plugin", nil, map[string]apis.InstallablePlugin{}, true},
		{"twice", &DummyPlugin{}, map[string]apis.InstallablePlugin{"dummy": &DummyPlugin{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.shouldPanic {
					t.Errorf("Recovered panic is %v; shouldPanic=%v", r, tt.shouldPanic)
				}
			}()

			r := &installablePluginRegistry{
				plugins: tt.plugins,
			}
			r.AddPlugin(tt.plugin)
		})
	}
}

func Test_installablePluginRegistry_FindPlugin(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		plugins    map[string]apis.InstallablePlugin
		want       apis.InstallablePlugin
		wantErr    bool
	}{
		{"ok", "dummy", map[string]apis.InstallablePlugin{"dummy": &DummyPlugin{}}, &DummyPlugin{}, false},
		{"not found", "dummy", map[string]apis.InstallablePlugin{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &installablePluginRegistry{
				plugins: tt.plugins,
			}
			got, err := r.FindPlugin(tt.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("installablePluginRegistry.FindPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("installablePluginRegistry.FindPlugin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_installablePluginRegistry_ListPlugins(t *testing.T) {
	tests := []struct {
		name    string
		plugins map[string]apis.InstallablePlugin
		want    apis.InstallablePluginList
	}{
		{"one plugin", map[string]apis.InstallablePlugin{"dummy1": &DummyPlugin{name: "dummy1", phase: 1}}, []apis.InstallablePlugin{&DummyPlugin{name: "dummy1", phase: 1}}},
		{"two different phase", map[string]apis.InstallablePlugin{"dummy1": &DummyPlugin{name: "dummy1", phase: 2}, "dummy2": &DummyPlugin{phase: 1}}, []apis.InstallablePlugin{&DummyPlugin{phase: 1}, &DummyPlugin{name: "dummy1", phase: 2}}},
		{"three different phase", map[string]apis.InstallablePlugin{"dummy1": &DummyPlugin{name: "dummy1", phase: 2}, "dummy2": &DummyPlugin{phase: 1}, "dummy3": &DummyPlugin{phase: 1}}, []apis.InstallablePlugin{&DummyPlugin{phase: 1}, &DummyPlugin{phase: 1}, &DummyPlugin{name: "dummy1", phase: 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &installablePluginRegistry{
				plugins: tt.plugins,
			}
			if got := r.ListPlugins(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("installablePluginRegistry.ListPlugins() = %v, want %v", got, tt.want)
			}
		})
	}
}
