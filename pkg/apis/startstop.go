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
}

// The message type for notifying the run cli command about new status and monitoring messages from the plugin.
type MonitoringMessage struct {
	Box     string
	Message string
}

// The terminating message. If this message will be send, the run command will shutdown all other start stop plugins and ends.
var TerminatingMessage = &MonitoringMessage{
	Box:     "terminating",
	Message: "terminating",
}
