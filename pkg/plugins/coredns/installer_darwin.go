package coredns

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/utils/sudos"
)

const launchctlConfig = "/Library/LaunchDaemons/de.chrfritz.minikube-support.coredns.plist"
const dotMinikubeResolverPath = "/etc/resolver/minikube"

func (i *installer) installSpecific() error {
	e := sh.InitSudo()
	if e != nil {
		return fmt.Errorf("can not init sudo: %s", e)
	}

	err := i.setupLaunchCtrl()
	if err != nil {
		return err
	}

	return i.writeResolverConfig()
}

// setupLaunchCtrl setups the launch daemon configuration and loads them using the macOS util launchctl.
func (i *installer) setupLaunchCtrl() error {
	if !runAsDaemon {
		return nil
	}

	e := i.writeLaunchCtlConfig()
	if e != nil {
		return fmt.Errorf("can not write launchctl config: %s", e)
	}
	_, e = sh.RunSudoCmd("launchctl", "load", launchctlConfig)
	if e != nil {
		return fmt.Errorf("can not load coredns launch daemon: %s", e)
	}
	return nil
}

func (i *installer) uninstallSpecific() error {
	_, e := sh.RunSudoCmd("launchctl", "unload", launchctlConfig)
	if e != nil {
		logrus.Debugf("can not unload coredns launch daemon: %s", e)
	}

	_, e = sh.RunSudoCmd("rm", launchctlConfig)
	if e != nil {
		logrus.Debugf("can not remove coredns launch daemon config: %s", e)
		return nil
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
	return os.WriteFile(i.prefix.coreFile(), []byte(config), 0644)
}

func (i *installer) writeLaunchCtlConfig() error {
	config := `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>de.chrfritz.minikube-support.coredns</string>
		<key>ProgramArguments</key>
		<array>
			<string>` + i.prefix.binary() + `</string>
			<string>-conf</string>
			<string>` + i.prefix.coreFile() + `</string>
			<string>-pidfile</string>
			<string>` + i.prefix.pidFile() + `</string>
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>UserName</key>
		<string>root</string>
		<key>StandardErrorPath</key>
		<string>` + i.prefix.errorLogFile() + `</string>
		<key>StandardOutPath</key>
		<string>` + i.prefix.logFile() + `</string>
	</dict>
</plist>
`

	return sudos.WriteFileAsRoot(launchctlConfig, []byte(config))
}

func (i *installer) writeResolverConfig() error {
	config := "nameserver ::1"
	return sudos.WriteFileAsRoot(dotMinikubeResolverPath, []byte(config))
}
