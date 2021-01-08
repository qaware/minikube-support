package packagemanager

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=packagemanager.go

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// PackageManager is a simple abstraction about the different package managers for the different operating systems.
// For example on macOS we use brew to mange local 3rd party tools.
type PackageManager interface {
	// Get the name of the package osPackageManager
	fmt.Stringer

	// Installs the given package with this package osPackageManager
	Install(pkg string) error

	// IsInstalled checks if the given package is already installed.
	IsInstalled(pkg string) (bool, error)

	// Updates the given package with this package osPackageManager
	Update(pkg string) error

	// Uninstalls the given package with this package osPackageManager
	Uninstall(pkg string) error
}

type osSpecific interface {
	PackageManager
	IsAvailable() bool
}

var findMutex = sync.Mutex{}
var managerQueue queue

// The singleton package osPackageManager instance. It depends on the current operating system.
var osPackageManager PackageManager

func init() {
	heap.Init(&managerQueue)
}

// GetPackageManager returns the os specific package osPackageManager.
func GetPackageManager() PackageManager {
	if osPackageManager == nil {
		findOsPackageManager()
	}
	return osPackageManager
}

// SetOsPackageManager sets the os specific package manager. It is mainly for testing purposes.
func SetOsPackageManager(manager PackageManager) {
	if manager != nil {
		osPackageManager = manager
	}
}

// RegisterManager registers all package managers which can be possible on the current operating system.
func RegisterManager(manager osSpecific, priority int) {
	item := &item{
		manager:  manager,
		priority: priority,
	}
	heap.Push(&managerQueue, item)
}

// InstallOrUpdate either installs or update the given package. Depending on if the package is already installed or not.
func InstallOrUpdate(pkg string) error {
	manager := GetPackageManager()
	installed, e := manager.IsInstalled(pkg)
	if e != nil {
		return e
	}
	if installed {
		return manager.Update(pkg)
	} else {
		return manager.Install(pkg)
	}
}

// findOsPackageManager tries to find the appropriate package manager for the current operating system.
// It will panic if there is no valid package manager.
func findOsPackageManager() {
	if osPackageManager != nil {
		return
	}
	findMutex.Lock()
	defer findMutex.Unlock()
	logrus.Info("Trying to find system package manager")
	for managerQueue.Len() > 0 {
		item := heap.Pop(&managerQueue).(*item)
		logrus.Debugf("Trying %s", item.manager.String())
		if item.manager.IsAvailable() {
			osPackageManager = item.manager
			return
		}
	}
	panic("Can not find any system package manager. Please refer the documentation how to install one for your operating system.")
}

// SelfInstalledUsingPackageManager checks if the minikube-support tools were
// installed using the same PackageManager returned by GetPackageManager().
func SelfInstalledUsingPackageManager() bool {
	installed, err := GetPackageManager().IsInstalled("minikube-support")
	if err != nil {
		panic("Can not check if the minikube-support command was installed using the os pacakge manager.")
	}
	return installed
}
