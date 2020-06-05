package cmd

import (
	"testing"

	"github.com/spf13/cobra"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/test"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

var (
	assertExample = `  # Asserts against a $KUBECONFIG cluster the values defined in the assert file.
  kubectl kuttl assert <path/to/assertfile.yaml>`
)

// newAssertCmd returns a new initialized instance of the assert sub command
func newAssertCmd() *cobra.Command {
	timeout := 30
	options := harness.TestSuite{}

	assertCmd := &cobra.Command{
		Use:     "assert",
		Short:   "Asserts the declared state to be true.",
		Long:    `Asserts the declared state provided as an argument to be true in the $KUBECONFIG cluster. Valid arguments are a YAML file, URL to a YAML file.`,
		Example: assertExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			options.TestDirs = args

			if isSet(flags, "timeout") {
				options.Timeout = timeout
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			testutils.RunTests("kuttl", "assert-cmd", 1, func(t *testing.T) {
				harness := test.Harness{
					TestSuite: options,
					T:         t,
				}
				//step.goL349
				harness.Run()
			})
		},
	}

	assertCmd.Flags().IntVar(&timeout, "timeout", 30, "The timeout to use as default for TestSuite configuration.")

	return assertCmd
}
