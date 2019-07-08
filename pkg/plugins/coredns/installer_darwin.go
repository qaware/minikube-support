package coredns

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/qaware/minikube-support/pkg/sh"
	"os"
)

const launcctlConfig = "/Library/LaunchDaemons/de.chrfritz.minikube-support.coredns.plist"

func (i *installer) installSpecific() error {
	e := sh.InitSudo()
	if e != nil {
		return fmt.Errorf("can not init sudo: %s", e)
	}

	e = i.writeLaunchCtlConfig()
	if e != nil {
		return fmt.Errorf("can not write launchctl config: %s", e)
	}
	_, e = sh.RunCmd("sudo", "launchctl", "load", launcctlConfig)
	if e != nil {
		return fmt.Errorf("can not load coredns launch daemon: %s", e)
	}
	return nil
}
func (i *installer) updateSpecific() error {

	return nil
}
func (i *installer) uninstallSpecific() error {
	_, e := sh.RunCmd("sudo", "launchctl", "unload", launcctlConfig)
	if e != nil {
		return fmt.Errorf("can not unload coredns launch daemon: %s", e)
	}
	_, e = sh.RunCmd("sudo", "rm", launcctlConfig)
	if e != nil {
		return fmt.Errorf("can not remove coredns launch daemon config: %s", e)
	}
	return nil
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
			<string>` + i.prefix + `bin/coredns</string>
			<string>-conf</string>
			<string>` + i.prefix + `etc/corefile</string>
			<string>-pidfile</string>
			<string>` + i.prefix + `var/run/coredns.pid</string>
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>UserName</key>
		<string>root</string>
		<key>StandardErrorPath</key>
		<string>` + i.prefix + `var/log/coredns.error.log</string>
		<key>StandardOutPath</key>
		<string>` + i.prefix + `var/log/coredns.log</string>
	</dict>
</plist>
`

	command := sh.ExecCommand("sudo", "/bin/sh", "-c", shellquote.Join("sed", "-n", "w "+launcctlConfig))
	command.Env = append(command.Env, os.Environ()...)
	defer command.Wait()
	writer, e := command.StdinPipe()
	if e != nil {
		return fmt.Errorf("write launchctl config failed: %s", e)
	}
	_, e = writer.Write([]byte(config))
	if e != nil {
		return fmt.Errorf("write launchctl config failed: %s", e)
	}
	e = command.Start()
	if e != nil {
		return fmt.Errorf("write launchctl config failed: %s", e)
	}

	e = writer.Close()

	if e != nil {
		return fmt.Errorf("write launchctl config failed: %s", e)
	}
	return nil
}
