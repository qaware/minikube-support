package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalPlugin(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
		want  bool
	}{
		{"LOCAL_TOOLS_INSTALL", LOCAL_TOOLS_INSTALL, true},
		{"LOCAL_TOOLS_CONFIG", LOCAL_TOOLS_CONFIG, true},
		{"CLUSTER_INIT", CLUSTER_INIT, false},
		{"CLUSTER_CONFIG", CLUSTER_CONFIG, false},
		{"CLUSTER_TOOLS_INSTALL", CLUSTER_TOOLS_INSTALL, false},
		{"CLUSTER_TOOLS_CONFIG", CLUSTER_TOOLS_CONFIG, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLocalPlugin(&DummyPlugin{phase: tt.phase})
			assert.Equal(t, tt.want, got)
		})
	}
}

type DummyPlugin struct {
	phase Phase
}

func (d DummyPlugin) String() string {
	panic("implement me")
}

func (d DummyPlugin) Install() {
	panic("implement me")
}

func (d DummyPlugin) Update() {
	panic("implement me")
}

func (d DummyPlugin) Uninstall(_ bool) {
	panic("implement me")
}

func (d DummyPlugin) Phase() Phase {
	return d.phase
}
