package coredns

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
)

func Test_runner(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	tmpdir, e := ioutil.TempDir(os.TempDir(), "coredns_test")

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
		sh.ExecCommand = exec.Command
	}()
	assert.NoError(t, e)
	paths := newCoreDnsPaths(tmpdir)
	test := testutils.TestProcessResponse{
		Command:        paths.binary(),
		Args:           []string{"-conf", paths.coreFile(), "-pidfile", paths.pidFile()},
		ResponseStatus: 0,
		Stdout:         "log",
		Stderr:         "err",
		Delay:          2 * time.Second,
	}
	testutils.SetTestProcessResponse(test)

	assert.NoError(t, os.MkdirAll(paths.logDir(), 0755))

	r := &runner{prefix: paths}
	assert.NoError(t, r.Start())
	time.Sleep(500 * time.Millisecond)
	assert.NoError(t, r.Stop())

	assert.FileExists(t, paths.logFile())
	assert.FileExists(t, paths.errorLogFile())
}
