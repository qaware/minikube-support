package sudos

import (
	"os"
)

// MkdirAll does the same as os.MkdirAll() but it will be executed as sub process with root rights (sudo)
func MkdirAll(path string, mod int) error {
	return os.MkdirAll(path, os.FileMode(mod))
}

// Chown does the same as os.Chown() but it will be executed as sub process with root rights (sudo)
func Chown(path string, uid int, gid int, recursive bool) error {
	return os.Chown(path, uid, gid)
}

// RemoveAll does the same as os.RemoveAll() but it will be executed as sub process with root rights (sudo)
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func WriteFileAsRoot(path string, content []byte) error {
	return os.WriteFile(path, content, os.FileMode(0644))
}
