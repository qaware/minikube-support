// +build aix dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package coredns

import "io/ioutil"

func (i *installer) installSpecific() error {
	// nothing to do at the moment
	return nil
}
func (i *installer) uninstallSpecific() error {
	// nothing to do at the moment
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
	return ioutil.WriteFile(i.prefix.coreFile(), []byte(config), 0644)
}
