// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package sudos

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qaware/minikube-support/pkg/sh"
)

// MkdirAll does the same as os.MkdirAll() but it will be executed as sub process with root rights (sudo)
func MkdirAll(path string, mod int) error {
	resp, e := sh.RunSudoCmd("mkdir", "-p", "-m", strconv.FormatInt(int64(mod), 8), path)
	if e != nil {
		return errors.Wrapf(e, "can not create directory %s: %s", path, resp)
	}
	return nil
}

// Chown does the same as os.Chown() but it will be executed as sub process with root rights (sudo)
func Chown(path string, uid int, gid int, recursive bool) error {
	var args []string
	if recursive {
		args = []string{"-R"}
	}
	args = append(args, fmt.Sprintf("%d:%d", uid, gid), path)

	s, e := sh.RunSudoCmd("chown", args...)
	if e != nil {
		return errors.Wrapf(e, "can not set owner for %s to current user:group (%d:%d): %s", path, uid, gid, s)
	}
	return nil
}

// RemoveAll does the same as os.RemoveAll() but it will be executed as sub process with root rights (sudo)
func RemoveAll(path string) error {
	resp, e := sh.RunSudoCmd("rm", "-R", path)
	if e != nil {
		return errors.Wrapf(e, "can not remove %s recursive: %s", path, resp)
	}
	return nil
}
