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

// Case contains all of the test steps and the Kubernetes client and other global configuration
// for a test.
type Case struct {
	Steps              []*Step
	Name               string
	Dir                string
	SkipDelete         bool
	Timeout            int
	PreferredNamespace string
	RunLabels          labels.Set

	Client          func(forceNew bool) (client.Client, error)
	DiscoveryClient func() (discovery.DiscoveryInterface, error)

	Logger testutils.Logger
	// Suppress is used to suppress logs
	Suppress []string
}

type namespace struct {
	Name        string
	AutoCreated bool
}

// DeleteNamespace deletes a namespace in Kubernetes after we are done using it.
func (t *Case) DeleteNamespace(cl client.Client, ns *namespace) error {
	if !ns.AutoCreated {
		t.Logger.Log("Skipping deletion of user-supplied namespace:", ns.Name)
		return nil
	}

	t.Logger.Log("Deleting namespace:", ns.Name)

	ctx := context.Background()
	if t.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
		defer cancel()
	}

	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns.Name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	}

	if err := cl.Delete(ctx, nsObj); k8serrors.IsNotFound(err) {
		t.Logger.Logf("Namespace already cleaned up.")
	} else if err != nil {
		return err
	}

	return wait.PollUntilContextCancel(ctx, 100*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		actual := &corev1.Namespace{}
		err = cl.Get(ctx, client.ObjectKey{Name: ns.Name}, actual)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
}

// CreateNamespace creates a namespace in Kubernetes to use for a test.
func (t *Case) CreateNamespace(test *testing.T, cl client.Client, ns *namespace) error {
	if !ns.AutoCreated {
		t.Logger.Log("Skipping creation of user-supplied namespace:", ns.Name)
		return nil
	}
	t.Logger.Log("Creating namespace:", ns.Name)

	ctx := context.Background()
	if t.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
		defer cancel()
	}

	if !t.SkipDelete {
		test.Cleanup(func() {
			if err := t.DeleteNamespace(cl, ns); err != nil {
				test.Error(err)
			}
		})
	}

	return cl.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns.Name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	})
}

// NamespaceExists gets namespace and returns true if it exists
func (t *Case) NamespaceExists(namespace string) (bool, error) {
	cl, err := t.Client(false)
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

// CollectEvents gathers all events from namespace and prints it out to log
func (t *Case) CollectEvents(namespace string) {
	ctx := context.TODO()
	cl, err := t.Client(false)
	if err != nil {
		t.Logger.Log("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return
	}
	eventutils.CollectAndLog(ctx, cl, namespace, t.Name, t.Logger)
}

// Run runs a test case including all of its steps.
func (t *Case) Run(test *testing.T, rep report.TestReporter) {
	defer rep.Done()
	setupReport := rep.Step("setup")
	ns, err := t.determineNamespace()
	if err != nil {
		setupReport.Failure(err.Error())
		test.Fatal(err)
	}

	cl, err := t.Client(false)
	if err != nil {
		setupReport.Failure(err.Error())
		test.Fatal(err)
	}

	clients := map[string]client.Client{"": cl}

	for _, testStep := range t.Steps {
		if clients[testStep.Kubeconfig] != nil || testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			continue
		}

		cl, err = newClient(testStep.Kubeconfig, testStep.Context)(false)
		if err != nil {
			setupReport.Failure(err.Error())
			test.Fatal(err)
		}

		clients[testStep.Kubeconfig] = cl
	}

	for kc, c := range clients {
		if err = t.CreateNamespace(test, c, ns); k8serrors.IsAlreadyExists(err) {
			t.Logger.Logf("namespace %q already exists, using kubeconfig %q", ns.Name, kc)
		} else if err != nil {
			setupReport.Failure("failed to create test namespace", err)
			test.Fatal(err)
		}
	}

	for _, testStep := range t.Steps {
		stepReport := rep.Step("step " + testStep.String())
		testStep.Client = t.Client
		if testStep.Kubeconfig != "" {
			testStep.Client = newClient(testStep.Kubeconfig, testStep.Context)
		}
		testStep.DiscoveryClient = t.DiscoveryClient
		if testStep.Kubeconfig != "" {
			testStep.DiscoveryClient = newDiscoveryClient(testStep.Kubeconfig, testStep.Context)
		}
		testStep.Logger = t.Logger.WithPrefix(testStep.String())
		stepReport.AddAssertions(len(testStep.Asserts))
		stepReport.AddAssertions(len(testStep.Errors))

		errs := []error{}

		// Set-up client/namespace for lazy-loaded Kubeconfig
		if testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			cl, err = testStep.Client(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to lazy-load kubeconfig: %w", err))
			} else if err = t.CreateNamespace(test, cl, ns); k8serrors.IsAlreadyExists(err) {
				t.Logger.Logf("namespace %q already exists", ns.Name)
			} else if err != nil {
				errs = append(errs, fmt.Errorf("failed to create test namespace: %w", err))
			}
		}

		// Run test case only if no setup errors are encountered
		if len(errs) == 0 {
			errs = append(errs, testStep.Run(test, ns.Name)...)
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

	if funk.Contains(t.Suppress, "events") {
		t.Logger.Logf("skipping kubernetes event logging")
	} else {
		t.CollectEvents(ns.Name)
	}
}

func (t *Case) determineNamespace() (*namespace, error) {
	ns := &namespace{
		Name:        t.PreferredNamespace,
		AutoCreated: false,
	}
	// no preferred ns, means we auto-create with petnames
	if t.PreferredNamespace == "" {
		ns.Name = fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-"))
		ns.AutoCreated = true
	} else {
		exist, err := t.NamespaceExists(t.PreferredNamespace)
		if err != nil {
			return nil, fmt.Errorf("failed to determine existence of namespace %q: %w", t.PreferredNamespace, err)
		}
		if !exist {
			ns.AutoCreated = true
		}
	}
	// if we have a preferred namespace, and it already exists, we do NOT auto-create
	return ns, nil
}

// LoadTestSteps loads all of the test steps for a test case.
func (t *Case) LoadTestSteps() error {
	testStepFiles, err := files.CollectTestStepFiles(t.Dir, t.Logger)
	if err != nil {
		return err
	}

	testSteps := []*Step{}

	for index, files := range testStepFiles {
		testStep := &Step{
			Timeout:       t.Timeout,
			Index:         int(index),
			SkipDelete:    t.SkipDelete,
			Dir:           t.Dir,
			TestRunLabels: t.RunLabels,
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

	t.Steps = testSteps
	return nil
}

func newClient(kubeconfig, context string) func(bool) (client.Client, error) {
	return func(bool) (client.Client, error) {
		config, err := kubernetes.BuildConfigWithContext(kubeconfig, context)
		if err != nil {
			return nil, err
		}

		return kubernetes.NewRetryClient(config, client.Options{
			Scheme: kubernetes.Scheme(),
		})
	}
}

func newDiscoveryClient(kubeconfig, context string) func() (discovery.DiscoveryInterface, error) {
	return func() (discovery.DiscoveryInterface, error) {
		config, err := kubernetes.BuildConfigWithContext(kubeconfig, context)
		if err != nil {
			return nil, err
		}

		return discovery.NewDiscoveryClientForConfig(config)
	}
}
