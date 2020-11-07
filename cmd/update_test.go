package cmd

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"testing"

	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/stretchr/testify/assert"

	"github.com/spf13/cobra"
)

func TestUpdateOptions_Run(t *testing.T) {
	tests := []struct {
		name                string
		phase               apis.Phase
		includeLocalPlugins bool
		updated             bool
	}{
		{"cluster plugin without local", apis.CLUSTER_TOOLS_INSTALL, false, true},
		{"cluster plugin with local", apis.CLUSTER_TOOLS_INSTALL, true, true},
		{"local plugin without local", apis.LOCAL_TOOLS_INSTALL, false, false},
		{"local plugin with local", apis.LOCAL_TOOLS_INSTALL, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &DummyPlugin{
				phase: tt.phase,
			}
			registry := plugins.NewInstallablePluginRegistry()
			registry.AddPlugin(plugin)
			i := UpdateOptions{
				registry:            registry,
				includeLocalPlugins: tt.includeLocalPlugins,
			}

			i.Run(&cobra.Command{}, []string{})

			assert.False(t, plugin.installRun)
			assert.Equal(t, tt.updated, plugin.updateRun)
			assert.False(t, plugin.uninstallRun)
		})
	}
}
