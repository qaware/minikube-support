package ingress

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sort"
	"testing"

	"k8s.io/api/extensions/v1beta1"
)

func Test_k8sIngress_AddedEvent(t *testing.T) {
	tests := []struct {
		name         string
		ingress      v1beta1.Ingress
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
			k8s := &k8sIngress{
				recordManager:    manager,
				currentIngresses: make(map[string]ingressEntry),
			}
			if err := k8s.AddedEvent(&tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sIngress.AddedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(manager.addedHosts, tt.wantAddHosts) {
				t.Errorf("k8sIngress.AddedEvent() addedHosts = %v, wantAddedHosts %v", manager.addedHosts, tt.wantAddHosts)
			}
			if !reflect.DeepEqual(manager.addedAlias, tt.wantAddAlias) {
				t.Errorf("k8sIngress.AddedEvent() wantAddAlias = %v, wantAddAlias %v", manager.addedAlias, tt.wantAddAlias)
			}
		})
	}
}

func Test_k8sIngress_UpdatedEvent(t *testing.T) {
	tests := []struct {
		name             string
		currentIngresses map[string]ingressEntry
		ingress          v1beta1.Ingress
		wantAddHosts     []string
		wantAddAlias     []string
		wantRemovedHosts []string
		wantErr          bool
	}{
		{
			"add host",
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{"t/t": {
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
			map[string]ingressEntry{},
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
			k8s := &k8sIngress{
				recordManager:    manager,
				currentIngresses: tt.currentIngresses,
			}
			if err := k8s.UpdatedEvent(&tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sIngress.UpdatedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(manager.addedHosts, tt.wantAddHosts) {
				t.Errorf("k8sIngress.UpdatedEvent() addedHosts = %v, wantAddedHosts %v", manager.addedHosts, tt.wantAddHosts)
			}
			if !reflect.DeepEqual(manager.addedAlias, tt.wantAddAlias) {
				t.Errorf("k8sIngress.UpdatedEvent() wantAddAlias = %v, wantAddAlias %v", manager.addedAlias, tt.wantAddAlias)
			}
			if !reflect.DeepEqual(manager.removedHosts, tt.wantRemovedHosts) {
				t.Errorf("k8sIngress.UpdatedEvent() removedHosts = %v, wantRemovedHosts %v", manager.removedHosts, tt.wantRemovedHosts)
			}
		})
	}
}

func Test_k8sIngress_DeletedEvent(t *testing.T) {
	tests := []struct {
		name             string
		ingress          v1beta1.Ingress
		wantRemovedHosts []string
		wantErr          bool
	}{
		{"ok", createDummyIngress("t", "t", "", "", "1"), []string{"1"}, false},
		{"ok 1", createDummyIngress("t", "t", "", "", "1", "2"), []string{"1", "2"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newTestManager(t)
			k8s := &k8sIngress{
				recordManager:    manager,
				currentIngresses: make(map[string]ingressEntry),
			}

			if err := k8s.DeletedEvent(&tt.ingress); (err != nil) != tt.wantErr {
				t.Errorf("k8sIngress.DeletedEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			sort.Strings(manager.removedHosts)
			if !reflect.DeepEqual(manager.removedHosts, tt.wantRemovedHosts) {
				t.Errorf("k8sIngress.DeletedEvent() removedHosts = %v, wantRemovedHosts %v", manager.removedHosts, tt.wantRemovedHosts)
			}
		})
	}
}

func createDummyIngress(name string, ns string, targetIp string, targetHost string, hosts ...string) v1beta1.Ingress {
	return v1beta1.Ingress{
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
