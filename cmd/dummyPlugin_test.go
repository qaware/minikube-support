package cmd

import (
	"fmt"

	"github.com/qaware/minikube-support/pkg/apis"
)

type DummyPlugin struct {
	run          func(chan *apis.MonitoringMessage)
	name         string
	failStart    bool
	failStop     bool
	installRun   bool
	updateRun    bool
	uninstallRun bool
	purge        bool
	phase        apis.Phase
}

func (p *DummyPlugin) Install() {
	p.installRun = true
}

func (p *DummyPlugin) Update() {
	p.updateRun = true
}

func (p *DummyPlugin) Uninstall(purge bool) {
	p.uninstallRun = true
	p.purge = purge
}

func (p *DummyPlugin) String() string {
	if p.name != "" {
		return p.name
	}
	return "dummy"
}

func (p *DummyPlugin) Start(m chan *apis.MonitoringMessage) (boxName string, err error) {
	if p.failStart {
		return "", fmt.Errorf("fail")
	}
	if p.run != nil {
		go p.run(m)
	}
	return p.String(), nil
}

func (p *DummyPlugin) Stop() error {
	if p.failStop {
		return fmt.Errorf("fail")
	}
	return nil
}
func (p *DummyPlugin) IsSingleRunnable() bool {
	return false
}
func (p *DummyPlugin) Phase() apis.Phase {
	return p.phase
}
