package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/kudobuilder/kuttl/pkg/test"
)

var (
	errorsExample = `  # Asserts "errors" against a $KUBECONFIG cluster the values defined in the assert file.
  kubectl kuttl errors <path/to/errorsfile.yaml> <path/to/errorsfile.yaml>...`
)

// newErrorsCmd returns a new initialized instance of the errors sub command
func newErrorsCmd() *cobra.Command {
	timeout := 5
	namespace := "default"

	errorsCmd := &cobra.Command{
		Use:     "errors",
		Short:   "Asserts the declared errors state to NOT be true.",
		Long:    `Asserts the declared errors state provided as an argument to not be true in the $KUBECONFIG cluster. Valid arguments are a YAML file, URL to a YAML file.`,
		Example: errorsExample,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("one file argument is required")
			}
			return test.Errors(namespace, timeout, args...)
		},
	}

	errorsCmd.Flags().IntVar(&timeout, "timeout", 5, "The timeout to use as default for error evaluation.")
	errorsCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to use for test errors.")
	return errorsCmd
}
