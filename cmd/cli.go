package cmd

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	packageManagerOs "github.com/qaware/minikube-support/pkg/packagemanager/os"
	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/spf13/cobra"
	"os"
)

// RootCommandOptions stores the the values for global command flags like the kubeconfig and context name.
type RootCommandOptions struct {
	kubeConfig                string
	contextName               string
	githubAccessToken         string
	installablePluginRegistry apis.InstallablePluginRegistry
	startStopPluginRegistry   apis.StartStopPluginRegistry
	preRunInit                []PreRunInit
	contextNameSupplier       ContextNameSupplier
}

// PreRunInit defines the interface for small helper functions which will perform
// a late initialization of some already initialized object instances.
type PreRunInit func(options *RootCommandOptions) error

var rootCmd *cobra.Command

// NewRootCmd initializes the root cobra command including all the flags on root level.
func NewRootCmd() (*cobra.Command, *RootCommandOptions) {
	options := NewRootCmdOptions()
	cmd := &cobra.Command{
		Use:               "minikube-support",
		Short:             "The minikube-support tools helps to integrate minikube better into your local os.",
		PersistentPreRunE: options.preRun,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&options.kubeConfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")
	flags.StringVar(&options.contextName, "context", "", "The name of the kubeconfig context to use")
	flags.StringVar(&options.githubAccessToken, "ghAccessToken", "", "The github access token to access private repositories or avoid rate limiting.\nSee https://github.blog/2013-05-16-personal-api-tokens/ for information about how to create such a token.")

	return cmd, options
}

// preRun will be called by cobra as PersistentPreRun function before the actual command will be executed.
// This allows to inject configuration and argument values into already pre initialized objects.
func (o *RootCommandOptions) preRun(_ *cobra.Command, _ []string) error {
	var errors *multierror.Error
	for _, f := range o.preRunInit {
		errors = multierror.Append(errors, f(o))
	}
	return errors.ErrorOrNil()
}

// AddPreRunInitFunction adds a new PreRunInit function to the list for preRun()
func (o *RootCommandOptions) AddPreRunInitFunction(f PreRunInit) {
	o.preRunInit = append(o.preRunInit, f)
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
	runCmd := NewRunCommand(options.startStopPluginRegistry, options.contextNameSupplier)
	rootCmd.AddCommand(runCmd)
	for _, plugin := range options.startStopPluginRegistry.ListPlugins() {
		if plugin.IsSingleRunnable() {
			runCmd.AddCommand(NewRunSingleCommand(plugin))
		}
	}
	packageManagerOs.InitOsPackageManager()
}
