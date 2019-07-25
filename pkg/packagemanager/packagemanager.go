package packagemanager

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=packagemanager.go

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// PackageManager is a simple abstraction about the different package managers for the different operating systems.
// For example on macOS we use brew to mange local 3rd party tools.
type PackageManager interface {
	// Get the name of the package osPackageManager
	fmt.Stringer

	// Installs the given package with this package osPackageManager
	Install(pkg string) error

	// Updates the given package with this package osPackageManager
	Update(pkg string) error

	// Uninstalls the given package with this package osPackageManager
	Uninstall(pkg string) error
}

// The singleton package osPackageManager instance. It depends on the current operating system.
var osPackageManager PackageManager

// Get the os specific package osPackageManager.
func GetPackageManager() PackageManager {
	if osPackageManager == nil {
		logrus.Panicf("Can not retrieve package osPackageManager.")
	}
	return osPackageManager
}

func SetOsPackageManager(manager PackageManager) {
	if manager != nil {
		osPackageManager = manager
	}
}
