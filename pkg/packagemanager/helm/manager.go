package helm

import (
	"github.com/pkg/errors"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/sh"
	"strings"
)

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=manager.go

type Manager interface {
	Init() error
	AddRepository(name string, url string) error
	UpdateRepository() error
	Install(chart string, release string, namespace string, values map[string]interface{}, wait bool)
	Uninstall(release string, purge bool)
}

func NewHelmManager(context kubernetes.ContextHandler) (Manager, error) {
	version, e := getHelmVersion()
	if e != nil {
		return nil, e
	}

	if strings.HasPrefix(version, "v3.") {
		return NewHelm3Manager(context), nil
	} else if strings.HasPrefix(version, "Client: v2.") {
		return NewHelm2Manager(context), nil
	}
	return nil, errors.New("can not determ helm version")
}

func getHelmVersion() (string, error) {
	response, e := sh.RunCmd("helm", "version", "-c", "--short")
	if e != nil {
		return "", errors.Wrap(e, "can not check helm version")
	}
	return response, nil
}
