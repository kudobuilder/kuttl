package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/report"
	"github.com/kudobuilder/kuttl/pkg/test"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

var (
	testExample = `  Run tests configured by kuttl-test.yaml:
    kubectl kuttl test

  Load a specific test configuration:
    kubectl kuttl test --config test.yaml

  Run tests against an existing Kubernetes cluster:
    kubectl kuttl test ./test/integration/

  Run tests against an existing Kubernetes cluster, and install manifests, and CRDs for the tests:
    kubectl kuttl test --crd-dir ./config/crds/ --manifests-dir ./test/manifests/ ./test/integration/

  Run a Kubernetes control plane and install manifests and CRDs for the running tests:
    kubectl kuttl test --start-control-plane  --crd-dir ./config/crds/ --manifests-dir ./test/manifests/ ./test/integration/

  Run tests against an existing Kubernetes cluster with a JUnit XML file output:
    kubectl kuttl test ./test/integration/ --report xml
`
)

// newTestCmd creates the test command for the CLI
// nolint:gocyclo
func newTestCmd() *cobra.Command {
	configPath := ""
	crdDir := ""
	manifestDirs := []string{}
	testToRun := ""
	startControlPlane := false
	attachControlPlaneOutput := false
	startKIND := false
	kindConfig := ""
	kindContext := ""
	skipDelete := false
	skipClusterDelete := false
	parallel := 0
	artifactsDir := ""
	mockControllerFile := ""
	timeout := 30
	reportFormat := ""
	namespace := ""
	suppress := []string{}

	options := harness.TestSuite{}

	testCmd := &cobra.Command{
		Use:   "test [flags]... [test directories]...",
		Short: "Test KUTTL and Operators.",
		Long: `Runs integration tests against a Kubernetes cluster.

The test operator supports connecting to an existing Kubernetes cluster or it can start a Kubernetes API server during the test run.
It can also apply manifests before running the tests. If no arguments are provided, the test harness will attempt to
load the test configuration from kuttl-test.yaml.

For more detailed documentation, visit: https://kuttl.dev`,
		Example: testExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			// If a config is not set and kuttl-test.yaml exists, set configPath to kuttl-test.yaml.
			if configPath == "" {
				if _, err := os.Stat("kuttl-test.yaml"); err == nil {
					configPath = "kuttl-test.yaml"
				} else {
					log.Println("running without a 'kuttl-test.yaml' configuration")
				}
			}

			// Load the configuration YAML into options.
			if configPath != "" {
				objects, err := testutils.LoadYAMLFromFile(configPath, testutils.TemplatingContext{})
				if err != nil {
					return err
				}

				for _, obj := range objects {
					kind := obj.GetObjectKind().GroupVersionKind().Kind

					if kind == "TestSuite" {
						switch ts := obj.(type) {
						case *harness.TestSuite:
							options = *ts
						case *unstructured.Unstructured:
							log.Println(fmt.Errorf("bad configuration in file %q", configPath))
						}

					} else {
						log.Println(fmt.Errorf("unknown object type: %s", kind))
					}
				}
			}

			// Override configuration file options with any command line flags if they are set.
			options.ReportName = "kuttl-test"
			if isSet(flags, "crd-dir") {
				options.CRDDir = crdDir
			}

			if isSet(flags, "manifest-dir") {
				options.ManifestDirs = manifestDirs
			}

			if isSet(flags, "start-control-plane") {
				options.StartControlPlane = startControlPlane
			}

			if isSet(flags, "attach-control-plane-output") {
				options.AttachControlPlaneOutput = attachControlPlaneOutput
			}

			if isSet(flags, "start-kind") {
				options.StartKIND = startKIND
			}

			if isSet(flags, "kind-config") {
				options.StartKIND = true
				options.KINDConfig = kindConfig
			}

			if isSet(flags, "kind-context") {
				options.KINDContext = kindContext
			}

			if options.KINDContext == "" {
				options.KINDContext = harness.DefaultKINDContext
			}

			if options.StartControlPlane && options.StartKIND {
				return errors.New("only one of --start-control-plane and --start-kind can be set")
			}

			// after control-plane && start=kind check
			if options.AttachControlPlaneOutput && !options.StartControlPlane {
				return errors.New("only use --attach-control-plane-output with --start-control-plane")
			}

			if isSet(flags, "skip-delete") {
				options.SkipDelete = skipDelete
			}

			if isSet(flags, "skip-cluster-delete") {
				options.SkipClusterDelete = skipClusterDelete
			}

			if isSet(flags, "parallel") {
				options.Parallel = parallel
			}

			if isSet(flags, "report") {
				var ftype = report.Type(strings.ToLower(reportFormat))
				options.ReportFormat = reportType(ftype)
			}

			if isSet(flags, "artifacts-dir") {
				options.ArtifactsDir = artifactsDir
			}

			if isSet(flags, "namespace") {
				if strings.TrimSpace(namespace) == "" {
					return errors.New(`setting namespace explicitly to "" or empty string is not supported`)
				}
				options.Namespace = namespace
			}

			if isSet(flags, "suppress-log") {
				suppressSet := make(map[string]struct{})
				for _, s := range append(options.Suppress, suppress...) {
					suppressSet[strings.ToLower(s)] = struct{}{}
				}
				options.Suppress = make([]string, len(suppressSet))
				i := 0
				for s := range suppressSet {
					options.Suppress[i] = s
					i++
				}
			}

			if isSet(flags, "timeout") {
				options.Timeout = timeout
			}

			if len(args) != 0 {
				log.Println("kutt-test config testdirs is overridden with args: [", strings.Join(args, ", "), "]")
				options.TestDirs = args
			}

			if len(options.TestDirs) == 0 {
				return errors.New("no test directories provided, please provide either --config or test directories on the command line")
			}
			var APIServerArgs []string
			var err error
			if mockControllerFile != "" {
				APIServerArgs, err = testutils.ReadMockControllerConfig(mockControllerFile)
			} else {
				APIServerArgs = testutils.APIServerDefaultArgs
			}
			if err != nil {
				return err
			}
			options.ControlPlaneArgs = APIServerArgs

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			testutils.RunTests("kuttl", testToRun, options.Parallel, func(t *testing.T) {
				harness := test.Harness{
					TestSuite: options,
					T:         t,
				}

				harness.Run()
			})
		},
	}

	testCmd.Flags().StringVar(&configPath, "config", "", "Path to file to load base test settings from (these may be overridden with command-line arguments).")
	testCmd.Flags().StringVar(&crdDir, "crd-dir", "", "Directory to load CustomResourceDefinitions from prior to running the tests.")
	testCmd.Flags().StringSliceVar(&manifestDirs, "manifest-dir", []string{}, "One or more directories containing manifests to apply before running the tests.")
	testCmd.Flags().StringVar(&testToRun, "test", "", "If set, the specific test case to run.")
	testCmd.Flags().BoolVar(&startControlPlane, "start-control-plane", false, "Start a local Kubernetes control plane for the tests (requires etcd and kube-apiserver binaries, cannot be used with --start-kind).")
	testCmd.Flags().BoolVar(&attachControlPlaneOutput, "attach-control-plane-output", false, "Attaches control plane to stdout when using --start-control-plane.")
	testCmd.Flags().StringVar(&mockControllerFile, "control-plane-config", "", "Path to file to load controller-runtime APIServer configuration arguments (only useful when --startControlPlane).")
	testCmd.Flags().BoolVar(&startKIND, "start-kind", false, "Start a KIND cluster for the tests (cannot be used with --start-control-plane).")
	testCmd.Flags().StringVar(&kindConfig, "kind-config", "", "Specify the KIND configuration file path (implies --start-kind, cannot be used with --start-control-plane).")
	testCmd.Flags().StringVar(&kindContext, "kind-context", "", "Specify the KIND context name to use (default: kind).")
	testCmd.Flags().StringVar(&artifactsDir, "artifacts-dir", "", "Directory to output kind logs to (if not specified, the current working directory).")
	testCmd.Flags().BoolVar(&skipDelete, "skip-delete", false, "If set, do not delete resources created during tests (helpful for debugging test failures, implies --skip-cluster-delete).")
	testCmd.Flags().BoolVar(&skipClusterDelete, "skip-cluster-delete", false, "If set, do not delete the mocked control plane or kind cluster.")
	// The default value here is only used for the help message. The default is actually enforced in RunTests.
	testCmd.Flags().IntVar(&parallel, "parallel", 8, "The maximum number of tests to run at once.")
	testCmd.Flags().IntVar(&timeout, "timeout", 30, "The timeout to use as default for TestSuite configuration.")
	testCmd.Flags().StringVar(&reportFormat, "report", "", "Specify JSON|XML for report.  Report location determined by --artifacts-dir.")
	testCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to use for tests. Provided namespaces must exist prior to running tests.")
	testCmd.Flags().StringSliceVar(&suppress, "suppress-log", []string{}, "Suppress logging for these kinds of logs (events).")
	// This cannot be a global flag because pkg/test/utils.RunTests calls flag.Parse which barfs on unknown top-level flags.
	// Putting it here at least does not advertise it on a level where using it is impossible.
	test.SetFlags(testCmd.Flags())

	return testCmd
}

func reportType(ftype report.Type) string {
	switch ftype {
	case report.JSON:
		fallthrough
	case report.XML:
		return string(ftype)
	default:
		return ""
	}
}

// isSet returns true if a flag is set on the command line.
func isSet(flagSet *pflag.FlagSet, name string) bool {
	found := false

	flagSet.Visit(func(flag *pflag.Flag) {
		if flag.Name == name {
			found = true
		}
	})

	return found
}
