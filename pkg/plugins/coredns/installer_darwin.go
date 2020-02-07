package coredns

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/kballard/go-shellquote"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
)

const launchctlConfig = "/Library/LaunchDaemons/de.chrfritz.minikube-support.coredns.plist"
const dotMinikubeResolverPath = "/etc/resolver/minikube"

func (i *installer) installSpecific() error {
	e := sh.InitSudo()
	if e != nil {
		return fmt.Errorf("can not init sudo: %s", e)
	}

	e = i.writeLaunchCtlConfig()
	if e != nil {
		return fmt.Errorf("can not write launchctl config: %s", e)
	}
	_, e = sh.RunSudoCmd("launchctl", "load", launchctlConfig)
	if e != nil {
		return fmt.Errorf("can not load coredns launch daemon: %s", e)
	}

	return i.writeResolverConfig()
}

func (i *installer) uninstallSpecific() error {
	_, e := sh.RunSudoCmd("launchctl", "unload", launchctlConfig)
	if e != nil {
		return fmt.Errorf("can not unload coredns launch daemon: %s", e)
	}

	_, e = sh.RunSudoCmd("rm", launchctlConfig)
	if e != nil {
		return fmt.Errorf("can not remove coredns launch daemon config: %s", e)
	}

	_, e = sh.RunSudoCmd("rm", dotMinikubeResolverPath)
	if e != nil {
		return fmt.Errorf("can not remove coredns minikube resolver config: %s", e)
	}
	return nil
}

func (i *installer) writeConfig() error {
	config := `
. {
    reload
    health :8054
    bind 127.0.0.1 
    bind ::1
    log

    grpc minikube 127.0.0.1:8053
}
192.168.64.1:53  {
    forward . /etc/resolv.conf
}
`
	return ioutil.WriteFile(path.Join(i.prefix, "etc", "corefile"), []byte(config), 0644)
}

func (i *installer) writeLaunchCtlConfig() error {
	binary := filepath.Join(i.prefix, "bin", "coredns")
	corefile := filepath.Join(i.prefix, "etc", "corefile")
	pidFile := filepath.Join(i.prefix, "var", "run", "coredns.pid")
	logFile := filepath.Join(i.prefix, "var", "log", "coredns.log")
	errorLogFile := filepath.Join(i.prefix, "var", "log", "coredns.error.log")
	config := `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>de.chrfritz.minikube-support.coredns</string>
		<key>ProgramArguments</key>
		<array>
			<string>` + binary + `</string>
			<string>-conf</string>
			<string>` + corefile + `</string>
			<string>-pidfile</string>
			<string>` + pidFile + `</string>
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>UserName</key>
		<string>root</string>
		<key>StandardErrorPath</key>
		<string>` + errorLogFile + `</string>
		<key>StandardOutPath</key>
		<string>` + logFile + `</string>
	</dict>
</plist>
`

	return i.writeFileAsRoot(launchctlConfig, []byte(config))
}

func (i *installer) writeResolverConfig() error {
	config := "nameserver ::1"
	return i.writeFileAsRoot(dotMinikubeResolverPath, []byte(config))
}

func (i *installer) writeFileAsRoot(path string, content []byte) error {
	command := sh.ExecSudoCommand("/bin/sh", "-c", shellquote.Join("sed", "-n", "w "+path))
	command.Env = append(command.Env, os.Environ()...)
	defer func() {
		if e := command.Wait(); e != nil {
			logrus.Errorf("Unable to wait for writing %s: %s", path, e)
		}
	}()

	writer, e := command.StdinPipe()
	if e != nil {
		return fmt.Errorf("write content into %s failed: %s", path, e)
	}

	_, e = writer.Write(content)
	if e != nil {
		return fmt.Errorf("write content into %s failed: %s", path, e)
	}

	e = command.Start()
	if e != nil {
		return fmt.Errorf("write content into %s failed: %s", path, e)
	}

	e = writer.Close()
	if e != nil {
		return fmt.Errorf("write content into %s failed: %s", path, e)
	}
	return nil
}
