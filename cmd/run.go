package cmd

import (
	"fmt"
	"github.com/buger/goterm"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var boxConfig = [][]string{{"ingresses", "minikube-tunnel"}, {"ingresses", "dummy", "minikube-tunnel"}}

type RunOptions struct {
	plugins        []apis.StartStopPlugin
	messageChannel chan *apis.MonitoringMessage
	activePlugins  []string
	lastMessages   map[string]*apis.MonitoringMessage
}

func NewRunOptions() *RunOptions {
	return &RunOptions{
		messageChannel: make(chan *apis.MonitoringMessage),
		plugins:        plugins.GetStartStopPlugins(),
		lastMessages:   map[string]*apis.MonitoringMessage{},
	}
}

func NewRunCommand() *cobra.Command {
	options := NewRunOptions()

	command := &cobra.Command{
		Use:   "run",
		Short: "Run all or one of the available plugins.",
		Run:   options.Run,
	}

	return command
}

func (i *RunOptions) Run(cmd *cobra.Command, args []string) {
	e := i.startPlugins()
	if e != nil {
		logrus.Errorf("Unable to start at least one plugin: %s", e)
		return
	}

	i.handleSignals()
	go func() {
		i.messageChannel <- &apis.MonitoringMessage{Box: "ingresses", Message: "http://chr-fritz.de\nhttps://chr-fritz.de"}
		i.messageChannel <- &apis.MonitoringMessage{Box: "dummy", Message: "http://chr-fritz.de\nhttps://chr-fritz.de"}
	}()

	for message := range i.messageChannel {
		if message == apis.TerminatingMessage {
			return
		}
		i.lastMessages[message.Box] = message
		_ = i.renderBoxes()
	}
}

func (i *RunOptions) startPlugins() error {
	var errors *multierror.Error
	for _, plugin := range i.plugins {
		boxName, err := plugin.Start(i.messageChannel)
		if err != nil {
			errors = multierror.Append(errors, err)
		} else {
			i.activePlugins = append(i.activePlugins, boxName)
			i.lastMessages[boxName] = &apis.MonitoringMessage{Box: boxName}
		}
	}
	return errors.ErrorOrNil()
}

func (i *RunOptions) handleSignals() {
	signalsChannel := make(chan os.Signal)
	signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalsChannel
		logrus.Infof("Got signal %s. Terminating all plugins", sig)

		for _, plugin := range i.plugins {
			logrus.Debugf("Terminating plugin: %s", plugin)
			e := plugin.Stop()
			if e != nil {
				logrus.Warnf("Unable to terminate plugin %s: %s", plugin, e)
			}
		}
		i.messageChannel <- apis.TerminatingMessage
	}()
}

func (i *RunOptions) renderBoxes() error {
	goterm.Clear()
	goterm.MoveCursor(1, 1)
	var errors *multierror.Error
	vBoxes := len(boxConfig)
	vBoxHeigh := goterm.Height() / vBoxes

	for line, boxLineConfig := range boxConfig {
		hBoxes := len(boxLineConfig)
		hBoxWidth := goterm.Width() / hBoxes

		for col, boxName := range boxLineConfig {
			message := i.lastMessages[boxName]
			if message == nil {
				continue
			}

			box := goterm.NewBox(hBoxWidth, vBoxHeigh, 0)

			_, e := fmt.Fprintf(box, "%s Status:\n%s", message.Box, strings.ReplaceAll(message.Message, "\t", "    "))
			if e != nil {
				errors = multierror.Append(errors, e)
			}
			_, e = goterm.Print(goterm.MoveTo(box.String(), hBoxWidth*col+1, vBoxHeigh*line+1))
			if e != nil {
				errors = multierror.Append(errors, e)
			}
		}
	}

	goterm.Flush() // Call it every time at the end of rendering

	return nil
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
