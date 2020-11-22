package testutils

// MockInitSudo adds the TestProcessResponse for `sh.initSudo` to `testutils.testProcessResponses`.
func MockInitSudo() {
	MockWithoutResponse(0, "which", "sudo")
	MockWithoutResponse(0, "sudo", "echo", "")
}

// MockWithoutResponse adds a simple TestProcessResponse.
func MockWithoutResponse(returnStatus int, cmd string, args ...string) {
	test := TestProcessResponse{
		Command:        cmd,
		Args:           args,
		ResponseStatus: returnStatus,
	}

	AddTestProcessResponse(test)
}

func MockWithStdOut(stdOut string, returnStatus int, cmd string, args ...string) {
	test := TestProcessResponse{
		Command:        cmd,
		Args:           args,
		ResponseStatus: returnStatus,
		Stdout:         stdOut,
	}

	AddTestProcessResponse(test)
}
