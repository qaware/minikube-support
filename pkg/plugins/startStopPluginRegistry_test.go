package plugins

import "github.com/chr-fritz/minikube-support/pkg/apis"

func (p *DummyPlugin) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	return p.String(), nil
}

func (p *DummyPlugin) Stop() error {
	return nil
}
