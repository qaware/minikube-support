package minikube

import (
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/chr-fritz/minikube-support/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func Test_ip_addVmIp(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	manager := newTestManager(t)
	i := &ip{
		dnsBackendManager: manager,
	}
	i.addVmIp()

	assert.Equal(t, "vm.minikube", manager.addedHosts[0])
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
