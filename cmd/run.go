package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/buger/goterm"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/sh"
)

var boxConfig = [][]string{{"k8sdns-ingress", "k8sdns-service"}, {"coredns-grpc", "minikube-tunnel"}, {"logs"}}

const BORDER_STRING = "─ │ ┌ ┐ └ ┘"

var terminalLock = sync.Mutex{}
var terminalWidth = goterm.Width
var terminalHeight = goterm.Height
var terminalPrint = func(a ...interface{}) (int, error) {
	terminalLock.Lock()
	defer terminalLock.Unlock()
	return goterm.Print(a)
}

type RunOptions struct {
	plugins           []apis.StartStopPlugin
	messageChannel    chan *apis.MonitoringMessage
	activePlugins     []string
	activePluginsLock sync.RWMutex
	lastMessages      map[string]*apis.MonitoringMessage
	lastMessagesLock  sync.RWMutex
	contextName       ContextNameSupplier
}

type ContextNameSupplier func() string

func NewRunOptions(registry apis.StartStopPluginRegistry, contextName ContextNameSupplier) *RunOptions {
	if contextName == nil {
		contextName = func() string { return "no contextName supplier set" }
	}
	return &RunOptions{
		messageChannel:    make(chan *apis.MonitoringMessage),
		plugins:           registry.ListPlugins(),
		lastMessages:      map[string]*apis.MonitoringMessage{},
		contextName:       contextName,
		activePluginsLock: sync.RWMutex{},
		lastMessagesLock:  sync.RWMutex{},
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

func (i *RunOptions) Run(_ *cobra.Command, _ []string) {
	if e := sh.InitSudo(); e != nil {
		logrus.Errorf("`minikube-support run` requires sudo for some plugins. Initialize sudo failed: %s", e)
	}

	go i.startPlugins()
	terminalLock.Lock()
	goterm.Clear()
	terminalLock.Unlock()
	i.handleSignals()

	for message := range i.messageChannel {
		if message == apis.TerminatingMessage {
			return
		}
		i.lastMessagesLock.Lock()
		i.lastMessages[message.Box] = message
		i.lastMessagesLock.Unlock()
		_ = i.renderBoxes()
	}
}

func (i *RunOptions) startPlugins() {
	for _, plugin := range i.plugins {
		boxName, err := plugin.Start(i.messageChannel)
		if err != nil {
			logrus.Errorf("Unable to start plugin %s: %s", plugin, err)
		} else {
			i.activePluginsLock.Lock()
			i.activePlugins = append(i.activePlugins, boxName)
			i.activePluginsLock.Unlock()
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
			p := plugin
			go func() {
				logrus.Debugf("Terminating plugin: %s", p)
				e := p.Stop()
				if e != nil {
					logrus.Warnf("Unable to terminate plugin %s: %s", p, e)
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
	terminalLock.Lock()
	goterm.MoveCursor(1, 1)
	terminalLock.Unlock()
	errors = multierror.Append(errors, printHeader(i.contextName(), terminalWidth()))

	nextY := 1 + yOffset
	for line, boxLineConfig := range boxConfig {
		hBoxSizes := calcBoxSize(terminalWidth(), len(boxLineConfig))
		nextX := 1
		for col, boxName := range boxLineConfig {
			i.lastMessagesLock.RLock()
			message := i.lastMessages[boxName]
			i.lastMessagesLock.RUnlock()
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
	terminalLock.Lock()
	goterm.Flush()
	terminalLock.Unlock()
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

func printHeader(k8sContext string, width int) error {
	left := fmt.Sprintf("Kubernetes Kontext: %s", k8sContext)
	right := time.Now().Format(time.UnixDate)

	spaceLen := width - len(left) - len(right)
	if spaceLen < 1 {
		spaceLen = 1
	}

	space := strings.Repeat(" ", spaceLen)

	_, err := terminalPrint(left, space, right)
	return err
}
