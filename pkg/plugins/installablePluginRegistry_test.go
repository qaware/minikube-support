package plugins

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateInstallCommands(t *testing.T) {
	plugin := initTestRegistry()
	plugin.checkCommand(t, CreateInstallCommands(), "Installs the dummy plugin.", "install")
}

func (p *DummyPlugin) checkCommand(t *testing.T, cmds []*cobra.Command, short string, function string) {
	if len(cmds) != 1 {
		t.Errorf("want one command. %v given", len(cmds))
	}
	cmd := cmds[0]
	assert.Equal(t, short, cmd.Short)
	cmd.Run(cmd, []string{})
	assert.Equal(t, function, p.executedFunction)
}

func TestCreateUpdateCommands(t *testing.T) {
	plugin := initTestRegistry()
	plugin.checkCommand(t, CreateUpdateCommands(), "Updates the dummy plugin.", "update")
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

			plugin := initTestRegistry()
			commands := CreateUninstallCommands()
			tt.flag(commands[0].Flags())
			plugin.checkCommand(t, commands, "Uninstall the dummy plugin.", fmt.Sprint("uninstall ", tt.purge))
		})
	}
}

func TestGetInstallablePlugins(t *testing.T) {
	initTestRegistry()
	if got := GetInstallablePlugins(); len(got) != 1 {
		t.Errorf("len(GetInstallablePlugins()) = %v, want %v", got, 1)
	}
}

func initTestRegistry() *DummyPlugin {
	installPlugins = newInstallablePluginRegistry()
	plugin := &DummyPlugin{}
	installPlugins.addPlugins(plugin)
	return plugin
}

type DummyPlugin struct {
	executedFunction string
}

func (p *DummyPlugin) String() string {
	return "dummy"
}

func (p *DummyPlugin) Install() {
	p.executedFunction = "install"
}

func (p *DummyPlugin) Update() {
	p.executedFunction = "update"
}

func (p *DummyPlugin) Uninstall(purge bool) {
	p.executedFunction = fmt.Sprintf("uninstall %v", purge)
}
