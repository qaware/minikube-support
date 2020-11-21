package coredns

import (
	"github.com/qaware/minikube-support/pkg/testutils"
	"os"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	testHook()

	testutils.StandardHelperProcess(t)
}
