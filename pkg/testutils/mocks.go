package testutils

// MockInitSudo adds the TestProcessResponse for `sh.initSudo` to `testutils.TestProcessResponses`.
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

	TestProcessResponses = append(TestProcessResponses, test)
}
