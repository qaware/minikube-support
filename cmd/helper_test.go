package cmd

import (
	"testing"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstallCommands(t *testing.T) {
	plugin, registry := initTestRegistry(apis.LOCAL_TOOLS_INSTALL)
	plugin.checkCommand(t, createInstallCommands(registry), "Installs the dummy local plugin.", true, false, false)
}

func (p *DummyPlugin) checkCommand(t *testing.T, cmds []*cobra.Command, short string, installCalled bool, updateCalled bool, uninstallCalled bool) {
	if len(cmds) != 1 {
		t.Errorf("want one command. %v given", len(cmds))
	}
	cmd := cmds[0]
	assert.Equal(t, short, cmd.Short)
	cmd.Run(cmd, []string{})
	assert.Equal(t, installCalled, p.installRun)
	assert.Equal(t, updateCalled, p.updateRun)
	assert.Equal(t, uninstallCalled, p.uninstallRun)
}

func TestCreateUpdateCommands(t *testing.T) {
	plugin, registry := initTestRegistry(apis.LOCAL_TOOLS_INSTALL)
	plugin.checkCommand(t, CreateUpdateCommands(registry), "Updates the dummy local plugin.", false, true, false)
}

func TestCreateUninstallCommands(t *testing.T) {
	tests := []struct {
		name      string
		flag      func(set *pflag.FlagSet)
		purge     bool
		wantPanic bool
	}{
		{"missing flag", func(set *pflag.FlagSet) {}, true, true},
		{"purge", func(set *pflag.FlagSet) { set.Bool("purge", true, "") }, true, false},
		{"no purge", func(set *pflag.FlagSet) { set.Bool("purge", false, "") }, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("Not expected panic. Want %v got %v", tt.wantPanic, r)
				}
			}()

			plugin, registry := initTestRegistry(apis.CLUSTER_TOOLS_INSTALL)
			commands := createUninstallCommands(registry)
			tt.flag(commands[0].Flags())
			plugin.checkCommand(t, commands, "Uninstall the dummy cluster plugin.", false, false, true)
			assert.Equal(t, tt.purge, plugin.purge)
		})
	}
}

func TestGetInstallablePlugins(t *testing.T) {
	_, registry := initTestRegistry(apis.CLUSTER_TOOLS_INSTALL)
	if got := registry.ListPlugins(); len(got) != 1 {
		t.Errorf("len(GetInstallablePlugins()) = %v, want %v", got, 1)
	}
}

func initTestRegistry(phase apis.Phase) (*DummyPlugin, apis.InstallablePluginRegistry) {
	installPlugins := plugins.NewInstallablePluginRegistry()
	plugin := &DummyPlugin{phase: phase}
	installPlugins.AddPlugins(plugin)
	return plugin, installPlugins
}
