package packagemanager

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/qaware/minikube-support/pkg/packagemanager/fake"
	"reflect"
	"testing"
)

func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		name      string
		init      bool
		want      PackageManager
		wantPanic bool
	}{
		{"no osPackageManager", false, nil, true},
		{"a osPackageManager", true, &DummyManager{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("Not expected panic. Want %v got %v", tt.wantPanic, r)
				}
			}()

			if tt.init {
				osPackageManager = &DummyManager{}
			} else {
				osPackageManager = nil
			}
			if got := GetPackageManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPackageManager() = %v, want %v", got, tt.want)
			}
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

type DummyManager struct{}

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
