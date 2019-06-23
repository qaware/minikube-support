package coredns

import (
	"bytes"
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/utils"
	"net"
	"time"
)

const GrpcPluginName = "coredns-grpc"

type grpcPlugin struct {
	server            *Server
	monitoringChannel chan *apis.MonitoringMessage
	terminationChan   chan bool
}

// NewGrpcPlugin initializes a new StartStopPlugin that will controls the lifecycle of the Server instance.
func NewGrpcPlugin() apis.StartStopPlugin {
	return &grpcPlugin{
		terminationChan: make(chan bool),
	}
}

// GetServer returns the backend grpc server if it is the grpcPlugin.
// Otherwise it returns an error.
func GetServer(plugin apis.StartStopPlugin) (*Server, error) {
	p, ok := plugin.(*grpcPlugin)
	if !ok {
		return nil, fmt.Errorf("try to get server from unknown plugin type %s", plugin)
	}
	if p == nil {
		return nil, fmt.Errorf("no grpcPlugin found")
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

	go utils.Ticker(p.listRRsForUI, p.terminationChan, 5*time.Second)

	return GrpcPluginName, nil
}

// listRRsForUI creates a list of all currently stored resource records
// and sends them using the monitoring channel.
func (p *grpcPlugin) listRRsForUI() {
	rrs := p.server.ListRRs()
	buf := bytes.Buffer{}
	for _, v := range rrs {
		buf.WriteString(v.String())
		buf.WriteByte('\n')
	}
	p.monitoringChannel <- &apis.MonitoringMessage{
		Box:     GrpcPluginName,
		Message: buf.String(),
	}
}

// Stop terminates the Server instance.
func (p *grpcPlugin) Stop() error {
	p.terminationChan <- true
	p.server.Stop()
	return nil
}
