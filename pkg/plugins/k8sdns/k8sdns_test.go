package k8sdns

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"testing"

	"k8s.io/api/extensions/v1beta1"
)

func Test_k8sIngress_AddedEvent(t *testing.T) {
	tests := []struct {
		name         string
		ingress      *v1beta1.Ingress
		wantAddHosts []string
		wantAddAlias []string
		wantErr      bool
	}{
		{"ok only ip", createDummyIngress("t", "t", "127.0.0.1", "", "1"), []string{"1"}, []string{}, false},
		{"ok both", createDummyIngress("t", "t", "127.0.0.1", "l", "1"), []string{"1"}, []string{"1"}, false},
		{"ok only host", createDummyIngress("t", "t", "", "l", "1"), []string{}, []string{"1"}, false},
		{"no target", createDummyIngress("t", "t", "", "", "1"), []string{}, []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newTestManager(t)
			k8s := &k8sDns{
				recordManager:  manager,
				currentEntries: make(map[string]*entry),
				accessor:       ingressAccessor{},
			}
			if err := k8s.AddedEvent(tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sDns.AddedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantAddHosts, manager.addedHosts)
			assert.Equal(t, tt.wantAddAlias, manager.addedAlias)
		})
	}
}

func Test_k8sIngress_UpdatedEvent(t *testing.T) {
	tests := []struct {
		name             string
		currentIngresses map[string]*entry
		ingress          *v1beta1.Ingress
		wantAddHosts     []string
		wantAddAlias     []string
		wantRemovedHosts []string
		wantErr          bool
	}{
		{
			"add host",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{},
				targetIps:   []string{},
				targetHosts: []string{},
			}},
			createDummyIngress("t", "t", "127.0.0.1", "", "1"),
			[]string{"1"},
			[]string{},
			[]string{},
			false,
		},
		{
			"add alias",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{},
				targetIps:   []string{},
				targetHosts: []string{},
			}},
			createDummyIngress("t", "t", "", "localhost", "1"),
			[]string{},
			[]string{"1"},
			[]string{},
			false,
		},
		{
			"no target",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1"},
				targetIps:   []string{"127.0.0.1"},
				targetHosts: []string{},
			}},
			createDummyIngress("t", "t", "", "", "1"),
			[]string{},
			[]string{},
			[]string{"1"},
			false,
		},
		{
			"other ip",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1"},
				targetIps:   []string{"127.0.0.2"},
				targetHosts: []string{},
			}},
			createDummyIngress("t", "t", "127.0.0.1", "", "1"),
			[]string{"1"},
			[]string{},
			[]string{"1"},
			false,
		},
		{
			"other target host",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1"},
				targetIps:   []string{},
				targetHosts: []string{"dummy"},
			}},
			createDummyIngress("t", "t", "", "localhost", "1"),
			[]string{},
			[]string{"1"},
			[]string{"1"},
			false,
		},
		{
			"new hostname",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1"},
				targetIps:   []string{"127.0.0.1"},
				targetHosts: nil,
			}},
			createDummyIngress("t", "t", "127.0.0.1", "", "1", "2"),
			[]string{"2"},
			[]string{},
			[]string{},
			false,
		},
		{
			"remove hostname",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1", "2"},
				targetIps:   []string{"127.0.0.1"},
				targetHosts: nil,
			}},
			createDummyIngress("t", "t", "127.0.0.1", "", "1"),
			[]string{},
			[]string{},
			[]string{"2"},
			false,
		},
		{
			"new hostname with alias",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1"},
				targetIps:   nil,
				targetHosts: []string{"localhost"},
			}},
			createDummyIngress("t", "t", "", "localhost", "1", "2"),
			[]string{},
			[]string{"2"},
			[]string{},
			false,
		},
		{
			"remove hostname with alias",
			map[string]*entry{"t/t": {
				name:        "t",
				namespace:   "t",
				hostNames:   []string{"1", "2"},
				targetIps:   nil,
				targetHosts: []string{"localhost"},
			}},
			createDummyIngress("t", "t", "", "localhost", "1"),
			[]string{},
			[]string{},
			[]string{"2"},
			false,
		},
		{
			"no old entry",
			map[string]*entry{},
			createDummyIngress("t", "t", "", "localhost", "1"),
			[]string{},
			[]string{"1"},
			[]string{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newTestManager(t)
			k8s := &k8sDns{
				recordManager:  manager,
				currentEntries: tt.currentIngresses,
				accessor:       ingressAccessor{},
			}
			if err := k8s.UpdatedEvent(tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sDns.UpdatedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.wantAddHosts, manager.addedHosts)
			assert.Equal(t, tt.wantAddAlias, manager.addedAlias)
			assert.Equal(t, tt.wantRemovedHosts, manager.removedHosts)
		})
	}
}

func Test_k8sIngress_DeletedEvent(t *testing.T) {
	tests := []struct {
		name             string
		ingress          *v1beta1.Ingress
		wantRemovedHosts []string
		wantErr          bool
	}{
		{"ok", createDummyIngress("t", "t", "", "", "1"), []string{"1"}, false},
		{"ok 1", createDummyIngress("t", "t", "", "", "1", "2"), []string{"1", "2"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newTestManager(t)
			k8s := &k8sDns{
				recordManager:  manager,
				currentEntries: make(map[string]*entry),
				accessor:       ingressAccessor{},
			}

			if err := k8s.DeletedEvent(tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sDns.DeletedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			sort.Strings(manager.removedHosts)
			assert.Equal(t, tt.wantRemovedHosts, manager.removedHosts)
		})
	}
}

func createDummyIngress(name string, ns string, targetIp string, targetHost string, hosts ...string) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: v1meta.ObjectMeta{Name: name, Namespace: ns},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{{Hosts: hosts}},
		},
		Status: v1beta1.IngressStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: targetIp}, {Hostname: targetHost}}}},
	}
}

type testManager struct {
	t            *testing.T
	addedHosts   []string
	addedAlias   []string
	removedHosts []string
}

func newTestManager(t *testing.T) *testManager {
	return &testManager{t, make([]string, 0), make([]string, 0), make([]string, 0)}
}

func (m *testManager) AddHost(hostName string, ip string) error {
	m.addedHosts = append(m.addedHosts, hostName)
	assert.NotEmpty(m.t, ip)
	return nil
}

func (m *testManager) AddAlias(hostName string, target string) error {
	m.addedAlias = append(m.addedAlias, hostName)
	assert.NotEmpty(m.t, target)
	return nil
}

func (m *testManager) RemoveHost(hostName string) {
	m.removedHosts = append(m.removedHosts, hostName)
}

func Test_k8sIngress_PostEvent(t *testing.T) {
	tests := []struct {
		name           string
		currentEntries map[string]*entry
		wantMessage    string
		wantErr        bool
	}{
		{
			"no entries",
			map[string]*entry{},
			"Name | Namespace | Typ | Hostname | Targets\n",
			false,
		}, {
			"one ingress",
			map[string]*entry{"test.abc": {
				name:      "test",
				namespace: "test",
				typ:       "Ingress",
				hostNames: []string{"host.abc"},
				targetIps: []string{"ip"},
			}},
			"Name | Namespace | Typ     | Hostname | Targets\ntest | test      | Ingress | host.abc | ip\n",
			false,
		}, {
			"one service",
			map[string]*entry{"test.abc": {
				name:      "test",
				namespace: "test",
				typ:       "Service",
				hostNames: []string{"host.abc"},
				targetIps: []string{"ip"},
			}},
			"Name | Namespace | Typ     | Hostname | Targets\ntest | test      | Service | host.abc | ip\n",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageChannel := make(chan *apis.MonitoringMessage, 1)
			k8s := &k8sDns{
				messageChannel: messageChannel,
				currentEntries: tt.currentEntries,
			}
			if err := k8s.PostEvent(); (err != nil) != tt.wantErr {
				t.Errorf("PostEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			msg := <-messageChannel
			assert.Equal(t, tt.wantMessage, msg.Message)
		})
	}
}
