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
// for a test.
type Case struct {
	Steps              []*Step
	Name               string
	Dir                string
	SkipDelete         bool
	Timeout            int
	PreferredNamespace string
	RunLabels          labels.Set

	GetClient          func(forceNew bool) (client.Client, error)
	GetDiscoveryClient func() (discovery.DiscoveryInterface, error)
	ns                 *namespace

	Logger testutils.Logger
	// Suppress is used to suppress logs
	Suppress []string
}

type namespace struct {
	name         string
	userSupplied bool
}

func (c *Case) maybeDeleteNamespace(cl client.Client, kubeconfigPath string) error {
	if c.SkipDelete {
		c.Logger.Logf(maybeAppendKubeConfigInfo(
			"Skipping namespace deletion as requested with skipDelete=true.", kubeconfigPath))
		return nil
	}
	c.Logger.Log(maybeAppendKubeConfigInfo("Deleting namespace %s", kubeconfigPath), c.ns.name)

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
		c.Logger.Logf("Namespace %s already cleaned up.", c.ns.name)
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

func (c *Case) createNamespace(test *testing.T, cl client.Client, kubeconfigPath string) error {
	c.Logger.Log("Creating namespace:", c.ns.name)

	ctx := context.Background()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.Timeout)*time.Second)
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

	if c.ns.userSupplied && k8serrors.IsAlreadyExists(err) {
		// For backwards compatibility this is the only case where we do not schedule namespace deletion.
		// Arguably, if the namespace name was not user-supplied and yet existed before the test case,
		// it's possible that we're clashing with a different test case, and should abort this whole test case
		// early, not even mentioning namespace deletion. However, it is hard to be sure - this might also
		// be a valid use case, e.g. different kubeconfigs that yet refer to the same cluster,
		// especially combined with lazy kubeconfig loading.
		test.Cleanup(func() {
			c.Logger.Logf(maybeAppendKubeConfigInfo(
				"Skipping deletion of pre-existing user supplied namespace %s", kubeconfigPath), c.ns.name)
		})
	} else {
		test.Cleanup(func() {
			if err := c.maybeDeleteNamespace(cl, kubeconfigPath); err != nil {
				test.Error(err)
			}
		})
	}

	if k8serrors.IsAlreadyExists(err) {
		c.Logger.Logf(maybeAppendKubeConfigInfo("namespace %q already exists", kubeconfigPath), c.ns.name)
		return nil
	}
	return err
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

		// Set-up client/namespace for lazy-loaded Kubeconfig.
		if testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			cl, err := testStep.Client(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to lazy-load kubeconfig: %w", err))
			} else if err = c.createNamespace(test, cl, testStep.Kubeconfig); err != nil {
				errs = append(errs, fmt.Errorf("failed to create test namespace: %w", err))
			}
		}

		// Run test case only if no setup errors are encountered.
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
		if err = c.createNamespace(test, cl, kubeConfigPath); err != nil {
			setupReport.Failure(maybeAppendKubeConfigInfo("failed to create test namespace", kubeConfigPath), err)
			test.Fatal(err)
		}
	}
}

func maybeAppendKubeConfigInfo(msg, kubeconfigPath string) string {
	if kubeconfigPath == "" {
		return msg
	}
	return msg + fmt.Sprintf(" (using kubeconfig %q)", kubeconfigPath)
}

func (c *Case) determineNamespace() {
	if c.PreferredNamespace == "" {
		c.ns = &namespace{
			name:         fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-")),
			userSupplied: false,
		}
	} else {
		c.ns = &namespace{
			name:         c.PreferredNamespace,
			userSupplied: true,
		}
	}
}

// LoadTestSteps loads all the test steps for a test case.
func (c *Case) LoadTestSteps() error {
	c.determineNamespace()

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
