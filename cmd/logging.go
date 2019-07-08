package cmd

import (
	"github.com/golang/glog"
	"github.com/qaware/minikube-support/pkg/logging"
	"github.com/spf13/cobra"
)

func init() {
	glog.V(0)
	loggerConfig := logging.InitFlags(rootCmd.PersistentFlags())
	cobra.OnInitialize(loggerConfig.Initialize)
}
