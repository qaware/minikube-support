package testutils

import (
	"fmt"
	"os"
	"os/exec"
)

// This command fakes the exec command. This should be only used in Tests.
func FakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// Extract the command name and arguments of the mocked exec call using FakeExecCommand()
func ExtractMockedCommandAndArgs() (string, []string) {
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}
	return args[0], args[1:]
}
