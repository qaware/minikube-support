package cmd

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/version"
	"github.com/spf13/cobra"
)

type VersionOptions struct{}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func NewVersionCommand() *cobra.Command {
	versionOptions := NewVersionOptions()

	return &cobra.Command{
		Use:   "version",
		Short: "Show the version information",
		Run:   versionOptions.run,
	}
}

func (v *VersionOptions) run(cmd *cobra.Command, args []string) {
	fmt.Printf("Version:\t%s@%s on %s\nBuild Date:\t%s\nGo Version:\t%s\n",
		version.Version,
		version.Revision,
		version.Branch,
		version.BuildDate,
		version.GoVersion)
}
