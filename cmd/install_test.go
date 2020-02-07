package cmd

import (
	"testing"

	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/stretchr/testify/assert"

	"github.com/spf13/cobra"
)

func TestInstallOptions_Run(t *testing.T) {
	plugin := &DummyPlugin{}
	registry := plugins.NewInstallablePluginRegistry()
	registry.AddPlugin(plugin)
	i := InstallOptions{
		registry: registry,
	}
	i.Run(&cobra.Command{}, []string{})
	assert.True(t, plugin.installRun)
	assert.False(t, plugin.updateRun)
	assert.False(t, plugin.uninstallRun)
}
