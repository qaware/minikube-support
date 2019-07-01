package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/spf13/cobra"
)

func TestUpdateOptions_Run(t *testing.T) {
	plugin := &DummyPlugin{}
	registry := plugins.NewInstallablePluginRegistry()
	registry.AddPlugin(plugin)
	i := UpdateOptions{
		registry: registry,
	}
	i.Run(&cobra.Command{}, []string{})
	assert.False(t, plugin.installRun)
	assert.True(t, plugin.updateRun)
	assert.False(t, plugin.uninstallRun)
}
