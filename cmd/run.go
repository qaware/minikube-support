package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/awesome-gocui/gocui"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/utils"
)

var boxConfig = [][]string{{"k8sdns-ingress", "k8sdns-service"}, {"coredns-grpc", "minikube-tunnel"}, {"logs"}}
var newGui = gocui.NewGui

type RunOptions struct {
	plugins          []apis.StartStopPlugin
	gui              *gocui.Gui
	messageChannel   chan *apis.MonitoringMessage
	lastMessages     map[string]*apis.MonitoringMessage
	lastMessagesLock sync.RWMutex
	contextName      ContextNameSupplier
}

type ContextNameSupplier func() string

func NewRunOptions(registry apis.StartStopPluginRegistry, contextName ContextNameSupplier) *RunOptions {
	if contextName == nil {
		contextName = func() string { return "no contextName supplier set" }
	}
	return &RunOptions{
		messageChannel:   make(chan *apis.MonitoringMessage, 100),
		plugins:          registry.ListPlugins(),
		lastMessages:     map[string]*apis.MonitoringMessage{},
		contextName:      contextName,
		lastMessagesLock: sync.RWMutex{},
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
	i.handleSignals()
	gui, e := newGui(gocui.Output256, true)
	if e != nil {
		logrus.Errorf("Can not start gui: %s", e)
		return
	}
	defer gui.Close()
	i.gui = gui
	i.gui.SetManager(i)
	if e = i.registerKeybindings(gui); e != nil {
		logrus.Errorf("Can not register keybindings: %s", e)
		return
	}
	go i.receiveMessages()
	i.header()
	if err := gui.MainLoop(); err != nil && !gocui.IsQuit(err) {
		logrus.Error(err)
	}
}

func (i *RunOptions) registerKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, i.quit); err != nil {
		return err
	}
	key, modifier := gocui.MustParse("q")
	if err := g.SetKeybinding("", key, modifier, i.quit); err != nil {
		return err
	}
	return nil
}

func (i *RunOptions) quit(_ *gocui.Gui, _ *gocui.View) error {
	logrus.Info("ctrl+c called")
	i.stopPlugins()
	return gocui.ErrQuit
}

func (i *RunOptions) startPlugins() {
	for _, plugin := range i.plugins {
		_, err := plugin.Start(i.messageChannel)
		if err != nil {
			logrus.Errorf("Unable to start plugin %s: %s", plugin, err)
		}
	}
}

func (i *RunOptions) receiveMessages() {
	for message := range i.messageChannel {
		if message == apis.TerminatingMessage {
			return
		}
		i.lastMessagesLock.Lock()
		i.lastMessages[message.Box] = message
		i.lastMessagesLock.Unlock()

		i.gui.UpdateAsync(updateBox(message))
	}
}

func updateBox(message *apis.MonitoringMessage) func(gui *gocui.Gui) error {
	msg := apis.CloneMonitoringMessage(message)
	return func(gui *gocui.Gui) error {
		view, e := gui.View(msg.Box)
		if e != nil {
			return e
		}
		view.Clear()
		view.WriteString(padLeft(msg.Message, 1))
		return nil
	}
}

func padLeft(message string, spaces uint8) string {
	padding := strings.Repeat(" ", int(spaces))
	return strings.ReplaceAll(padding+message, "\n", "\n"+padding)
}

func (i *RunOptions) handleSignals() {
	signalsChannel := make(chan os.Signal)
	signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalsChannel
		logrus.Infof("Got signal %s. Terminating all plugins", sig)
		i.stopPlugins()
	}()
}

func (i *RunOptions) stopPlugins() {
	for _, plugin := range i.plugins {
		p := plugin
		go func() {
			logrus.Debugf("Terminating plugin: %s", p)
			e := p.Stop()
			if e != nil {
				logrus.Warnf("Unable to terminate plugin %s: %s", p, e)
			}
			logrus.Debugf("Plugin %s successfully terminated", p)
		}()
	}
	i.messageChannel <- apis.TerminatingMessage
}

func (i *RunOptions) Layout(gui *gocui.Gui) error {
	x, y := gui.Size()
	header, e := gui.SetView("header", -1, -1, x, 1, 0)
	if e != nil && !gocui.IsUnknownView(e) {
		return e
	}
	header.Frame = false

	yOffset := 1
	boxHeights := calcBoxSize(y-yOffset, len(boxConfig))

	nextY := yOffset
	for line, boxLineConfig := range boxConfig {
		boxWidths := calcBoxSize(x, len(boxLineConfig))
		nextX := 0
		for col, boxName := range boxLineConfig {
			pluginView, e := gui.SetView(boxName, nextX, nextY, nextX+boxWidths[col]-1, nextY+boxHeights[line]-1, 0)
			if e != nil && !gocui.IsUnknownView(e) {
				return e
			}
			pluginView.Frame = true
			pluginView.Title = " Status of " + strings.Title(boxName) + " "
			nextX += boxWidths[col]
		}
		nextY += boxHeights[line]
	}

	return nil
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

func (i *RunOptions) header() chan bool {
	done := make(chan bool)
	go utils.Ticker(func() {
		i.gui.Update(func(gui *gocui.Gui) error {
			headerView, e := gui.View("header")
			if e != nil {
				return e
			}
			width, _ := headerView.Size()
			headerView.Clear()
			_, e = fmt.Fprint(headerView, createHeader(i.contextName(), width-1))
			return e
		})
	}, done, 1*time.Second)
	return done
}

func createHeader(k8sContext string, width int) string {
	left := fmt.Sprintf("Kubernetes Kontext: %s", k8sContext)
	right := time.Now().Format(time.UnixDate)

	spaceLen := width - len(left) - len(right)
	if spaceLen < 1 {
		spaceLen = 1
	}

	space := strings.Repeat(" ", spaceLen)

	return left + space + right
}
