package coredns

import (
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"

	"github.com/chr-fritz/minikube-support/pkg/apis"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		plugin  apis.StartStopPlugin
		want    Manager
		wantErr bool
	}{
		{"ok", initMockGrpcPlugin(), &grpcManager{plugin: initMockGrpcPlugin()}, false},
		{"nil plugin", nil, nil, true},
		{"other plugin", testPlugin{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewManager(tt.plugin)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grpcManager_AddHost(t *testing.T) {
	tests := []struct {
		name        string
		hostName    string
		ip          string
		wantErr     bool
		plugin      *grpcPlugin
		recordFound bool
		wantType    uint16
	}{
		{"ok", "test.", "127.0.0.1", false, initMockGrpcPlugin(), true, dns.TypeA},
		{"okv6", "test.", "::1", false, initMockGrpcPlugin(), true, dns.TypeAAAA},
		{"no server", "test.", "::1", true, &grpcPlugin{}, false, dns.TypeAAAA},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &grpcManager{
				plugin: tt.plugin,
			}
			if err := m.AddHost(tt.hostName, tt.ip); (err != nil) != tt.wantErr {
				t.Errorf("grpcManager.AddHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if m.plugin.server != nil {
				checkResourceRecord(t, m.plugin.server, tt.hostName, dns.Type(tt.wantType), tt.recordFound)
			}
		})
	}
}

func Test_grpcManager_AddAlias(t *testing.T) {
	tests := []struct {
		name        string
		hostName    string
		target      string
		plugin      *grpcPlugin
		wantErr     bool
		recordFound bool
	}{
		{"ok", "test.", "target.test.", initMockGrpcPlugin(), false, true},
		{"ok", "test.", "target.test.", &grpcPlugin{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &grpcManager{
				plugin: tt.plugin,
			}
			if err := m.AddAlias(tt.hostName, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("grpcManager.AddAlias() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if m.plugin.server != nil {
				checkResourceRecord(t, m.plugin.server, tt.hostName, dns.Type(dns.TypeCNAME), tt.recordFound)
			}
		})
	}
}

func Test_grpcManager_RemoveHost(t *testing.T) {
	tests := []struct {
		name            string
		hostName        string
		serverInit      bool
		expectedRecords int
	}{
		{"remove domain", "domain", true, 1},
		{"remove domain no server", "domain", false, 4},
		{"remove domain1", "domain1", true, 3},
		{"remove other", "other", true, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			srv := NewServer()
			assert.NoError(t, srv.AddA("domain.", net.ParseIP("127.0.0.2")))
			assert.NoError(t, srv.AddA("domain1.", net.ParseIP("127.0.0.2")))
			assert.NoError(t, srv.AddAAAA("domain.", net.ParseIP("::1")))
			assert.NoError(t, srv.AddCNAME("domain.", "domain1."))

			var m Manager
			if tt.serverInit {
				m = &grpcManager{plugin: &grpcPlugin{server: srv}}
			} else {
				m = &grpcManager{&grpcPlugin{}}
			}

			m.RemoveHost(tt.hostName)

			got := srv.ListRRs()
			assert.Equal(t, tt.expectedRecords, len(got))
		})
	}
}

func initMockGrpcPlugin() *grpcPlugin {
	return &grpcPlugin{
		server: NewServer(),
	}
}
