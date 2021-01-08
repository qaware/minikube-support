package minikube

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/sh"
)

type tunnel struct {
	command        *exec.Cmd
	contextHandler kubernetes.ContextHandler
	runWait        *sync.WaitGroup
}

func NewTunnel(handler kubernetes.ContextHandler) apis.StartStopPlugin {
	group := &sync.WaitGroup{}
	group.Add(1)
	return &tunnel{contextHandler: handler, runWait: group}
}

const tunnelBoxName = "minikube-tunnel"

func (*tunnel) IsSingleRunnable() bool {
	return true
}

func (t *tunnel) Start(monitoringChannel chan *apis.MonitoringMessage) (boxName string, err error) {
	e := sh.InitSudo()
	monitoringChannel <- &apis.MonitoringMessage{Box: tunnelBoxName, Message: "Starting minikube tunnel"}
	if e != nil {
		return "", fmt.Errorf("unable to enter sudo mode for minikube tunnel: %e", e)
	}

	isMinikube, e := t.contextHandler.IsMinikube()
	if e != nil {
		return "", e
	}
	if !isMinikube {
		return "", nil
	}

	t.command = sh.ExecSudoCommand("minikube", "tunnel")
	t.command.Env = append(t.command.Env, os.Environ()...)
	stdoutPipe, e := t.command.StdoutPipe()
	if e != nil {
		return "", fmt.Errorf("can not open stdout: %s", e)
	}

	scanner := initScanner(stdoutPipe)
	go scanForStatusMessages(scanner, monitoringChannel)

	go func() {
		e := t.command.Start()
		if e != nil {
			logrus.Errorf("Can not start minikube tunnel: %s", e)
		}
		t.runWait.Done()
		e = t.command.Wait()
		if e != nil {
			logrus.Errorf("Unexpected end of minikube tunnel: %s", e)
		}
	}()

	return tunnelBoxName, nil
}

func scanForStatusMessages(scanner *bufio.Scanner, monitoringChannel chan *apis.MonitoringMessage) {
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		monitoringChannel <- &apis.MonitoringMessage{
			Box:     tunnelBoxName,
			Message: text,
		}
	}
}

func (t *tunnel) Stop() error {
	t.runWait.Wait()
	return t.command.Process.Signal(syscall.SIGTERM)
}
func (t *tunnel) String() string {
	return tunnelBoxName
}

func initScanner(closer io.ReadCloser) *bufio.Scanner {
	scanner := bufio.NewScanner(closer)
	scanner.Split(statusSplitter)
	return scanner
}

const statusSeperator = "Status:"

func statusSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	sep := len([]byte(statusSeperator))

	if i := bytes.Index(data, []byte(statusSeperator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
