package sh

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// Alias for os/exec.Command to allow mocking the command executions in tests.
// see https://npf.io/2015/06/testing-exec-command/ for more information about.
var ExecCommand = executeCommand

// Executes echo with sudo and connected stdin, stdout and stderr to ask for the users password.
// This allows to use sudo for commands which internal process stdout and stderr which are not connected to the current tty.
func InitSudo() error {
	command := ExecCommand("sudo", "echo", "")
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

// executeCommand is a wrapper for exec.Command() to log the executed command.
func executeCommand(name string, arg ...string) *exec.Cmd {
	logrus.Tracef("Executing command: %s %s", name, strings.Join(arg, " "))
	return exec.Command(name, arg...)
}
