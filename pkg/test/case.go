package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	eventsbeta1 "k8s.io/api/events/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/report"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// testStepRegex contains one capturing group to determine the index of a step file.
var testStepRegex = regexp.MustCompile(`^(\d+)-(?:[^\.]+)(?:\.yaml)?$`)

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

	if err := cl.Delete(ctx, nsObj); err != nil {
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

// byFirstTimestamp sorts a slice of events by first timestamp, using their involvedObject's name as a tie breaker.
type byFirstTimestamp []eventsbeta1.Event

func (o byFirstTimestamp) Len() int      { return len(o) }
func (o byFirstTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o byFirstTimestamp) Less(i, j int) bool {
	if o[i].ObjectMeta.CreationTimestamp.Equal(&o[j].ObjectMeta.CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].ObjectMeta.CreationTimestamp.Before(&o[j].ObjectMeta.CreationTimestamp)
}

// byFirstTimestampV1 sorts a slice of eventsv1 by first timestamp, using their involvedObject's name as a tie breaker.
type byFirstTimestampV1 []eventsv1.Event

func (o byFirstTimestampV1) Len() int      { return len(o) }
func (o byFirstTimestampV1) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o byFirstTimestampV1) Less(i, j int) bool {
	if o[i].ObjectMeta.CreationTimestamp.Equal(&o[j].ObjectMeta.CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].ObjectMeta.CreationTimestamp.Before(&o[j].ObjectMeta.CreationTimestamp)
}

// byFirstTimestampCoreV1 sorts a slice of corev1 by first timestamp, using their involvedObject's name as a tie breaker.
type byFirstTimestampCoreV1 []corev1.Event

func (o byFirstTimestampCoreV1) Len() int      { return len(o) }
func (o byFirstTimestampCoreV1) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o byFirstTimestampCoreV1) Less(i, j int) bool {
	if o[i].ObjectMeta.CreationTimestamp.Equal(&o[j].ObjectMeta.CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].ObjectMeta.CreationTimestamp.Before(&o[j].ObjectMeta.CreationTimestamp)
}

// CollectEvents gathers all events from namespace and prints it out to log
func (t *Case) CollectEvents(namespace string) {
	cl, err := t.Client(false)
	if err != nil {
		t.Logger.Log("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return
	}

	err = t.collectEventsV1(cl, namespace)
	if err != nil {
		t.Logger.Log("Trying with events eventsv1beta1 API...")
		err = t.collectEventsBeta1(cl, namespace)
		if err != nil {
			t.Logger.Log("Trying with events corev1 API...")
			err = t.collectEventsCoreV1(cl, namespace)
			if err != nil {
				t.Logger.Log("All event APIs failed")
			}
		}
	}
}

func (t *Case) collectEventsBeta1(cl client.Client, namespace string) error {
	eventsList := &eventsbeta1.EventList{}

	err := cl.List(context.TODO(), eventsList, client.InNamespace(namespace))
	if err != nil {
		t.Logger.Logf("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return err
	}

	events := eventsList.Items
	sort.Sort(byFirstTimestamp(events))

	t.Logger.Logf("%s events from ns %s:", t.Name, namespace)
	printEventsBeta1(events, t.Logger)
	return nil
}

func (t *Case) collectEventsV1(cl client.Client, namespace string) error {
	eventsList := &eventsv1.EventList{}

	err := cl.List(context.TODO(), eventsList, client.InNamespace(namespace))
	if err != nil {
		t.Logger.Logf("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return err
	}

	events := eventsList.Items
	sort.Sort(byFirstTimestampV1(events))

	t.Logger.Logf("%s events from ns %s:", t.Name, namespace)
	printEventsV1(events, t.Logger)
	return nil
}

func (t *Case) collectEventsCoreV1(cl client.Client, namespace string) error {
	eventsList := &corev1.EventList{}

	err := cl.List(context.TODO(), eventsList, client.InNamespace(namespace))
	if err != nil {
		t.Logger.Logf("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return err
	}

	events := eventsList.Items
	sort.Sort(byFirstTimestampCoreV1(events))

	t.Logger.Logf("%s events from ns %s:", t.Name, namespace)
	printEventsCoreV1(events, t.Logger)
	return nil
}

func printEventsBeta1(events []eventsbeta1.Event, logger testutils.Logger) {
	for _, e := range events {
		// time type regarding action reason note reportingController related
		logger.Logf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
			e.ObjectMeta.CreationTimestamp,
			e.Type,
			shortString(&e.Regarding),
			e.Action,
			e.Reason,
			e.Note,
			e.ReportingController,
			shortString(e.Related))
	}
}

func printEventsV1(events []eventsv1.Event, logger testutils.Logger) {
	for _, e := range events {
		// time type regarding action reason note reportingController related
		logger.Logf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
			e.ObjectMeta.CreationTimestamp,
			e.Type,
			shortString(&e.Regarding),
			e.Action,
			e.Reason,
			e.Note,
			e.ReportingController,
			shortString(e.Related))
	}
}

func printEventsCoreV1(events []corev1.Event, logger testutils.Logger) {
	for _, e := range events {
		// time type regarding action reason note reportingController related
		logger.Logf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
			e.ObjectMeta.CreationTimestamp,
			e.Type,
			shortString(&e.InvolvedObject),
			e.Action,
			e.Reason,
			e.Message,
			e.ReportingController,
			shortString(e.Related))
	}
}

func shortString(obj *corev1.ObjectReference) string {
	if obj == nil {
		return ""
	}
	fieldRef := ""
	if obj.FieldPath != "" {
		fieldRef = "." + obj.FieldPath
	}
	return fmt.Sprintf("%s %s%s",
		obj.GroupVersionKind().GroupKind().String(),
		obj.Name,
		fieldRef)
}

// Run runs a test case including all of its steps.
func (t *Case) Run(test *testing.T, ts *report.Testsuite) {
	setupReport := report.NewCase("setup")
	ns, err := t.determineNamespace()
	if err != nil {
		setupReport.Failure = report.NewFailure(err.Error(), nil)
		ts.AddTestcase(setupReport)
		test.Fatal(err)
	}

	cl, err := t.Client(false)
	if err != nil {
		setupReport.Failure = report.NewFailure(err.Error(), nil)
		ts.AddTestcase(setupReport)
		test.Fatal(err)
	}

	clients := map[string]client.Client{"": cl}

	for _, testStep := range t.Steps {
		if clients[testStep.Kubeconfig] != nil || testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			continue
		}

		cl, err = newClient(testStep.Kubeconfig)(false)
		if err != nil {
			setupReport.Failure = report.NewFailure(err.Error(), nil)
			ts.AddTestcase(setupReport)
			test.Fatal(err)
		}

		clients[testStep.Kubeconfig] = cl
	}

	for _, c := range clients {
		if err := t.CreateNamespace(test, c, ns); err != nil {
			setupReport.Failure = report.NewFailure(err.Error(), nil)
			ts.AddTestcase(setupReport)
			test.Fatal(err)
		}
	}
	ts.AddTestcase(setupReport)

	for _, testStep := range t.Steps {
		tc := report.NewCase("step " + testStep.String())
		testStep.Client = t.Client
		if testStep.Kubeconfig != "" {
			testStep.Client = newClient(testStep.Kubeconfig)
		}
		testStep.DiscoveryClient = t.DiscoveryClient
		if testStep.Kubeconfig != "" {
			testStep.DiscoveryClient = newDiscoveryClient(testStep.Kubeconfig)
		}
		testStep.Logger = t.Logger.WithPrefix(testStep.String())
		tc.Assertions += len(testStep.Asserts)
		tc.Assertions += len(testStep.Errors)

		errs := []error{}

		if testStep.KubeconfigLoading == v1beta1.KubeconfigLoadingLazy {
			cl, err = newClient(testStep.Kubeconfig)(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to lazy-load kubeconfig '%v': %w", testStep.Kubeconfig, err))
			} else {
				clients[testStep.Kubeconfig] = cl
			}
		}

		errs = append(errs, testStep.Run(test, ns.Name)...)
		if len(errs) > 0 {
			caseErr := fmt.Errorf("failed in step %s", testStep.String())
			tc.Failure = report.NewFailure(caseErr.Error(), errs)

			test.Error(caseErr)
			for _, err := range errs {
				test.Error(err)
			}
		}
		ts.AddTestcase(tc)
		if len(errs) > 0 {
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

// CollectTestStepFiles collects a map of test steps and their associated files
// from a directory.
func (t *Case) CollectTestStepFiles() (map[int64][]string, error) {
	testStepFiles := map[int64][]string{}

	files, err := os.ReadDir(t.Dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		index, err := getIndexFromFile(file.Name())
		if err != nil {
			return nil, err
		}
		if index < 0 {
			t.Logger.Log("Ignoring", file.Name(), "as it does not match file name regexp:", testStepRegex.String())
			continue
		}

		if testStepFiles[index] == nil {
			testStepFiles[index] = []string{}
		}

		testStepPath := filepath.Join(t.Dir, file.Name())

		if file.IsDir() {
			testStepDir, err := os.ReadDir(testStepPath)
			if err != nil {
				return nil, err
			}

			for _, testStepFile := range testStepDir {
				testStepFiles[index] = append(testStepFiles[index], filepath.Join(
					testStepPath, testStepFile.Name(),
				))
			}
		} else {
			testStepFiles[index] = append(testStepFiles[index], testStepPath)
		}
	}

	return testStepFiles, nil
}

// getIndexFromFile returns the index derived from fileName's prefix, ex. "01-foo.yaml" has index 1.
// If an index isn't found, -1 is returned.
func getIndexFromFile(fileName string) (int64, error) {
	matches := testStepRegex.FindStringSubmatch(fileName)
	if len(matches) != 2 {
		return -1, nil
	}

	i, err := strconv.ParseInt(matches[1], 10, 32)
	return i, err
}

// LoadTestSteps loads all of the test steps for a test case.
func (t *Case) LoadTestSteps() error {
	testStepFiles, err := t.CollectTestStepFiles()
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

func newClient(kubeconfig string) func(bool) (client.Client, error) {
	return func(bool) (client.Client, error) {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		return testutils.NewRetryClient(config, client.Options{
			Scheme: testutils.Scheme(),
		})
	}
}

func newDiscoveryClient(kubeconfig string) func() (discovery.DiscoveryInterface, error) {
	return func() (discovery.DiscoveryInterface, error) {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		return discovery.NewDiscoveryClientForConfig(config)
	}
}
