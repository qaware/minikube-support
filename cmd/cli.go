package cmd

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/spf13/cobra"
	"os"
)

// RootCommandOptions stores the the values for global command flags like the kubeconfig and context name.
type RootCommandOptions struct {
	kubeConfig                string
	contextName               string
	installablePluginRegistry apis.InstallablePluginRegistry
	startStopPluginRegistry   apis.StartStopPluginRegistry
}

var rootCmd *cobra.Command

// NewRootCmd initializes the root cobra command including all the flags on root level.
func NewRootCmd() (*cobra.Command, *RootCommandOptions) {
	cmd := &cobra.Command{
		Use:   "minikube-support",
		Short: "The minikube-support tools helps to integrate minikube better into your local os.",
	}

	options := NewRootCmdOptions()

	flags := cmd.PersistentFlags()
	flags.StringVar(&options.kubeConfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")
	flags.StringVar(&options.contextName, "context", "", "The name of the kubeconfig context to use")

	return cmd, options
}

// NewRootCmdOptions creates a new option structure for the the root cli command.
func NewRootCmdOptions() *RootCommandOptions {
	options := &RootCommandOptions{
		installablePluginRegistry: plugins.NewInstallablePluginRegistry(),
		startStopPluginRegistry:   plugins.NewStartStopPluginRegistry(),
	}
	return options
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// initialize basic commands
	var options *RootCommandOptions
	rootCmd, options = NewRootCmd()
	rootCmd.AddCommand(NewVersionCommand(), NewCompletionCmd())

	// initializes plugins
	initPlugins(options)

	// initialize install, update and uninstall
	rootCmd.AddCommand(
		NewInstallCommand(options.installablePluginRegistry),
		NewUpdateCommand(options.installablePluginRegistry),
		NewUninstallCommand(options.installablePluginRegistry))

	// initializes run commands
	runCmd := NewRunCommand(options.startStopPluginRegistry)
	rootCmd.AddCommand(runCmd)
	for _, plugin := range options.startStopPluginRegistry.ListPlugins() {
		if plugin.IsSingleRunnable() {
			runCmd.AddCommand(NewRunSingleCommand(plugin))
		}
	}
}
