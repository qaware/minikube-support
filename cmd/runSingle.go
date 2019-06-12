package cmd

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// RunSingleOptions contains all options and information that are needed to run a single plugin from the command line.
type RunSingleOptions struct {
	plugin         apis.StartStopPlugin
	messageChannel chan *apis.MonitoringMessage
}

// NewRunSingleOptions create a new instance of the RunSingleOptions for the given plugin.
func NewRunSingleOptions(plugin apis.StartStopPlugin) *RunSingleOptions {
	return &RunSingleOptions{
		plugin:         plugin,
		messageChannel: make(chan *apis.MonitoringMessage),
	}
}

// NewRunSingleCommand creates a new run command for the given plugin.
func NewRunSingleCommand(plugin apis.StartStopPlugin) *cobra.Command {
	options := NewRunSingleOptions(plugin)

	return &cobra.Command{
		Use:   plugin.String(),
		Short: fmt.Sprintf("Only run the %s plugin", plugin.String()),
		Run:   options.Run,
	}
}

// Run starts the plugin and waits for new status messages.
func (i *RunSingleOptions) Run(cmd *cobra.Command, args []string) {
	signalsChannel := make(chan os.Signal)
	signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)
	startError := make(chan error)

	go func() {
		_, e := i.plugin.Start(i.messageChannel)
		if e != nil {
			startError <- e
		}
	}()

	logrus.Infof("Plugin %s successfully started. Waiting for status...", i.plugin)
	for {
		select {
		case message := <-i.messageChannel:
			logrus.Infof("New %s status:\n%s", i.plugin, message.Message)
		case sig := <-signalsChannel:
			logrus.Debugf("Received signal %s terminating plugin: %s", sig, i.plugin)
			e := i.plugin.Stop()
			if e != nil {
				logrus.Warnf("Unable to terminate plugin %s: %s", i.plugin, e)
			}
			return
		case e := <-startError:
			logrus.Errorf("Can not start plugin %s: %s", i.plugin, e)
			return
		}
	}
}
