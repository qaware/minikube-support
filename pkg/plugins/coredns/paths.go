package coredns

import (
	"os"
	"path/filepath"
)

// prefix is a helper type to define all possible paths coredns files and directories at one place.
type prefix string

func newCoreDnsPaths(p string) prefix {
	return prefix(p)
}
func (p prefix) String() string       { return string(p) }
func (p prefix) binDir() string       { return filepath.Join(p.String(), binDir) }
func (p prefix) binary() string       { return filepath.Join(p.String(), binDir, "coredns") }
func (p prefix) etcDir() string       { return filepath.Join(p.String(), etcDir) }
func (p prefix) coreFile() string     { return filepath.Join(p.String(), etcDir, "corefile") }
func (p prefix) runDir() string       { return filepath.Join(p.String(), runDir) }
func (p prefix) pidFile() string      { return filepath.Join(p.String(), runDir, "coredns.pid") }
func (p prefix) logDir() string       { return filepath.Join(p.String(), logDir) }
func (p prefix) logFile() string      { return filepath.Join(p.String(), logDir, "coredns.log") }
func (p prefix) errorLogFile() string { return filepath.Join(p.String(), logDir, "coredns.error.log") }

const binDir = "bin"
const etcDir = "etc"
const runDir = "var" + string(os.PathSeparator) + "run"
const logDir = "var" + string(os.PathSeparator) + "log"
