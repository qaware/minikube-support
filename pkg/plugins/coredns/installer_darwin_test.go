// +build darwin

package coredns

import (
	"github.com/kballard/go-shellquote"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"os/exec"
	"testing"
)

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
		ExpectedStdin:     string(content),
		AltResponseStatus: 10,
	}

	testutils.TestProcessResponses = append(testutils.TestProcessResponses, test)
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}
