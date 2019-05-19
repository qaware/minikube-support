package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/spf13/cobra"
)

func TestUninstallOptions_Run(t *testing.T) {
	plugin := &DummyPlugin{}
	i := UninstallOptions{
		plugins: []apis.InstallablePlugin{plugin},
	}
	i.Run(&cobra.Command{}, []string{})
	assert.False(t, plugin.installRun)
	assert.False(t, plugin.updateRun)
	assert.True(t, plugin.uninstallRun)
}
