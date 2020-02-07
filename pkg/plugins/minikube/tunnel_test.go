package minikube

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_initScanner(t *testing.T) {
	file, e := os.Open("minikube-tunnel.txt")
	check(t, e)
	scanner := initScanner(file)

	expectedText := "	machine: minikube\n	pid: 68980\n	route: 10.96.0.0/12 -> 192.168.64.13\n	minikube: Running\n	services: []\n    errors: \n		minikube: no errors\n		router: no errors\n		loadbalancer emulator: no errors\n"

	expectedTexts := []string{
		"", expectedText, expectedText,
	}

	var texts []string
	i := 0
	for scanner.Scan() {
		texts = append(texts, scanner.Text())
		i++
	}

	assert.Equal(t, 3, i, "invalid count of messages")
	if !reflect.DeepEqual(texts, expectedTexts) {
		t.Errorf("invalid texts found: expected %s got %s", expectedTexts, texts)
	}
}

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf("Error occured: %s", e)
	}
}

func Test_tunnel_Start(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	wantBoxName := "minikube-tunnel"
	monitoringChannel := make(chan *apis.MonitoringMessage)
	handler := fake.NewContextHandler(nil, nil)
	handler.MiniKube = true
	mkt := NewTunnel(handler)
	var count = 0
	go func() {
		for message := range monitoringChannel {
			assert.Equal(t, wantBoxName, message.Box)
			assert.NotEmpty(t, message.Message)
			count++
		}
	}()
	gotBoxName, err := mkt.Start(monitoringChannel)
	time.Sleep(2 * time.Second)
	if (err != nil) != false {
		t.Errorf("tunnel.Start() error = %v, wantErr %v", err, false)
		return
	}
	if gotBoxName != wantBoxName {
		t.Errorf("tunnel.Start() = %v, want %v", gotBoxName, wantBoxName)
	}
	assert.Equal(t, 4, count)
}

func Test_tunnel_Start_noMinikube(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	wantBoxName := "minikube-tunnel"
	monitoringChannel := make(chan *apis.MonitoringMessage)
	handler := fake.NewContextHandler(nil, nil)
	handler.MiniKube = false
	mkt := NewTunnel(handler)
	var count = 0
	go func() {
		for message := range monitoringChannel {
			assert.Equal(t, wantBoxName, message.Box)
			assert.NotEmpty(t, message.Message)
			count++
		}
	}()
	gotBoxName, err := mkt.Start(monitoringChannel)
	assert.Equal(t, "", gotBoxName)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, count)
}

func Test_tunnel_Stop(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	monitoringChannel := make(chan *apis.MonitoringMessage, 4)
	handler := fake.NewContextHandler(nil, nil)
	handler.MiniKube = true
	mkt := NewTunnel(handler)

	_, _ = mkt.Start(monitoringChannel)
	time.Sleep(100 * time.Millisecond)
	err := mkt.Stop()
	if (err != nil) != false {
		t.Errorf("tunnel.Stop() error = %v, wantErr %v", err, false)
		return
	}
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	cmd, args := testutils.ExtractMockedCommandAndArgs()
	switch cmd {
	case "sudo":
		cmd, args := args[0], args[1:]
		switch cmd {
		case "minikube":
			cmd, _ := args[0], args[1:]
			switch cmd {
			case "tunnel":
				bytes, _ := ioutil.ReadFile("minikube-tunnel.txt")
				_, _ = fmt.Fprint(os.Stdout, string(bytes))
				time.Sleep(1 * time.Second)
				_, _ = fmt.Fprint(os.Stdout, string(bytes))
			}
		case "echo":
			_, _ = fmt.Fprint(os.Stdout, strings.Join(args, " "))
		}
	case "minikube":
		cmd, _ := args[0], args[1:]
		switch cmd {
		case "ip":
			_, _ = fmt.Fprintln(os.Stdout, "127.0.0.1")
		}
	case "which":
		cmd, _ := args[0], args[1:]
		switch cmd {
		case "sudo":
			_, _ = fmt.Fprintln(os.Stdout, "sudo")
		}
	default:
		os.Exit(1)
	}
}
