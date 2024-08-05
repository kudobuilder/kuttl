package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kudobuilder/kuttl/pkg/version"
)

// NewKuttlCmd creates a new root command for kuttlctl
func NewKuttlCmd() *cobra.Command {
	configFlags := genericclioptions.NewConfigFlags(true)

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

	cmd.AddCommand(newAssertCmd())
	cmd.AddCommand(newErrorsCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newVersionCmd())
	configFlags.AddFlags(cmd.PersistentFlags())

	return cmd
}
