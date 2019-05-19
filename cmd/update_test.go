package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/spf13/cobra"
)

func TestUpdateOptions_Run(t *testing.T) {
	plugin := &DummyPlugin{}
	i := UpdateOptions{
		plugins: []apis.InstallablePlugin{plugin},
	}
	i.Run(&cobra.Command{}, []string{})
	assert.False(t, plugin.installRun)
	assert.True(t, plugin.updateRun)
	assert.False(t, plugin.uninstallRun)
}
