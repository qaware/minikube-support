package apis

import "fmt"

// The StartStopPlugin interface defines the interface for all plugins of the "run" command.
// They uses a channel to send new status updates which are shown by the "run" command.
type StartStopPlugin interface {
	// Must return the name of the plugin. This name will also be used for single commands.
	fmt.Stringer

	// Start the command and sets the channel used to show status and monitoring messages.
	// Returns a string that indicates the box name which is also present in the monitoring messages.
	Start(chan *MonitoringMessage) (boxName string, err error)

	// Stops the plugin for graceful shutdown.
	Stop() error

	// IsSingleRunnable determs if this plugin can be started using the "run <plugin>" command.
	IsSingleRunnable() bool
}

// StartStopPluginRegistry is the registry which collects all StartStopPlugins and provides easy access to them.
type StartStopPluginRegistry interface {
	// AddPlugin adds a single plugin to the registry.
	AddPlugin(plugin StartStopPlugin)
	// AddPlugins adds the given list of plugins to the registry.
	AddPlugins(plugins ...StartStopPlugin)
	// ListPlugins returns a list of all currently registered plugins in the registry.
	ListPlugins() []StartStopPlugin
	// FindPlugin tries to find and return a plugin with the given name. Otherwise it would return an error.
	FindPlugin(name string) (StartStopPlugin, error)
}

// The message type for notifying the run cli command about new status and monitoring messages from the plugin.
type MonitoringMessage struct {
	Box     string
	Message string
}

// CloneMonitoringMessage creates a copy of the given MonitoringMessage.
func CloneMonitoringMessage(message *MonitoringMessage) MonitoringMessage {
	return MonitoringMessage{
		Box:     message.Box,
		Message: message.Message,
	}
}

// The terminating message. If this message will be send, the run command will shutdown all other start stop plugins and ends.
var TerminatingMessage = &MonitoringMessage{
	Box:     "terminating",
	Message: "terminating",
}
