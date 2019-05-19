package cmd

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
)

type DummyPlugin struct {
	run          func(chan *apis.MonitoringMessage)
	failStart    bool
	failStop     bool
	installRun   bool
	updateRun    bool
	uninstallRun bool
	purge        bool
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

func (*DummyPlugin) String() string {
	return "dummy"
}

func (p *DummyPlugin) Start(m chan *apis.MonitoringMessage) (boxName string, err error) {
	if p.failStart {
		return "", fmt.Errorf("fail")
	}
	go p.run(m)
	return p.String(), nil
}

func (p *DummyPlugin) Stop() error {
	if p.failStop {
		return fmt.Errorf("fail")
	}
	return nil
}
