package packagemanager

import (
	"container/heap"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/qaware/minikube-support/pkg/packagemanager/fake"
)

func Test_findOsPackageManager(t *testing.T) {
	tests := []struct {
		name               string
		registeredManagers []osSpecific
		wantPanic          bool
		wantManager        osSpecific
	}{
		{"empty", []osSpecific{}, true, nil},
		{"one not available", []osSpecific{DummyManager{available: false}}, true, nil},
		{"one available", []osSpecific{DummyManager{available: true}}, false, DummyManager{available: true}},
		{"one available one not", []osSpecific{DummyManager{available: true}, DummyManager{available: false}}, false, DummyManager{available: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managerQueue = queue{}
			heap.Init(&managerQueue)
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("Not expected panic. Want %v got %v", tt.wantPanic, r)
				}
			}()
			for _, manager := range tt.registeredManagers {
				RegisterManager(manager, 1)
			}

			findOsPackageManager()

			assert.Equal(t, tt.wantManager, osPackageManager)
		})
	}
}

func TestInstallOrUpdate(t *testing.T) {
	tests := []struct {
		name           string
		pkg            string
		isInstalled    bool
		isInstalledErr error
		wantInstalled  bool
		wantUpdate     bool
		wantErr        bool
	}{
		{"install", "dummy", false, nil, true, false, false},
		{"update", "dummy", true, nil, false, true, false},
		{"check error", "dummy", false, errors.New("dummy"), false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := fake.NewMockPackageManager(ctrl)
			SetOsPackageManager(manager)
			manager.EXPECT().IsInstalled(tt.pkg).Return(tt.isInstalled, tt.isInstalledErr)
			if tt.wantInstalled {
				manager.EXPECT().Install(tt.pkg)
			}
			if tt.wantUpdate {
				manager.EXPECT().Update(tt.pkg)
			}

			if err := InstallOrUpdate(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("InstallOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type DummyManager struct {
	available bool
}

func (DummyManager) String() string {
	return "dummy"
}
func (DummyManager) Install(pkg string) error {
	panic("implement me")
}
func (DummyManager) Update(pkg string) error {
	panic("implement me")
}
func (DummyManager) Uninstall(pkg string) error {
	panic("implement me")
}
func (DummyManager) IsInstalled(pkg string) (bool, error) {
	panic("implement me")
}
func (m DummyManager) IsAvailable() bool {
	return m.available
}
