package coredns

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/utils"
)

const GrpcPluginName = "coredns-grpc"

type grpcPlugin struct {
	server            *server
	monitoringChannel chan *apis.MonitoringMessage
	terminationChan   chan bool
	runner            Runner
}

// NewGrpcPlugin initializes a new StartStopPlugin that will controls the lifecycle of the server instance.
func NewGrpcPlugin(prefix string) apis.StartStopPlugin {
	return &grpcPlugin{
		terminationChan: make(chan bool),
		runner:          newRunner(newCoreDnsPaths(prefix)),
	}
}

// GetServer returns the backend grpc server if it is the grpcPlugin.
// Otherwise it returns an error.
func GetServer(plugin apis.StartStopPlugin) (*server, error) {
	p, ok := plugin.(*grpcPlugin)
	if !ok {
		return nil, fmt.Errorf("try to get server from unknown plugin type %s", plugin)
	}
	if p == nil {
		return nil, fmt.Errorf("no grpcPlugin found")
	}
	if p.server == nil {
		return nil, fmt.Errorf("grpcServer not initialized")
	}
	return p.server, nil
}

// String returns the plugin name.
func (grpcPlugin) String() string {
	return GrpcPluginName
}

func (grpcPlugin) IsSingleRunnable() bool {
	return false
}

// Start starts the server to allow registering new entries and answers queries from CoreDNS.
func (p *grpcPlugin) Start(monitoringChannel chan *apis.MonitoringMessage) (boxName string, e error) {
	p.monitoringChannel = monitoringChannel
	socket, e := net.Listen("tcp", ":8053")
	if e != nil {
		return "", fmt.Errorf("unable to open socket: %s", e)
	}

	p.server = NewServer()
	p.server.Start(socket)
	e = p.runner.Start()
	if e != nil {
		return "", fmt.Errorf("can not start coredns: %s", e)
	}
	go utils.Ticker(p.listRRsForUI, p.terminationChan, 2500*time.Millisecond)

	return GrpcPluginName, nil
}

// listRRsForUI creates a list of all currently stored resource records
// and sends them using the monitoring channel.
func (p *grpcPlugin) listRRsForUI() {
	rrs := p.server.ListRRs()
	rrStrings := make([]string, len(rrs))
	for i, v := range rrs {
		rrStrings[i] = strings.ReplaceAll(v.String(), "\t", "\t ") + "\n"
	}

	table, e := utils.FormatAsTable(rrStrings, "Name\t TTL\t Type\t RR\t Value\n")
	if e != nil {
		return
	}

	p.monitoringChannel <- &apis.MonitoringMessage{Box: GrpcPluginName, Message: table}
}

// Stop terminates the server instance.
func (p *grpcPlugin) Stop() error {
	p.terminationChan <- true
	p.server.Stop()
	return p.runner.Stop()
}
