package cmd

import (
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUninstallOptions_Run(t *testing.T) {
	// Todo: Improve test to ensure the reverse call order compared to install
	plugin := &DummyPlugin{}
	registry := plugins.NewInstallablePluginRegistry()
	registry.AddPlugin(plugin)
	i := UninstallOptions{
		registry: registry,
	}
	i.Run(&cobra.Command{}, []string{})
	assert.False(t, plugin.installRun)
	assert.False(t, plugin.updateRun)
	assert.True(t, plugin.uninstallRun)
}
