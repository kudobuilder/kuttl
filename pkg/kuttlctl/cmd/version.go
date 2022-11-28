package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyverno/kuttl/pkg/version"
)

var (
	versionExample = `  # Print the current installed KUTTL package version
  kubectl kuttl version`
)

// newVersionCmd returns a new initialized instance of the version sub command
func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the current KUTTL package version.",
		Long:    `Print the current installed KUTTL package version.`,
		Example: versionExample,
		RunE:    VersionCmd,
	}

	return versionCmd
}

// VersionCmd performs the version sub command
func VersionCmd(cmd *cobra.Command, args []string) error {
	kuttlVersion := version.Get()
	fmt.Printf("KUTTL Version: %s\n", fmt.Sprintf("%#v", kuttlVersion))
	return nil
}
