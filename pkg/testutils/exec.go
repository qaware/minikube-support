package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

// This command fakes the exec command. This should be only used in Tests.
func FakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	if len(TestProcessResponses) > 0 {
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		_ = encoder.Encode(TestProcessResponses)
		cmd.Env = append(cmd.Env, "RESPONSE="+buf.String())
	}
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

func StandardHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	strResponse := os.Getenv("RESPONSE")
	var responses []TestProcessResponse
	if strResponse != "" {
		decoder := json.NewDecoder(bytes.NewBufferString(strResponse))
		_ = decoder.Decode(&responses)
	}

	executedCommand, executedArgs := ExtractMockedCommandAndArgs()

	response := FindTestProcessResponse(responses, executedCommand, executedArgs)
	status := response.ResponseStatus

	if response.ExpectedStdin != "" {
		buffer := make([]byte, len(response.ExpectedStdin)+10)
		_, _ = os.Stdin.Read(buffer)
		if !reflect.DeepEqual(buffer, []byte(response.ExpectedStdin)) {
			status = response.AltResponseStatus
		}
	}

	defer os.Exit(status)
	_, _ = fmt.Fprint(os.Stderr, response.Stderr)
	_, _ = fmt.Fprint(os.Stdout, response.Stdout)
	time.Sleep(response.delay)
}

func FindTestProcessResponse(responses []TestProcessResponse, cmd string, args []string) *TestProcessResponse {
	for _, response := range responses {
		if response.Command != cmd {
			continue
		}
		if !reflect.DeepEqual(response.Args, args) {
			continue
		}
		return &response
	}
	return &TestProcessResponse{
		Stderr:         fmt.Sprintf("No matching response found for command line: \"%s %s\"", cmd, args),
		ResponseStatus: -2,
	}
}

type TestProcessResponse struct {
	Command        string
	Args           []string
	ResponseStatus int
	Stdout         string
	Stderr         string
	ExpectedStdin  string
	delay          time.Duration
	// AltResponseStatus will be returned if stdin got something else as expected
	AltResponseStatus int
}

var TestProcessResponses []TestProcessResponse
