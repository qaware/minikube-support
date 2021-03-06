package coredns

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qaware/minikube-support/pkg/apis"
)

func TestGetServer(t *testing.T) {
	tests := []struct {
		name    string
		plugin  apis.StartStopPlugin
		want    *server
		wantErr bool
	}{
		{"ok", &grpcPlugin{server: NewServer()}, NewServer(), false},
		{"invalid plugin", &testPlugin{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetServer(tt.plugin)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grpcPlugin_listRRsForUI(t *testing.T) {
	p := &grpcPlugin{
		server:            NewServer(),
		monitoringChannel: make(chan *apis.MonitoringMessage),
	}
	assert.NoError(t, p.server.AddAAAA("localhost", net.ParseIP("::1")))
	assert.NoError(t, p.server.AddA("localhost", net.ParseIP("127.0.0.1")))
	go p.listRRsForUI()
	msg := <-p.monitoringChannel
	assert.Equal(t, GrpcPluginName, msg.Box)
	assert.Equal(t, "Name       | TTL | Type | RR   | Value\nlocalhost. | 10  | IN   | A    | 127.0.0.1\nlocalhost. | 10  | IN   | AAAA | ::1\n", msg.Message)
}

type testPlugin struct{}

func (testPlugin) String() string {
	return "test-plugin"
}

func (testPlugin) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	panic("implement me")
}

func (testPlugin) Stop() error {
	panic("implement me")
}
func (testPlugin) IsSingleRunnable() bool {
	return false
}
