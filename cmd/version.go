package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qaware/minikube-support/version"
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
	parsedDate, _ := strconv.ParseInt(version.CommitDate, 10, 64)
	commitDate := time.Unix(parsedDate, 0).Format("2006-01-02 15:04:05 Z07:00")
	fmt.Printf(`Minikube-Support Tools

Version:     %s
Commit:      %s
Commit Date: %s
Branch:      %s
`,
		version.Version,
		version.Revision,
		commitDate,
		version.Branch,
	)
}
