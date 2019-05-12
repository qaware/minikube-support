package sh

import (
	"os"
	"os/exec"
)

// Alias for os/exec.Command to allow mocking the command executions in tests.
// see https://npf.io/2015/06/testing-exec-command/ for more information about.
var ExecCommand = exec.Command

// Executes echo with sudo and connected stdin, stdout and stderr to ask for the users password.
// This allows to use sudo for commands which internal process stdout and stderr which are not connected to the current tty.
func InitSudo() error {
	command := ExecCommand("sudo", "echo", "")
	command.Env = os.Environ()
	return command.Run()
}
