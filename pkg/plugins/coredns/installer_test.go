package coredns

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qaware/minikube-support/pkg/github/fake"
	"github.com/stretchr/testify/assert"
)

func Test_installer_downloadCoreDns(t *testing.T) {
	tests := []struct {
		name          string
		versionError  bool
		downloadError bool
		createBaseDir bool
		wantErr       bool
	}{
		{"ok", false, false, true, false},
		{"version error", true, false, false, true},
		{"download error", false, true, false, true},
		{"no basedir", false, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ghClient := fake.NewMockClient(ctrl)
			tmpdir, e := os.MkdirTemp(os.TempDir(), "coredns_test")
			assetName := fmt.Sprintf("coredns_1.0.0_%s_%s.tgz", runtime.GOOS, runtime.GOARCH)
			assert.NoError(t, e)
			defer func() {
				defer ctrl.Finish()
				assert.NoError(t, os.RemoveAll(tmpdir))
			}()

			i := &installer{ghClient: ghClient, prefix: prefix(tmpdir)}

			if !tt.versionError {
				ghClient.EXPECT().GetLatestReleaseTag("coredns", "coredns").
					Return("v1.0.0", nil)
			} else {
				ghClient.EXPECT().GetLatestReleaseTag("coredns", "coredns").
					Return("", fmt.Errorf(""))
			}

			if !tt.downloadError {
				ghClient.EXPECT().
					DownloadReleaseAsset("coredns", "coredns", "v1.0.0", assetName).
					AnyTimes().
					Return(os.Open("fixtures/coredns.tar.gz"))
			} else {
				ghClient.EXPECT().
					DownloadReleaseAsset("coredns", "coredns", "v1.0.0", assetName).
					AnyTimes().
					Return(nil, fmt.Errorf(""))
			}
			if tt.createBaseDir {
				assert.NoError(t, os.MkdirAll(path.Join(tmpdir, "bin"), os.FileMode(0755)))
			}
			if err := i.downloadCoreDns(); (err != nil) != tt.wantErr {
				t.Errorf("downloadCoreDns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
