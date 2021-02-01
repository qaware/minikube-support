package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/plugins"
)

func TestUninstallOptions_Run(t *testing.T) {
	// Todo: Improve test to ensure the reverse call order compared to install
	tests := []struct {
		name                string
		phase               apis.Phase
		includeLocalPlugins bool
		uninstalled         bool
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
			i := UninstallOptions{
				registry:            registry,
				includeLocalPlugins: tt.includeLocalPlugins,
			}

			i.Run(&cobra.Command{}, []string{})

			assert.False(t, plugin.installRun)
			assert.False(t, plugin.updateRun)
			assert.Equal(t, tt.uninstalled, plugin.uninstallRun)
		})
	}
}
