package coredns

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=runner.go

import (
	"fmt"
	"github.com/qaware/minikube-support/pkg/sh"
	"os"
	"os/exec"
	"syscall"
)

// Runner defines the interface for running the coreDns server in non daemon mode.
type Runner interface {
	// Start starts the coreDns server.
	Start() error
	// Stop stops the coreDns server.
	Stop() error
}

type runner struct {
	prefix prefix
	*os.Process
}

type noOpRunner struct{}

func newRunner(prefix prefix) Runner {
	if runAsDaemon {
		return &noOpRunner{}
	} else {
		return &runner{
			prefix: prefix,
		}
	}
}

func (r runner) Start() error {
	cmd := sh.ExecSudoCommand(
		r.prefix.binary(),
		"-conf", r.prefix.coreFile(),
		"-pidfile", r.prefix.pidFile(),
	)
	e := r.run(cmd)
	r.Process = cmd.Process
	return e
}

func (r runner) run(cmd *exec.Cmd) error {
	stdOut, e := os.Create(r.prefix.logFile())
	if e != nil {
		return fmt.Errorf("can not open logfile %s for coredns: %s", r.prefix.logFile(), e)

	}

	stdErr, e := os.Create(r.prefix.errorLogFile())
	if e != nil {
		return fmt.Errorf("can not open error logfile %s for coredns: %s", r.prefix.errorLogFile(), e)
	}

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	return cmd.Start()
}

func (r runner) Stop() error {
	if r.Process == nil {
		return nil
	}
	return r.Process.Signal(syscall.SIGTERM)
}

func (n noOpRunner) Start() error {
	return nil
}

func (n noOpRunner) Stop() error {
	return nil
}
