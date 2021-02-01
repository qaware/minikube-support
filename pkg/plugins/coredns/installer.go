package coredns

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/utils/sudos"
)

type installer struct {
	ghClient github.Client
	prefix   prefix
}

const PluginName = "coredns"

func NewInstaller(prefix string, ghClient github.Client) apis.InstallablePlugin {
	return &installer{
		ghClient: ghClient,
		prefix:   newCoreDnsPaths(prefix),
	}
}

func (i *installer) String() string {
	return PluginName
}

func (i *installer) Install() {
	var errs *multierror.Error
	errs = multierror.Append(errs, sudos.MkdirAll(i.prefix.binDir(), 0755))
	errs = multierror.Append(errs, sudos.Chown(i.prefix.String(), os.Getuid(), os.Getgid(), true))

	errs = multierror.Append(errs, os.MkdirAll(i.prefix.etcDir(), 0755))
	errs = multierror.Append(errs, os.MkdirAll(i.prefix.runDir(), 0755))
	errs = multierror.Append(errs, os.MkdirAll(i.prefix.logDir(), 0755))
	errs = multierror.Append(errs, i.writeConfig())

	errs = multierror.Append(errs, i.downloadCoreDns())

	errs = multierror.Append(errs, i.installSpecific())

	if errs.Len() > 0 {
		logrus.Errorf("Unable to install coredns into %s:\n  Errors: %s", i.prefix, errs)
	}
}

func (i *installer) Update() {
	i.Uninstall(false)
	i.Install()
}

func (i *installer) Uninstall(_ bool) {
	var errs *multierror.Error

	errs = multierror.Append(errs, i.uninstallSpecific())
	errs = multierror.Append(errs, sudos.RemoveAll(i.prefix.String()))
	if errs.Len() > 0 {
		logrus.Errorf("Unable to uninstall coredns from %s:\n  Errors: %s", i.prefix, errs)
	}
}

func (i *installer) Phase() apis.Phase {
	return apis.LOCAL_TOOLS_CONFIG
}

func (i *installer) downloadCoreDns() error {
	tagName, e := i.ghClient.GetLatestReleaseTag("coredns", "coredns")
	if e != nil {
		return fmt.Errorf("can not get latest coredns version: %s", e)
	}
	version := strings.TrimPrefix(tagName, "v")

	assetName := fmt.Sprintf("coredns_%s_%s_%s.tgz", version, runtime.GOOS, runtime.GOARCH)
	bytes, e := i.ghClient.DownloadReleaseAsset("coredns", "coredns", tagName, assetName)
	if e != nil {
		return fmt.Errorf("can not download coredns binary: %s", e)
	}

	gzReader, e := gzip.NewReader(bytes)
	if e != nil {
		return fmt.Errorf("can not open gz reader for downloaded coredns: %s", e)
	}
	tarReader := tar.NewReader(gzReader)

	for {
		header, e := tarReader.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return fmt.Errorf("unable to extract next file from tar: %s", e)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		name := header.Name
		file, e := os.OpenFile(i.prefix.binary(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if e != nil {
			return fmt.Errorf("can not write file %s: %s", name, e)
		}

		_, e = io.Copy(file, tarReader)
		if e != nil {
			return fmt.Errorf("can not write file (%s) content: %s", name, e)
		}
	}
	return nil
}
