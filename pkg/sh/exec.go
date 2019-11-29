package sh

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Alias for os/exec.Command to allow mocking the command executions in tests.
// see https://npf.io/2015/06/testing-exec-command/ for more information about.
var ExecCommand = executeCommand
var sudoAvailable *bool
var sudoAvailableMutex = sync.Mutex{}

// Executes echo with sudo and connected stdin, stdout and stderr to ask for the users password.
// This allows to use sudo for commands which internal process stdout and stderr which are not connected to the current tty.
func InitSudo() error {
	command := ExecSudoCommand("echo", "")
	command.Env = append(command.Env, os.Environ()...)
	return command.Run()
}

// RunCmd executes the given command including the current environment and returns the combined output as string.
func RunCmd(command string, args ...string) (string, error) {
	cmd := ExecCommand(command, args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	bytes, e := cmd.CombinedOutput()
	return string(bytes), e
}

// RunSudoCmd executes the given command including the current environment using sudo and returns the combined output as string.
func RunSudoCmd(command string, args ...string) (string, error) {
	cmd := ExecSudoCommand(command, args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	bytes, e := cmd.CombinedOutput()
	return string(bytes), e
}

// ExecSudoCommand executes the given command in the same way as os.ExecCommand would do it.
// But with the difference that ExecSudoCommand prefixes the command with sudo.
func ExecSudoCommand(command string, args ...string) *exec.Cmd {
	if isSudoAvailable() {
		args = append([]string{command}, args...)
		command = "sudo"
	}
	return ExecCommand(command, args...)
}

// executeCommand is a wrapper for exec.Command() to log the executed command.
func executeCommand(name string, arg ...string) *exec.Cmd {
	logrus.Tracef("Executing command: %s %s", name, strings.Join(arg, " "))
	return exec.Command(name, arg...)
}

// IsExitCode checks the given error for a process exit error and if the exit code is the expected one.
func IsExitCode(err error, exitCode int) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode() == exitCode
	}
	return false
}

// isSudoAvailable checks if the command sudo exists. It returns true if it exists. otherwise false.
func isSudoAvailable() bool {
	if sudoAvailable != nil {
		return *sudoAvailable
	}
	sudoAvailableMutex.Lock()
	defer sudoAvailableMutex.Unlock()
	_, e := RunCmd("which", "sudo")
	available := e == nil
	sudoAvailable = &available
	return *sudoAvailable
}
