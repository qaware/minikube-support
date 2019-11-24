package k8sdns

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_ingressEntry_String(t *testing.T) {
	tests := []struct {
		ingressName string
		namespace   string
		want        string
	}{
		{"test", "default", "default/test"},
	}
	for _, tt := range tests {
		t.Run(tt.ingressName, func(t *testing.T) {
			e := entry{
				name:      tt.ingressName,
				namespace: tt.namespace,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("entry.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_hasTargets(t *testing.T) {
	tests := []struct {
		name        string
		targetIps   []string
		targetHosts []string
		want        bool
	}{
		{"none", []string{}, []string{}, false},
		{"ip only", []string{"127.0.01"}, []string{}, true},
		{"host only", []string{}, []string{"localhost"}, true},
		{"both", []string{"127.0.0.1"}, []string{"localhost"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				targetIps:   tt.targetIps,
				targetHosts: tt.targetHosts,
			}
			if got := e.hasTargets(); got != tt.want {
				t.Errorf("entry.hasTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getAddedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, nil},
		{"one new", []string{"1", "2"}, []string{"1"}, []string{"2"}},
		{"old empty", []string{"1", "2"}, []string{}, []string{"1", "2"}},
		{"new empty", []string{}, []string{"1", "2"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				hostNames: tt.hostNames,
			}
			if got := e.getAddedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getAddedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getUpdatedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, []string{"1", "2"}},
		{"one new", []string{"1", "2"}, []string{"1"}, []string{"1"}},
		{"old empty", []string{"1", "2"}, []string{}, nil},
		{"new empty", []string{}, []string{"1", "2"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				hostNames: tt.hostNames,
			}
			if got := e.getUpdatedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getUpdatedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getRemovedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, nil},
		{"one new", []string{"1", "2"}, []string{"1"}, nil},
		{"one old", []string{"1"}, []string{"1", "2"}, []string{"2"}},
		{"old empty", []string{"1", "2"}, []string{}, nil},
		{"new empty", []string{}, []string{"1", "2"}, []string{"1", "2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{hostNames: tt.hostNames}
			if got := e.getRemovedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getRemovedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Improves the debug output when tests fail
func (e *entry) debug() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"%s %s {hostNames=%s, targetIps=%s, targetHosts=%s}",
		e.typ,
		e.String(), e.hostNames, e.targetIps, e.targetHosts)
}
