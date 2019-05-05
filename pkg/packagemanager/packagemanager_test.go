package packagemanager

import (
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
