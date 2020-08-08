package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	eventsbeta1 "k8s.io/api/events/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kudobuilder/kuttl/pkg/report"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

var testStepRegex = regexp.MustCompile(`^(\d+)-([^.]+)(.yaml)?$`)

// Case contains all of the test steps and the Kubernetes client and other global configuration
// for a test.
type Case struct {
	Steps              []*Step
	Name               string
	Dir                string
	SkipDelete         bool
	Timeout            int
	PreferredNamespace string

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
func (t *Case) DeleteNamespace(ns *namespace) error {
	if !ns.AutoCreated {
		t.Logger.Log("Skipping deletion of user-supplied namespace:", ns.Name)
		return nil
	}

	t.Logger.Log("Deleting namespace:", ns.Name)

	cl, err := t.Client(false)
	if err != nil {
		return err
	}

	return cl.Delete(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns.Name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	})
}

// CreateNamespace creates a namespace in Kubernetes to use for a test.
func (t *Case) CreateNamespace(ns *namespace) error {
	if !ns.AutoCreated {
		t.Logger.Log("Skipping creation of user-supplied namespace:", ns.Name)
		return nil
	}
	t.Logger.Log("Creating namespace:", ns.Name)

	cl, err := t.Client(false)
	if err != nil {
		return err
	}

	return cl.Create(context.TODO(), &corev1.Namespace{
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
	if err != nil && !errors.IsNotFound(err) {
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

// CollectEvents gathers all events from namespace and prints it out to log
func (t *Case) CollectEvents(namespace string) {
	cl, err := t.Client(false)
	if err != nil {
		t.Logger.Log("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return
	}

	eventsList := &eventsbeta1.EventList{}

	err = cl.List(context.TODO(), eventsList, client.InNamespace(namespace))
	if err != nil {
		t.Logger.Logf("Failed to collect events for %s in ns %s: %v", t.Name, namespace, err)
		return
	}

	events := eventsList.Items
	sort.Sort(byFirstTimestamp(events))

	t.Logger.Logf("%s events from ns %s:", t.Name, namespace)
	printEvents(events, t.Logger)
}

func printEvents(events []eventsbeta1.Event, logger conversion.DebugLogger) {
	for _, e := range events {
		// time type reason kind message
		logger.Logf("%s\t%s\t%s\t%s", e.ObjectMeta.CreationTimestamp, e.Type, e.Reason, e.Note)
	}
}

// Run runs a test case including all of its steps.
func (t *Case) Run(test *testing.T, tc *report.Testcase) {
	test.Parallel()

	ns, err := t.determineNamespace()
	if err != nil {
		test.Fatal(err)
	}

	if err := t.CreateNamespace(ns); err != nil {
		test.Fatal(err)
	}

	if !t.SkipDelete {
		defer func() {
			if err := t.DeleteNamespace(ns); err != nil {
				test.Error(err)
			}
		}()
	}

	for _, testStep := range t.Steps {
		testStep.Client = t.Client
		testStep.DiscoveryClient = t.DiscoveryClient
		testStep.Logger = t.Logger.WithPrefix(testStep.String())
		tc.Assertions += len(testStep.Asserts)
		tc.Assertions += len(testStep.Errors)

		if !t.SkipDelete {
			defer func() {
				if err := testStep.Clean(ns.Name); err != nil {
					test.Error(err)
				}
			}()
		}

		if errs := testStep.Run(ns.Name); len(errs) > 0 {
			caseErr := fmt.Errorf("failed in step %s", testStep.String())
			tc.Failure = report.NewFailure(caseErr.Error(), errs)

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
		ns.Name = fmt.Sprintf("kudo-test-%s", petname.Generate(2, "-"))
		ns.AutoCreated = true
	}
	// if we have a preferred namespace, we do NOT auto-create
	return ns, nil
}

// CollectTestStepFiles collects a map of test steps and their associated files
// from a directory.
func (t *Case) CollectTestStepFiles() (map[int64][]string, error) {
	testStepFiles := map[int64][]string{}

	files, err := ioutil.ReadDir(t.Dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		matches := testStepRegex.FindStringSubmatch(file.Name())

		if len(matches) < 2 {
			t.Logger.Log("Ignoring", file.Name(), "as it does not match file name regexp:", testStepRegex.String())
			continue
		}

		index, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			return nil, err
		}

		if testStepFiles[index] == nil {
			testStepFiles[index] = []string{}
		}

		testStepPath := filepath.Join(t.Dir, file.Name())

		if file.IsDir() {
			testStepDir, err := ioutil.ReadDir(testStepPath)
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

// LoadTestSteps loads all of the test steps for a test case.
func (t *Case) LoadTestSteps() error {
	testStepFiles, err := t.CollectTestStepFiles()
	if err != nil {
		return err
	}

	testSteps := []*Step{}

	for index, files := range testStepFiles {
		testStep := &Step{
			Timeout: t.Timeout,
			Index:   int(index),
			Dir:     t.Dir,
			Asserts: []runtime.Object{},
			Apply:   []runtime.Object{},
			Errors:  []runtime.Object{},
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
