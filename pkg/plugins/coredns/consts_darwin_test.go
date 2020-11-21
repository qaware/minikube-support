// +build darwin

package coredns

import (
	"github.com/qaware/minikube-support/pkg/testutils"
	"os"
	"strings"
)

func testHook() {
	// adds mock for remove all in uninstall (no sudo!)
	cmd, args := testutils.ExtractMockedCommandAndArgs()
	if cmd == "sudo" && len(args) == 3 && args[0] == "rm" && args[1] == "-R" {
		p := args[2]
		expected := os.TempDir()
		if strings.HasPrefix(p, expected) {
			_ = os.RemoveAll(p)
			os.Exit(0)
			return
		}
		os.Exit(1)
	}
}
