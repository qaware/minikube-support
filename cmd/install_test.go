package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallOptions_Run(t *testing.T) {
	plugin := &DummyPlugin{}
	i := InstallOptions{
		plugins: []apis.InstallablePlugin{plugin},
	}
	i.Run(&cobra.Command{}, []string{})
	assert.True(t, plugin.installRun)
	assert.False(t, plugin.updateRun)
	assert.False(t, plugin.uninstallRun)
}
