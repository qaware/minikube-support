package packagemanager

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// PackageManager is a simple abstraction about the different package managers for the different operating systems.
// For example on macOS we use brew to mange local 3rd party tools.
type PackageManager interface {
	// Get the name of the package manager
	fmt.Stringer

	// Installs the given package with this package manager
	Install(pkg string) error

	// Updates the given package with this package manager
	Update(pkg string) error

	// Uninstalls the given package with this package manager
	Uninstall(pkg string) error
}

// The singleton package manager instance. It depends on the current operating system.
var manager PackageManager

// Get the os specific package manager.
func GetPackageManager() PackageManager {
	if manager == nil {
		logrus.Panicf("Can not retrieve package manager.")
	}
	return manager
}
