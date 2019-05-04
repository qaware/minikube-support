package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/logging"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func init() {
	glog.V(0)
	loggerConfig := logging.InitFlags(rootCmd.PersistentFlags())
	cobra.OnInitialize(loggerConfig.Initialize)
}
