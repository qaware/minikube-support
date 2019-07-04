package coredns

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/github"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

type installer struct {
	ghClient github.Client
	prefix   string
}

const PluginName = "coredns"

func NewInstaller(prefix string) apis.InstallablePlugin {
	return &installer{
		ghClient: github.NewClient(""),
		prefix:   prefix,
	}
}

func (i *installer) String() string {
	return PluginName
}

func (i *installer) Install() {
	var errs *multierror.Error
	errs = multierror.Append(errs, os.MkdirAll(path.Join(i.prefix, "bin"), 0755))
	errs = multierror.Append(errs, os.MkdirAll(path.Join(i.prefix, "etc"), 0755))
	errs = multierror.Append(errs, os.MkdirAll(path.Join(i.prefix, "var", "run"), 0777))
	errs = multierror.Append(errs, os.MkdirAll(path.Join(i.prefix, "var", "log"), 0777))
	errs = multierror.Append(errs, i.writeConfig())

	errs = multierror.Append(errs, i.downloadCoreDns())

	errs = multierror.Append(errs, i.installSpecific())

	if errs.Len() > 0 {
		logrus.Errorf("Unable to install coredns into %s:\n  Errors: %s", i.prefix, errs)
	}
}

func (i *installer) Update() {

}

func (i *installer) Uninstall(purge bool) {
	var errs *multierror.Error

	errs = multierror.Append(errs, i.uninstallSpecific())
	errs = multierror.Append(errs, os.RemoveAll(i.prefix))
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

		if header.Typeflag == tar.TypeReg {
			name := header.Name

			file, e := os.OpenFile(path.Join(i.prefix, "bin", name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if e != nil {
				return fmt.Errorf("can not write file %s: %s", name, e)
			}

			//e = file.Chown(header.Uid, header.Gid)
			//if e != nil {
			//	return fmt.Errorf("can not set uid, gid for file %s: %e", name, e)
			//}

			_, e = io.Copy(file, tarReader)
			if e != nil {
				return fmt.Errorf("can not write file (%s) content: %s", name, e)
			}
		}
	}
	return nil
}

func (i *installer) writeConfig() error {
	config := `
. {
    reload
    health :8054
    bind 127.0.0.1 
    bind ::1
    log

    grpc minikube 127.0.0.1:8053
}
`
	return ioutil.WriteFile(path.Join(i.prefix, "etc", "corefile"), []byte(config), 0644)
}

// CONFIG_FILE=$HOME/Development/chr-fritz/dev-tools/config/coredns/coredns.conf
//PREFIX=/opt/coredns
//PID_FILE=$PREFIX/var/run/coredns.run
//VERSION=1.5.0
//DOWNLOAD_LOCATION=https://github.com/coredns/coredns/releases/download/v${VERSION}/coredns_${VERSION}_darwin_amd64.tgz
//KUBE_CONFIG=$HOME/.kube/config
