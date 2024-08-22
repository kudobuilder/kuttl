package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kudobuilder/kuttl/pkg/k8s"
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

  # Run kuttl tests with an xml report
  kubectl kuttl test --report xml

  # Test 1 assertion file against a cluster
  kubectl kuttl assert ../01-assert.yaml

  # View kuttl version
  kubectl kuttl version
`,
		Version: version.Get().GitVersion,
	}

	cmd.PersistentFlags().StringVar(&k8s.ImpersonateAs, "as", "", "Username to impersonate for the operation. User could be a regular user or a service account in a namespace.")
	cmd.AddCommand(newAssertCmd())
	cmd.AddCommand(newErrorsCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}
