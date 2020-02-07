package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/buger/goterm"
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var boxConfig = [][]string{{"k8sdns-ingress", "k8sdns-service"}, {"coredns-grpc", "minikube-tunnel"}, {"logs"}}

const BORDER_STRING = "─ │ ┌ ┐ └ ┘"

var terminalWidth = goterm.Width
var terminalHeight = goterm.Height
var terminalPrint = goterm.Print

type RunOptions struct {
	plugins        []apis.StartStopPlugin
	messageChannel chan *apis.MonitoringMessage
	activePlugins  []string
	lastMessages   map[string]*apis.MonitoringMessage
	contextName    ContextNameSupplier
}

type ContextNameSupplier func() string

func NewRunOptions(registry apis.StartStopPluginRegistry, contextName ContextNameSupplier) *RunOptions {
	if contextName == nil {
		contextName = func() string { return "no contextName supplier set" }
	}
	return &RunOptions{
		messageChannel: make(chan *apis.MonitoringMessage),
		plugins:        registry.ListPlugins(),
		lastMessages:   map[string]*apis.MonitoringMessage{},
		contextName:    contextName,
	}
}

func NewRunCommand(registry apis.StartStopPluginRegistry, contextName ContextNameSupplier) *cobra.Command {
	options := NewRunOptions(registry, contextName)

	command := &cobra.Command{
		Use:   "run",
		Short: "Run all or one of the available plugins.",
		Run:   options.Run,
	}

	return command
}

func (i *RunOptions) Run(cmd *cobra.Command, args []string) {
	if e := sh.InitSudo(); e != nil {
		logrus.Errorf("`minikube-support run` requires sudo for some plugins. Initialize sudo failed: %s", e)
	}

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
	vBoxSizes := calcBoxSize(terminalHeight()-yOffset, len(boxConfig))

	goterm.MoveCursor(1, 1)
	errors = multierror.Append(errors, printHeader(i.contextName()))

	nextY := 1 + yOffset
	for line, boxLineConfig := range boxConfig {
		hBoxSizes := calcBoxSize(terminalWidth(), len(boxLineConfig))
		nextX := 1
		for col, boxName := range boxLineConfig {
			message := i.lastMessages[boxName]
			if message == nil {
				continue
			}
			box := goterm.NewBox(hBoxSizes[col], vBoxSizes[line], 0)
			box.Border = BORDER_STRING
			_, e := fmt.Fprintf(box, "%s Status:\n%s", strings.Title(message.Box), strings.ReplaceAll(message.Message, "\t", "    "))
			errors = multierror.Append(errors, e)

			_, e = terminalPrint(goterm.MoveTo(box.String(), nextX, nextY))

			nextX += hBoxSizes[col]
			errors = multierror.Append(errors, e)
		}
		nextY += vBoxSizes[line]
	}

	goterm.Flush()
	return errors.ErrorOrNil()
}

func calcBoxSize(available int, numBoxes int) []int {
	baseSize := available / numBoxes
	mod := available % numBoxes

	result := make([]int, numBoxes)

	for i := 0; i < numBoxes; i++ {
		size := baseSize
		if i < mod {
			size++
		}

		result[i] = size
	}

	return result
}

func printHeader(k8sContext string) error {
	left := fmt.Sprintf("Kubernetes Kontext: %s", k8sContext)
	right := time.Now().Format(time.UnixDate)

	spaceLen := terminalWidth() - len(left) - len(right)
	if spaceLen < 1 {
		spaceLen = 1
	}

	space := strings.Repeat(" ", spaceLen)

	_, err := terminalPrint(left, space, right)
	return err
}
