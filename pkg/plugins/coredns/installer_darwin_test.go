// +build darwin

package coredns

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kballard/go-shellquote"
	"github.com/qaware/minikube-support/pkg/github/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_installer_Install(t *testing.T) {
	ctrl := gomock.NewController(t)
	sh.ExecCommand = testutils.FakeExecCommand
	ghClient := fake.NewMockClient(ctrl)
	tmpdir, e := ioutil.TempDir(os.TempDir(), "coredns_test")
	assetName := fmt.Sprintf("coredns_1.0.0_%s_%s.tgz", runtime.GOOS, runtime.GOARCH)

	defer func() {
		defer ctrl.Finish()
		assert.NoError(t, os.RemoveAll(tmpdir))
		sh.ExecCommand = exec.Command
	}()
	assert.NoError(t, e)

	i := &installer{
		ghClient: ghClient,
		prefix:   prefix(tmpdir),
	}
	ghClient.EXPECT().
		GetLatestReleaseTag("coredns", "coredns").
		Return("v1.0.0", nil)
	ghClient.EXPECT().
		DownloadReleaseAsset("coredns", "coredns", "v1.0.0", assetName).
		Return(os.Open("fixtures/coredns.tar.gz"))

	mockWriteFileAsRoot(launchctlConfig, nil)
	mockWriteFileAsRoot(dotMinikubeResolverPath, nil)
	testutils.MockInitSudo()
	testutils.MockWithoutResponse(0, "sudo", "launchctl", "load", launchctlConfig)
	testutils.MockWithoutResponse(0, "sudo", "mkdir", "-p", "-m", "755", path.Join(tmpdir, "bin"))
	testutils.MockWithoutResponse(0, "sudo", "chown", "-R", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), path.Join(tmpdir))
	_ = os.MkdirAll(path.Join(tmpdir, "bin"), 0777)
	i.Install()

	assert.FileExists(t, path.Join(tmpdir, "bin", "coredns"))
	info, e := os.Stat(path.Join(tmpdir, "bin", "coredns"))
	assert.NoError(t, e)
	assert.Equal(t, os.FileMode(0755), info.Mode())
	assert.FileExists(t, path.Join(tmpdir, "etc", "corefile"))
	assert.DirExists(t, path.Join(tmpdir, "var/run"))
	assert.DirExists(t, path.Join(tmpdir, "var/log"))
}

func Test_installer_Uninstall(t *testing.T) {
	ctrl := gomock.NewController(t)
	sh.ExecCommand = testutils.FakeExecCommand
	ghClient := fake.NewMockClient(ctrl)
	tmpdir, e := ioutil.TempDir(os.TempDir(), "coredns_test")

	defer func() {
		defer ctrl.Finish()
		assert.NoError(t, os.RemoveAll(tmpdir))
		sh.ExecCommand = exec.Command
	}()
	assert.NoError(t, e)

	i := &installer{
		ghClient: ghClient,
		prefix:   prefix(tmpdir),
	}
	testutils.MockInitSudo()
	testutils.MockWithoutResponse(0, "sudo", "launchctl", "unload", launchctlConfig)
	testutils.MockWithoutResponse(0, "sudo", "rm", launchctlConfig)
	testutils.MockWithoutResponse(0, "sudo", "rm", dotMinikubeResolverPath)

	i.Uninstall(false)
	_, e = os.Stat(tmpdir)
	if e == nil || !os.IsNotExist(e) {
		t.Errorf("Prefix directory %s should be deleted. But it exists. %s", tmpdir, e)
	}
}

func Test_installer_writeLaunchCtlConfig(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	i := &installer{
		ghClient: nil,
		prefix:   "/prefix/",
	}
	mockWriteFileAsRoot(dotMinikubeResolverPath, []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>de.chrfritz.minikube-support.coredns</string>
		<key>ProgramArguments</key>
		<array>
			<string>/prefix/bin/coredns</string>
			<string>-conf</string>
			<string>/prefix/etc/corefile</string>
			<string>-pidfile</string>
			<string>/prefix/var/run/coredns.pid</string>
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>UserName</key>
		<string>root</string>
		<key>StandardErrorPath</key>
		<string>/prefix/var/log/coredns.error.log</string>
		<key>StandardOutPath</key>
		<string>/prefix/var/log/coredns.log</string>
	</dict>
</plist>
`))

	if err := i.writeLaunchCtlConfig(); err != nil {
		t.Errorf("writeLaunchCtlConfig() error = %v, wantErr false", err)
	}

}

func Test_installer_writeResolverConfig(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	i := &installer{
		ghClient: nil,
		prefix:   "",
	}
	mockWriteFileAsRoot(dotMinikubeResolverPath, []byte("nameserver ::1"))
	if err := i.writeResolverConfig(); err != nil {
		t.Errorf("writeResolverConfig() error = %v, wantErr false", err)
	}
}

func mockWriteFileAsRoot(path string, content []byte) {
	test := testutils.TestProcessResponse{
		Command:           "sudo",
		Args:              []string{"/bin/sh", "-c", shellquote.Join("sed", "-n", "w "+path)},
		ResponseStatus:    0,
		AltResponseStatus: 10,
	}
	if content != nil {
		test.ExpectedStdin = string(content)
	}

	testutils.TestProcessResponses = append(testutils.TestProcessResponses, test)
}
