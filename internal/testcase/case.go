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

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	"github.com/kudobuilder/kuttl/internal/report"
	"github.com/kudobuilder/kuttl/internal/step"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
	eventutils "github.com/kudobuilder/kuttl/internal/utils/events"
	"github.com/kudobuilder/kuttl/internal/utils/files"
	"github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

type getClientFuncType func(forceNew bool) (client.Client, error)
type getDiscoveryClientFuncType func() (discovery.DiscoveryInterface, error)

// Case contains all the test steps and the Kubernetes client and other global configuration
// for a test. It represents a leaf directory containing test step files.
// Case lifecycle:
//  1. gets created (directly by Harness.LoadTests()) before the test's dedicated testing.T
//     comes to file. The following steps are in the scope of the testing.T:
//  2. has .SetLogger() called to assign a logger
//  3. has .LoadTestSteps() called
//  4. has .Run() called, which:
//     4a. calls setup(), which: determines the namespace name, prepares the clients unless lazy, and prepares the namespaces
//     4b. for each step: sets the step up, prepares its client if lazy, and runs the step
type Case struct {
	steps              []*step.Step
	name               string
	dir                string
	skipDelete         bool
	timeout            int
	preferredNamespace string
	runLabels          labels.Set

	ns                 *namespace
	getClient          getClientFuncType
	getDiscoveryClient getDiscoveryClientFuncType

	logger testutils.Logger
	// List of log types which should be suppressed.
	suppressions []string
}

// namespace contains information about namespace name and its provenance.
type namespace struct {
	name        string
	autoCreated bool
}

// NewCase returns a new test case object.
func NewCase(name string, parentPath string, skipDelete bool, preferredNamespace string, timeout int, suppressions []string, runLabels labels.Set, getClientFunc getClientFuncType, getDiscoveryClientFunc getDiscoveryClientFuncType) *Case {
	return &Case{
		name:               name,
		dir:                filepath.Join(parentPath, name),
		skipDelete:         skipDelete,
		preferredNamespace: preferredNamespace,
		timeout:            timeout,
		suppressions:       suppressions,
		runLabels:          runLabels,
		getClient:          getClientFunc,
		getDiscoveryClient: getDiscoveryClientFunc,
	}
}

// GetName returns the name of the test case.
func (c *Case) GetName() string {
	return c.name
}

func (c *Case) deleteNamespace(cl client.Client, kubeconfigPath string) error {
	if !c.ns.autoCreated {
		c.logkcf(kubeconfigPath, "Skipping deletion of user-supplied namespace %q", c.ns.name)
		return nil
	}

	c.logkcf(kubeconfigPath, "Deleting namespace %q", c.ns.name)

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
		c.logkcf(kubeconfigPath, "Namespace %q already cleaned up.", c.ns.name)
	} else if err != nil {
		return fmt.Errorf("failed to delete namespace %q%s: %w", c.ns.name, getKubeConfigInfo(kubeconfigPath), err)
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
	if err != nil {
		return fmt.Errorf("waiting for namespace %q to be deleted timed out%s: %w", c.ns.name, getKubeConfigInfo(kubeconfigPath), err)
	}
	return nil
}

func (c *Case) createNamespace(test *testing.T, cl client.Client, kubeconfigPath string) error {
	if !c.ns.autoCreated {
		c.logkcf(kubeconfigPath, "Skipping creation of user-supplied namespace %q", c.ns.name)
		return nil
	}
	c.logkcf(kubeconfigPath, "Creating namespace %q", c.ns.name)

	ctx := test.Context()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.timeout)*time.Second)
		defer cancel()
	}

	if !c.skipDelete {
		test.Cleanup(func() {
			if err := c.deleteNamespace(cl, kubeconfigPath); err != nil {
				test.Error(err)
			}
		})
	}

	err := cl.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ns.name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	})
	if k8serrors.IsAlreadyExists(err) {
		c.logkcf(kubeconfigPath, "namespace %q already exists", c.ns.name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to create test namespace %q: %w", c.ns.name, err)
	}
	return nil
}

func (c *Case) namespaceExists(namespace string) (bool, error) {
	cl, err := c.getClient(false)
	if err != nil {
		return false, err
	}
	ns := &corev1.Namespace{}
	err = cl.Get(context.TODO(), client.ObjectKey{Name: namespace}, ns)
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, err
	}
	return ns.Name == namespace, nil
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
			} else if err = c.createNamespace(test, cl, testStep.Kubeconfig); err != nil {
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
	if err := c.determineNamespace(); err != nil {
		return err
	}

	clients, err := c.getEagerClients()
	if err != nil {
		return err
	}

	for kubeConfigPath, cl := range clients {
		if err := c.createNamespace(test, cl, kubeConfigPath); err != nil {
			return err
		}
	}
	return nil
}

// Returns clients for all steps other than the lazy loaded ones.
// The returned map will always contain at least a default client with a key of empty string.
// However, there may be more pairs, since each step may optionally specify a path to kubeconfig that should be used.
// This is useful for multi-cluster or multi-context tests.
// For those, the client will be keyed with the specified path to kubeconfig file.
func (c *Case) getEagerClients() (map[string]client.Client, error) {
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
	return clients, nil
}

func (c *Case) determineNamespace() error {
	if c.preferredNamespace == "" {
		// no preferred ns, means we auto-create with petnames
		c.ns = &namespace{
			name:        fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-")),
			autoCreated: true,
		}
		return nil
	}
	ns := &namespace{
		name:        c.preferredNamespace,
		autoCreated: false,
	}
	exist, err := c.namespaceExists(ns.name)
	if err != nil {
		return fmt.Errorf("failed to determine existence of namespace %q: %w", c.preferredNamespace, err)
	}
	if !exist {
		ns.autoCreated = true
	}
	c.ns = ns
	// if we have a preferred namespace, and it already exists, we do NOT auto-create
	return nil
}

// LoadTestSteps loads all the test steps for a test case.
func (c *Case) LoadTestSteps() error {
	testStepFiles, err := files.CollectTestStepFiles(c.dir, c.logger)
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
		}

		for _, file := range files {
			if err := testStep.LoadYAML(file); err != nil {
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

// logkcf behaves like logger.Logf, but potentially appends a note about which kubeconfig is being used.
// See also getKubeConfigInfo.
func (c *Case) logkcf(kubeConfigPath string, format string, args ...any) {
	c.logger.Log(fmt.Sprintf(format, args...) + getKubeConfigInfo(kubeConfigPath))
}

// getKubeConfigInfo returns a note about kubeConfig (with a space prepended), unless empty.
func getKubeConfigInfo(kubeConfigPath string) string {
	if kubeConfigPath == "" {
		return ""
	}
	return fmt.Sprintf(" (using kubeconfig %q)", kubeConfigPath)
}
