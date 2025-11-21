package testcase

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kfile "github.com/kudobuilder/kuttl/internal/file"
	"github.com/kudobuilder/kuttl/internal/kubernetes"
	"github.com/kudobuilder/kuttl/internal/report"
	"github.com/kudobuilder/kuttl/internal/step"
	"github.com/kudobuilder/kuttl/internal/template"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
	eventutils "github.com/kudobuilder/kuttl/internal/utils/events"
	"github.com/kudobuilder/kuttl/internal/utils/files"
	"github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

type getClientFuncType func(forceNew bool) (client.Client, error)
type getDiscoveryClientFuncType func() (discovery.DiscoveryInterface, error)

// CaseOption represents a functional option for configuring a Case.
type CaseOption func(*Case)

// WithSkipDelete sets whether to skip deletion of resources.
func WithSkipDelete(skip bool) CaseOption {
	return func(c *Case) {
		c.skipDelete = skip
	}
}

// WithNamespace sets the preferred namespace.
// If empty or not specified, a random namespace name will be generated.
func WithNamespace(ns string) CaseOption {
	return func(c *Case) {
		if ns == "" {
			c.ns = nil // To be filled in by the constructor.
		} else {
			c.ns = &namespace{
				name:         ns,
				userSupplied: true,
			}
		}
	}
}

// WithTimeout sets the timeout in seconds.
func WithTimeout(timeout int) CaseOption {
	return func(c *Case) {
		c.timeout = timeout
	}
}

// WithLogSuppressions sets the list of log types to suppress.
func WithLogSuppressions(suppressions []string) CaseOption {
	return func(c *Case) {
		c.suppressions = suppressions
	}
}

// WithIgnoreFiles sets the list of file patterns to ignore.
func WithIgnoreFiles(patterns []string) CaseOption {
	return func(c *Case) {
		c.ignoreFiles = patterns
	}
}

// WithRunLabels sets the run labels.
func WithRunLabels(runLabels labels.Set) CaseOption {
	return func(c *Case) {
		c.runLabels = runLabels
	}
}

func WithTemplateVars(vars map[string]any) CaseOption {
	return func(c *Case) {
		c.templateEnv = template.Env{Vars: vars}
	}
}

// WithClients sets both the client and discovery client functions.
func WithClients(getClientFunc getClientFuncType, getDiscoveryClientFunc getDiscoveryClientFuncType) CaseOption {
	return func(c *Case) {
		c.getClient = getClientFunc
		c.getDiscoveryClient = getDiscoveryClientFunc
	}
}

// Case contains all the test steps and the Kubernetes client and other global configuration
// for a test. It represents a leaf directory containing test step files.
// Case lifecycle:
//  1. gets created (directly by Harness.LoadTests()) before the test's dedicated testing.T
//     comes to file. At this point the namespace name is determined.
//     The following steps are in the scope of the testing.T:
//  2. has .SetLogger() called to assign a logger
//  3. has .LoadTestSteps() called
//  4. has .Run() called, which:
//     4a. calls setup(), which: prepares the clients unless lazy-loaded, and creates their namespaces if needed
//     (and in this case also schedules namespace deletion for test cleanup time)
//     4b. for each step: sets the step up, prepares its client if lazy-loaded, and runs the step
type Case struct {
	steps              []*step.Step
	name               string
	dir                string
	skipDelete         bool
	timeout            int
	runLabels          labels.Set
	ns                 *namespace
	getClient          getClientFuncType
	getDiscoveryClient getDiscoveryClientFuncType

	logger testutils.Logger
	// List of log types which should be suppressed.
	suppressions []string
	// List of file patterns to ignore when collecting test steps.
	ignoreFiles []string
	// Caution: the Vars element of this struct may be shared with other Case objects.
	templateEnv template.Env
}

// namespace contains information about namespace name and its provenance.
type namespace struct {
	name         string
	userSupplied bool
}

// NewCase returns a new test case object.
func NewCase(name string, parentPath string, options ...CaseOption) *Case {
	c := &Case{name: name, dir: filepath.Join(parentPath, name)}

	for _, option := range options {
		option(c)
	}

	if c.ns == nil {
		c.ns = &namespace{
			name:         fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-")),
			userSupplied: false,
		}
	}

	c.templateEnv.Namespace = c.ns.name

	return c
}

// GetName returns the name of the test case.
func (c *Case) GetName() string {
	return c.name
}

func (c *Case) deleteNamespace(cl clientWithKubeConfig) error {
	cl.Logf("Deleting namespace %q", c.ns.name)

	ctx := context.Background()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.timeout)*time.Second)
		defer cancel()
	}

	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ns.name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	}

	if err := cl.Delete(ctx, nsObj); k8serrors.IsNotFound(err) {
		cl.Logf("Namespace %q already cleaned up.", c.ns.name)
	} else if err != nil {
		return cl.Wrapf(err, "failed to delete namespace %q", c.ns.name)
	}

	err := wait.PollUntilContextCancel(ctx, 100*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		actual := &corev1.Namespace{}
		err = cl.Get(ctx, client.ObjectKey{Name: c.ns.name}, actual)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to check deletion of namespace %q: %w", c.ns.name, err)
		}
		return false, nil
	})
	return cl.Wrapf(err, "waiting for namespace %q to be deleted timed out", c.ns.name)
}

func (c *Case) createNamespace(test *testing.T, cl clientWithKubeConfig) error {
	cl.Logf("Creating namespace %q", c.ns.name)

	ctx := test.Context()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.timeout)*time.Second)
		defer cancel()
	}

	err := cl.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ns.name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	})
	if !c.skipDelete {
		test.Cleanup(func() {
			// Namespace cleanup is tracked per-client for multi-cluster tests.
			// See KEP-0008 for details on backward compatibility decisions.
			if c.ns.userSupplied && k8serrors.IsAlreadyExists(err) {
				cl.Logf("Skipping deletion of pre-existing user supplied namespace %s", c.ns.name)
			} else {
				if err := c.deleteNamespace(cl); err != nil {
					test.Error(err)
				}
			}
		})
	}

	if k8serrors.IsAlreadyExists(err) {
		cl.Logf("Namespace %q already exists", c.ns.name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to create test namespace %q: %w", c.ns.name, err)
	}
	return nil
}

func (c *Case) maybeReportEvents() {
	if funk.Contains(c.suppressions, "events") {
		c.logger.Logf("skipping kubernetes event logging")
		return
	}
	ctx := context.TODO()
	cl, err := c.getClient(false)
	if err != nil {
		c.logger.Log("Failed to collect events for %s in ns %s: %v", c.name, c.ns.name, err)
		return
	}
	eventutils.CollectAndLog(ctx, cl, c.ns.name, c.name, c.logger)
}

// Run runs a test case including all of its steps.
func (c *Case) Run(test *testing.T, rep report.TestReporter) {
	defer rep.Done()

	setupReport := rep.Step("setup")
	if err := c.setup(test); err != nil {
		setupReport.Failure(err.Error())
		test.Fatal(err)
	}

	for _, testStep := range c.steps {
		stepReport := rep.Step("step " + testStep.String())
		testStep.Setup(c.logger, c.getClient, c.getDiscoveryClient)
		stepReport.AddAssertions(len(testStep.Asserts))
		stepReport.AddAssertions(len(testStep.Errors))

		var errs []error

		// Set-up client/namespace for lazy-loaded Kubeconfig
		if testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			cl, err := testStep.Client(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to lazy-load kubeconfig: %w", err))
			} else if err = c.createNamespace(test, clientWithKubeConfig{cl, testStep.Kubeconfig, c.logger}); err != nil {
				errs = append(errs, err)
			}
		}

		// Run test case only if no setup errors are encountered
		if len(errs) == 0 {
			errs = append(errs, testStep.Run(test, c.ns.name)...)
		}

		if len(errs) > 0 {
			caseErr := fmt.Errorf("failed in step %s", testStep.String())
			stepReport.Failure(caseErr.Error(), errs...)

			test.Error(caseErr)
			for _, err := range errs {
				test.Error(err)
			}
			break
		}
	}

	c.maybeReportEvents()
}

func (c *Case) setup(test *testing.T) error {
	clients, err := c.getEagerClients()
	if err != nil {
		return err
	}

	for _, cl := range clients {
		if err := c.createNamespace(test, cl); err != nil {
			return err
		}
	}
	return nil
}

// Returns clients for all steps other than the lazy loaded ones.
// The returned slice will always contain at least a default client with an empty path.
// However, there may be more clients, since each step may optionally specify a path to kubeconfig that should be used.
// This is useful for multi-cluster or multi-context tests.
func (c *Case) getEagerClients() ([]clientWithKubeConfig, error) {
	defaultClient, err := c.getClient(false)
	if err != nil {
		return nil, err
	}

	clients := map[string]client.Client{"": defaultClient}

	for _, testStep := range c.steps {
		if clients[testStep.Kubeconfig] != nil || testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			continue
		}

		var cl client.Client
		if cl, err = kubernetes.NewClientFunc(testStep.Kubeconfig, testStep.Context)(false); err != nil {
			return nil, err
		}
		clients[testStep.Kubeconfig] = cl
	}

	var clientsWithPaths []clientWithKubeConfig
	for kubeConfigPath, cl := range clients {
		clientsWithPaths = append(clientsWithPaths, clientWithKubeConfig{
			Client:         cl,
			kubeConfigPath: kubeConfigPath,
			logger:         c.logger,
		})
	}
	return clientsWithPaths, nil
}

// LoadTestSteps loads all the test steps for a test case.
func (c *Case) LoadTestSteps() error {
	testStepFiles, err := files.CollectTestStepFiles(c.dir, c.logger, c.ignoreFiles)
	if err != nil {
		return err
	}

	testSteps := []*step.Step{}

	for index, files := range testStepFiles {
		testStep := &step.Step{
			Timeout:       c.timeout,
			Index:         int(index),
			SkipDelete:    c.skipDelete,
			Dir:           c.dir,
			TestRunLabels: c.runLabels,
			Asserts:       []client.Object{},
			Apply:         []client.Object{},
			Errors:        []client.Object{},
			TemplateEnv:   c.templateEnv,
		}

		for _, file := range files {
			if err := testStep.LoadYAML(kfile.Parse(file)); err != nil {
				return err
			}
		}

		testSteps = append(testSteps, testStep)
	}

	sort.Slice(testSteps, func(i, j int) bool {
		return testSteps[i].Index < testSteps[j].Index
	})

	c.steps = testSteps
	return nil
}

// SetLogger sets the logger for the test case.
func (c *Case) SetLogger(logger testutils.Logger) {
	c.logger = logger
}
