package test

import (
	"context"
	"fmt"
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

	"github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/kubernetes"
	"github.com/kudobuilder/kuttl/pkg/report"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
	eventutils "github.com/kudobuilder/kuttl/pkg/test/utils/events"
	"github.com/kudobuilder/kuttl/pkg/test/utils/files"
)

// Case contains all the test steps and the Kubernetes client and other global configuration
// for a test. It represents a leaf directory containing test step files.
// Case lifecycle:
//  1. gets created (directly by Harness.LoadTests()) before the test's dedicated testing.T
//     comes to file. The following steps are in the scope of the testing.T:
//  2. gets a .Logger assigned
//  3. has .LoadTestSteps() called
//  4. has .Run() called, which:
//     4a. calls setup(), which: determines the namespace name, prepares the clients unless lazy, and prepares the namespaces
//     4b. for each step: sets the step up, prepares its client if lazy, and runs the step
type Case struct {
	Steps              []*Step
	Name               string
	Dir                string
	SkipDelete         bool
	Timeout            int
	PreferredNamespace string
	RunLabels          labels.Set

	ns                 *namespace
	GetClient          func(forceNew bool) (client.Client, error)
	GetDiscoveryClient func() (discovery.DiscoveryInterface, error)

	Logger testutils.Logger
	// Suppress is used to suppress logs
	Suppress []string
}

// namespace contains information about namespace name and its provenance.
type namespace struct {
	name        string
	autoCreated bool
}

func (c *Case) deleteNamespace(cl client.Client) error {
	if !c.ns.autoCreated {
		c.Logger.Log("Skipping deletion of user-supplied namespace:", c.ns.name)
		return nil
	}

	c.Logger.Log("Deleting namespace:", c.ns.name)

	ctx := context.Background()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.Timeout)*time.Second)
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
		c.Logger.Logf("Namespace already cleaned up.")
	} else if err != nil {
		return err
	}

	return wait.PollUntilContextCancel(ctx, 100*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		actual := &corev1.Namespace{}
		err = cl.Get(ctx, client.ObjectKey{Name: c.ns.name}, actual)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
}

func (c *Case) createNamespace(test *testing.T, cl client.Client) error {
	if !c.ns.autoCreated {
		c.Logger.Log("Skipping creation of user-supplied namespace:", c.ns.name)
		return nil
	}
	c.Logger.Log("Creating namespace:", c.ns.name)

	ctx := context.Background()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.Timeout)*time.Second)
		defer cancel()
	}

	if !c.SkipDelete {
		test.Cleanup(func() {
			if err := c.deleteNamespace(cl); err != nil {
				test.Error(err)
			}
		})
	}

	return cl.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ns.name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	})
}

func (c *Case) namespaceExists(namespace string) (bool, error) {
	cl, err := c.GetClient(false)
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
	if funk.Contains(c.Suppress, "events") {
		c.Logger.Logf("skipping kubernetes event logging")
		return
	}
	ctx := context.TODO()
	cl, err := c.GetClient(false)
	if err != nil {
		c.Logger.Log("Failed to collect events for %s in ns %s: %v", c.Name, c.ns.name, err)
		return
	}
	eventutils.CollectAndLog(ctx, cl, c.ns.name, c.Name, c.Logger)
}

// Run runs a test case including all of its steps.
func (c *Case) Run(test *testing.T, rep report.TestReporter) {
	defer rep.Done()

	c.setup(test, rep)

	for _, testStep := range c.Steps {
		stepReport := rep.Step("step " + testStep.String())
		testStep.Setup(c.Logger, c.GetClient, c.GetDiscoveryClient)
		stepReport.AddAssertions(len(testStep.Asserts))
		stepReport.AddAssertions(len(testStep.Errors))

		var errs []error

		// Set-up client/namespace for lazy-loaded Kubeconfig
		if testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			cl, err := testStep.Client(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to lazy-load kubeconfig: %w", err))
			} else if err = c.createNamespace(test, cl); k8serrors.IsAlreadyExists(err) {
				c.Logger.Logf("namespace %q already exists", c.ns.name)
			} else if err != nil {
				errs = append(errs, fmt.Errorf("failed to create test namespace: %w", err))
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

func (c *Case) setup(test *testing.T, rep report.TestReporter) {
	setupReport := rep.Step("setup")
	if err := c.determineNamespace(); err != nil {
		setupReport.Failure(err.Error())
		test.Fatal(err)
	}

	cl, err := c.GetClient(false)
	if err != nil {
		setupReport.Failure(err.Error())
		test.Fatal(err)
	}

	clients := map[string]client.Client{"": cl}

	for _, testStep := range c.Steps {
		if clients[testStep.Kubeconfig] != nil || testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			continue
		}

		cl, err = kubernetes.NewClientFunc(testStep.Kubeconfig, testStep.Context)(false)
		if err != nil {
			setupReport.Failure(err.Error())
			test.Fatal(err)
		}

		clients[testStep.Kubeconfig] = cl
	}

	for kubeConfigPath, cl := range clients {
		if err = c.createNamespace(test, cl); k8serrors.IsAlreadyExists(err) {
			c.Logger.Logf(maybeAppendKubeConfigInfo("namespace %q already exists", kubeConfigPath), c.ns.name)
		} else if err != nil {
			setupReport.Failure("failed to create test namespace", err)
			test.Fatal(err)
		}
	}
}

func (c *Case) determineNamespace() error {
	if c.PreferredNamespace == "" {
		// no preferred ns, means we auto-create with petnames
		c.ns = &namespace{
			name:        fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-")),
			autoCreated: true,
		}
		return nil
	}
	ns := &namespace{
		name:        c.PreferredNamespace,
		autoCreated: false,
	}
	exist, err := c.namespaceExists(ns.name)
	if err != nil {
		return fmt.Errorf("failed to determine existence of namespace %q: %w", c.PreferredNamespace, err)
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
	testStepFiles, err := files.CollectTestStepFiles(c.Dir, c.Logger)
	if err != nil {
		return err
	}

	testSteps := []*Step{}

	for index, files := range testStepFiles {
		testStep := &Step{
			Timeout:       c.Timeout,
			Index:         int(index),
			SkipDelete:    c.SkipDelete,
			Dir:           c.Dir,
			TestRunLabels: c.RunLabels,
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

	c.Steps = testSteps
	return nil
}

// maybeAppendKubeConfigInfo appends a note about kubeConfig, unless empty.
func maybeAppendKubeConfigInfo(msg, kubeConfigPath string) string {
	if kubeConfigPath == "" {
		return msg
	}
	return msg + fmt.Sprintf(" (using kubeconfig %q)", kubeConfigPath)
}
