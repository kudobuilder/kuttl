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
	"github.com/kudobuilder/kuttl/pkg/kubernetes"
	"github.com/kudobuilder/kuttl/pkg/report"
	"github.com/kudobuilder/kuttl/pkg/test"
	"github.com/kudobuilder/kuttl/pkg/test/kind"
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
func newTestCmd() *cobra.Command { //nolint:gocyclo
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
	// TODO: remove after v0.16.0 deprecated
	mockControllerFile := ""
	timeout := 30
	reportFormat := ""
	reportName := "kuttl-report"
	reportGranularity := "kuttl-report"
	namespace := ""
	suppress := []string{}
	var runLabels labelSetValue

	options := harness.TestSuite{}

	testCmd := &cobra.Command{
		Use:   "test [flags]... [test suite]...",
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
				objects, err := kubernetes.LoadYAMLFromFile(configPath)
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

			// if we are working with a control plane we can not wait to delete ns (there is no ns controller)
			// this is added before flags potentially override.  control plane should skip ns and cluster delete but
			// perhaps there are cases where that is part of the test.  In general, there is no cluster to delete and
			// there is no namespace controller.
			if options.StartControlPlane {
				options.SkipDelete = true
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

			if isSet(flags, "report-name") {
				options.ReportName = reportName
			}

			if isSet(flags, "report-granularity") {
				if reportGranularity != "step" && reportGranularity != "test" {
					return fmt.Errorf("unrecognized report granularity %q", reportGranularity)
				}
				options.ReportGranularity = reportGranularity
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
			if mockControllerFile != "" {
				log.Println("use of --control-plane-config is deprecated and no longer functions")
			}

			return nil
		},
		Run: func(*cobra.Command, []string) {
			testutils.RunTests("kuttl", testToRun, options.Parallel, func(t *testing.T) {
				h := test.Harness{
					TestSuite: options,
					T:         t,
					RunLabels: runLabels.AsLabelSet(),
				}
				h.Run()
			})
		},
	}

	testCmd.Flags().StringVar(&configPath, "config", "", "Path to file to load base test settings from (these may be overridden with command-line arguments).")
	testCmd.Flags().StringVar(&crdDir, "crd-dir", "", "Directory to load CustomResourceDefinitions from prior to running the tests.")
	testCmd.Flags().StringSliceVar(&manifestDirs, "manifest-dir", []string{}, "One or more directories containing manifests to apply before running the tests.")
	testCmd.Flags().StringVar(&testToRun, "test", "", "If set, the specific test case to run (basename of the test case dir).")
	testCmd.Flags().BoolVar(&startControlPlane, "start-control-plane", false, "Start a local Kubernetes control plane for the tests (requires etcd and kube-apiserver binaries, cannot be used with --start-kind).")
	testCmd.Flags().BoolVar(&attachControlPlaneOutput, "attach-control-plane-output", false, "Attaches control plane to stdout when using --start-control-plane.")
	// TODO: remove after v0.16.0 deprecated mockControllerFile is not supported in the latest testenv
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
	testCmd.Flags().StringVar(&reportName, "report-name", "kuttl-report", "Name for the report.  Report location determined by --artifacts-dir and report file type determined by --report.")
	testCmd.Flags().StringVar(&reportGranularity, "report-granularity", "step", "Report granularity. Can be 'step' (default) or 'test'.")
	testCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to use for tests. Provided namespaces must exist prior to running tests.")
	testCmd.Flags().StringSliceVar(&suppress, "suppress-log", []string{}, "Suppress logging for these kinds of logs (events).")
	testCmd.Flags().Var(&runLabels, "test-run-labels", "Labels to use for this test run.")
	// This cannot be a global flag because pkg/test/utils.RunTests calls flag.Parse which barfs on unknown top-level flags.
	// Putting it here at least does not advertise it on a level where using it is impossible.
	kind.SetFlags(testCmd.Flags())

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
