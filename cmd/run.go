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
	"time"
)

var boxConfig = [][]string{{"ingresses", "minikube-tunnel"}, {"coredns-grpc"}}

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
	go i.startPlugins()
	goterm.Clear()
	i.handleSignals()

	for message := range i.messageChannel {
		if message == apis.TerminatingMessage {
			return
		}
		i.lastMessages[message.Box] = message
		_ = i.renderBoxes()
	}
}

func (i *RunOptions) startPlugins() {
	for _, plugin := range i.plugins {
		boxName, err := plugin.Start(i.messageChannel)
		if err != nil {
			logrus.Errorf("Unable to start plugin %s: %s", plugin, err)
		} else {
			i.activePlugins = append(i.activePlugins, boxName)
			i.messageChannel <- &apis.MonitoringMessage{Box: boxName, Message: "Starting..."}
		}
	}
}

func (i *RunOptions) handleSignals() {
	signalsChannel := make(chan os.Signal)
	signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalsChannel
		logrus.Infof("Got signal %s. Terminating all plugins", sig)

		for _, plugin := range i.plugins {
			go func() {
				logrus.Debugf("Terminating plugin: %s", plugin)
				e := plugin.Stop()
				if e != nil {
					logrus.Warnf("Unable to terminate plugin %s: %s", plugin, e)
				}
			}()
		}
		i.messageChannel <- apis.TerminatingMessage
	}()
}

func (i *RunOptions) renderBoxes() error {
	var errors *multierror.Error
	yOffset := 1
	vBoxes := len(boxConfig)
	vBoxHeight := (goterm.Height() - yOffset) / vBoxes

	goterm.MoveCursor(1, 1)
	errors = multierror.Append(errors, printHeader(""))
	for line, boxLineConfig := range boxConfig {
		hBoxes := len(boxLineConfig)
		hBoxWidth := goterm.Width() / hBoxes

		for col, boxName := range boxLineConfig {
			message := i.lastMessages[boxName]
			if message == nil {
				continue
			}

			box := goterm.NewBox(hBoxWidth, vBoxHeight, 0)

			_, e := fmt.Fprintf(box, "%s Status:\n%s", strings.Title(message.Box), strings.ReplaceAll(message.Message, "\t", "    "))
			errors = multierror.Append(errors, e)

			_, e = goterm.Print(goterm.MoveTo(box.String(), hBoxWidth*col+1, vBoxHeight*line+1+yOffset))
			errors = multierror.Append(errors, e)
		}
	}

	goterm.Flush()
	return errors.ErrorOrNil()
}

func printHeader(k8sContext string) error {
	left := fmt.Sprintf("Kubernetes Kontext: %s", k8sContext)
	right := time.Now().Format(time.UnixDate)

	spaceLen := goterm.Width() - len(left) - len(right)
	if spaceLen < 0 {
		spaceLen = 0
	}

	space := strings.Repeat(" ", spaceLen)

	_, err := goterm.Print(left, space, right)
	return err
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
