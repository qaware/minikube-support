// +build aix dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package coredns

func (i *installer) installSpecific() error   { return nil }
func (i *installer) uninstallSpecific() error { return nil }
func (i *installer) writeConfig() error       { return nil }
