package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kudobuilder/kuttl/pkg/version"
)

// NewKuttlCmd creates a new root command for kuttlctl
func NewKuttlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl-kuttl",
		Short: "CLI to Test Kubernetes",
		Long: `KUTTL CLI and future sub-commands can be used to manipulate, inspect and troubleshoot CRDs
and serves as an API aggregation layer.
`,
		SilenceUsage: true,
		Example: `  # Run integration tests against a Kubernetes cluster or mocked control plane.
  kubectl kuttl test

  # View kuttl version
  kubectl kuttl version
`,
		Version: version.Get().GitVersion,
	}

	cmd.AddCommand(newAssertCmd())
	cmd.AddCommand(newErrorsCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}
