package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/kudobuilder/kuttl/pkg/test"
)

var (
	assertExample = `  # Asserts against a $KUBECONFIG cluster the values defined in the assert file.
  kubectl kuttl assert <path/to/assertfile.yaml>`
)

// newAssertCmd returns a new initialized instance of the assert sub command
func newAssertCmd() *cobra.Command {
	timeout := 5
	namespace := "default"

	assertCmd := &cobra.Command{
		Use:     "assert",
		Short:   "Asserts the declared state to be true.",
		Long:    `Asserts the declared state provided as an argument to be true in the $KUBECONFIG cluster. Valid arguments are a YAML file, URL to a YAML file.`,
		Example: assertExample,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("one file argument is required")
			}
			return test.Assert(namespace, timeout, args...)
		},
	}

	assertCmd.Flags().IntVar(&timeout, "timeout", 5, "The timeout to use as default for TestSuite configuration.")
	assertCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to use for test assert.")

	return assertCmd
}
