package utils

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

func TestNamespaced(t *testing.T) {
	fake := FakeDiscoveryClient()

	for _, test := range []struct {
		testName    string
		resource    runtime.Object
		namespace   string
		shouldError bool
	}{
		{
			testName:  "namespaced resource",
			resource:  NewPod("hello", ""),
			namespace: "set-the-namespace",
		},
		{
			testName:  "namespace already set",
			resource:  NewPod("hello", "other"),
			namespace: "other",
		},
		{
			testName:  "not-namespaced resource",
			resource:  NewResource("v1", "Namespace", "hello", ""),
			namespace: "",
		},
		{
			testName:    "non-existent resource",
			resource:    NewResource("v1", "Blah", "hello", ""),
			shouldError: true,
		},
	} {
		test := test

		t.Run(test.testName, func(t *testing.T) {
			m, _ := meta.Accessor(test.resource)

			actualName, actualNamespace, err := Namespaced(fake, test.resource, "set-the-namespace")

			if test.shouldError {
				assert.NotNil(t, err)
				assert.Equal(t, "", actualName)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, m.GetName(), actualName)
			}

			assert.Equal(t, test.namespace, actualNamespace)
			assert.Equal(t, test.namespace, m.GetNamespace())
		})
	}
}

func TestGETAPIResource(t *testing.T) {
	fake := FakeDiscoveryClient()

	apiResource, err := GetAPIResource(fake, schema.GroupVersionKind{
		Kind:    "Pod",
		Version: "v1",
	})
	assert.Nil(t, err)
	assert.Equal(t, apiResource.Kind, "Pod")

	apiResource, err = GetAPIResource(fake, schema.GroupVersionKind{
		Kind:    "NonExistentResourceType",
		Version: "v1",
	})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "resource type not found")
}

func TestRetry(t *testing.T) {
	index := 0

	assert.Nil(t, Retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("ignore this error")
		}
		return nil
	}, func(err error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))

	assert.Equal(t, 2, index)
}

func TestRetryWithUnexpectedError(t *testing.T) {
	index := 0

	assert.Equal(t, errors.New("bad error"), Retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("bad error")
		}
		return nil
	}, func(err error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))
	assert.Equal(t, 1, index)
}

func TestRetryWithNil(t *testing.T) {
	assert.Equal(t, nil, Retry(context.TODO(), nil, IsJSONSyntaxError))
}

func TestRetryWithNilFromFn(t *testing.T) {
	assert.Equal(t, nil, Retry(context.TODO(), func(ctx context.Context) error {
		return nil
	}, IsJSONSyntaxError))
}

func TestRetryWithNilInFn(t *testing.T) {
	client := RetryClient{}
	var list runtime.Object
	assert.Error(t, Retry(context.TODO(), func(ctx context.Context) error {
		return client.Client.List(ctx, list)
	}, IsJSONSyntaxError))
}

func TestRetryWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	assert.Equal(t, errors.New("error"), Retry(ctx, func(context.Context) error {
		return errors.New("error")
	}, func(err error) bool { return true }))
}

func TestLoadYAML(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test.yaml")
	assert.Nil(t, err)
	defer tmpfile.Close()

	err = ioutil.WriteFile(tmpfile.Name(), []byte(`
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: nginx
  name: hello
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	objs, err := LoadYAMLFromFile(tmpfile.Name())
	assert.Nil(t, err)

	assert.Equal(t, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "nginx",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image": "nginx:1.7.9",
						"name":  "nginx",
					},
				},
			},
		},
	}, objs[0])

	assert.Equal(t, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"name": "hello",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image": "nginx:1.7.9",
						"name":  "nginx",
					},
				},
			},
		},
	}, objs[1])
}

func TestMatchesKind(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test.yaml")
	assert.Nil(t, err)
	defer tmpfile.Close()

	err = ioutil.WriteFile(tmpfile.Name(), []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: hello
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hello
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	objs, err := LoadYAMLFromFile(tmpfile.Name())
	assert.Nil(t, err)

	crd := NewResource("apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", "", "")
	pod := NewResource("v1", "Pod", "", "")
	svc := NewResource("v1", "Service", "", "")

	assert.False(t, MatchesKind(objs[0], crd))
	assert.True(t, MatchesKind(objs[0], pod))
	assert.True(t, MatchesKind(objs[0], pod, crd))
	assert.True(t, MatchesKind(objs[0], crd, pod))
	assert.False(t, MatchesKind(objs[0], crd, svc))

	assert.True(t, MatchesKind(objs[1], crd))
	assert.False(t, MatchesKind(objs[1], pod))
	assert.True(t, MatchesKind(objs[1], pod, crd))
	assert.True(t, MatchesKind(objs[1], crd, pod))
	assert.False(t, MatchesKind(objs[1], svc, pod))
}

func TestGetKubectlArgs(t *testing.T) {
	for _, test := range []struct {
		testName  string
		namespace string
		args      string
		env       map[string]string
		expected  []string
	}{
		{
			testName:  "namespace long, combined already set at end is not modified",
			namespace: "default",
			args:      "kubectl kuttl test --namespace=test-canary",
			expected: []string{
				"kubectl", "kuttl", "test", "--namespace=test-canary",
			},
		},
		{
			testName:  "namespace long already set at end is not modified",
			namespace: "default",
			args:      "kubectl kuttl test --namespace test-canary",
			expected: []string{
				"kubectl", "kuttl", "test", "--namespace", "test-canary",
			},
		},
		{
			testName:  "namespace short, combined already set at end is not modified",
			namespace: "default",
			args:      "kubectl kuttl test -n=test-canary",
			expected: []string{
				"kubectl", "kuttl", "test", "-n=test-canary",
			},
		},
		{
			testName:  "namespace short already set at end is not modified",
			namespace: "default",
			args:      "kubectl kuttl test -n test-canary",
			expected: []string{
				"kubectl", "kuttl", "test", "-n", "test-canary",
			},
		},
		{
			testName:  "namespace long, combined already set in middle is not modified",
			namespace: "default",
			args:      "kubectl kuttl --namespace=test-canary test",
			expected: []string{
				"kubectl", "kuttl", "--namespace=test-canary", "test",
			},
		},
		{
			testName:  "namespace long already set in middle is not modified",
			namespace: "default",
			args:      "kubectl kuttl --namespace test-canary test",
			expected: []string{
				"kubectl", "kuttl", "--namespace", "test-canary", "test",
			},
		},
		{
			testName:  "namespace short, combined already set in middle is not modified",
			namespace: "default",
			args:      "kubectl kuttl -n=test-canary test",
			expected: []string{
				"kubectl", "kuttl", "-n=test-canary", "test",
			},
		},
		{
			testName:  "namespace short already set in middle is not modified",
			namespace: "default",
			args:      "kubectl kuttl -n test-canary test",
			expected: []string{
				"kubectl", "kuttl", "-n", "test-canary", "test",
			},
		},
		{
			testName:  "namespace not set is appended",
			namespace: "default",
			args:      "kubectl kuttl test",
			expected: []string{
				"kubectl", "kuttl", "test", "--namespace", "default",
			},
		},
		{
			testName:  "unknown arguments do not break parsing with namespace is not set",
			namespace: "default",
			args:      "kubectl kuttl test --config kuttl-test.yaml",
			expected: []string{
				"kubectl", "kuttl", "test", "--config", "kuttl-test.yaml", "--namespace", "default",
			},
		},
		{
			testName:  "unknown arguments do not break parsing if namespace is set at beginning",
			namespace: "default",
			args:      "kubectl --namespace=test-canary kuttl test --config kuttl-test.yaml",
			expected: []string{
				"kubectl", "--namespace=test-canary", "kuttl", "test", "--config", "kuttl-test.yaml",
			},
		},
		{
			testName:  "unknown arguments do not break parsing if namespace is set at middle",
			namespace: "default",
			args:      "kubectl kuttl --namespace=test-canary test --config kuttl-test.yaml",
			expected: []string{
				"kubectl", "kuttl", "--namespace=test-canary", "test", "--config", "kuttl-test.yaml",
			},
		},
		{
			testName:  "unknown arguments do not break parsing if namespace is set at end",
			namespace: "default",
			args:      "kubectl kuttl test --config kuttl-test.yaml --namespace=test-canary",
			expected: []string{
				"kubectl", "kuttl", "test", "--config", "kuttl-test.yaml", "--namespace=test-canary",
			},
		},
		{
			testName:  "quotes are respected when parsing",
			namespace: "default",
			args:      "kubectl kuttl \"test quoted\"",
			expected: []string{
				"kubectl", "kuttl", "test quoted", "--namespace", "default",
			},
		},
		{
			testName:  "os ENV are expanded",
			namespace: "default",
			args:      "kubectl kuttl $TEST_FOO ${TEST_FOO}",
			env:       map[string]string{"TEST_FOO": "test"},
			expected: []string{
				"kubectl", "kuttl", "test", "test", "--namespace", "default",
			},
		},
		{
			testName:  "kubectl is not pre-pended if it is already present",
			namespace: "default",
			args:      "kubectl kuttl test",
			expected: []string{
				"kubectl", "kuttl", "test", "--namespace", "default",
			},
		},
	} {
		test := test

		t.Run(test.testName, func(t *testing.T) {

			if test.env != nil || len(test.env) > 0 {
				for key, value := range test.env {
					os.Setenv(key, value)
				}
				defer func() {
					for key := range test.env {
						os.Unsetenv(key)
					}
				}()
			}
			cmd, err := GetArgs(context.TODO(), harness.Command{
				Command:    test.args,
				Namespaced: true,
			}, test.namespace, nil)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, cmd.Args)
		})
	}
}

func TestRunScript(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		script         string
		wantedErr      bool
		expectedStdout bool
	}{
		{
			name:           `no script and no command`,
			command:        "",
			script:         "",
			wantedErr:      true,
			expectedStdout: false,
		},
		{
			name:           `script AND command`,
			command:        "echo 'hello'",
			script:         "for i in {1..5}; do echo $NAMESPACE; done",
			wantedErr:      true,
			expectedStdout: false,
		},
		// failure for script command as a command (reason we need a script script option)
		{
			name:           `command has a failing script command`,
			command:        "for i in {1..5}; do echo $NAMESPACE; done",
			script:         "",
			wantedErr:      true,
			expectedStdout: false,
		},
		{
			name:           `working script command`,
			command:        "",
			script:         "for i in {1..5}; do echo $NAMESPACE; done",
			wantedErr:      false,
			expectedStdout: true,
		},
	}

	for _, tt := range tests {

		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			hcmd := harness.Command{
				Command: tt.command,
				Script:  tt.script,
			}

			logger := NewTestLogger(t, "")
			// script runs with output
			_, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0)

			if tt.wantedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.expectedStdout {
				assert.True(t, stdout.Len() > 0)
			} else {
				assert.True(t, stdout.Len() == 0)
			}
		})
	}
}

func TestExtractGVKFromCRD(t *testing.T) {
	for _, test := range []struct {
		name         string
		inputCRDs    []runtime.Object
		expectedGVKs []schema.GroupVersionKind
		shouldError  bool
	}{
		{
			name: "Nominal. Unstructured CRDs",
			inputCRDs: []runtime.Object{
				NewCRDv1(t, "test", "test.net", "testresource", []string{"v1alpha1", "v1alpha2"}),
				NewCRDv1beta1(t, "test2", "test.net", "testresource2", "v1beta1"),
				NewCRDv1beta1(t, "test3", "test.net", "testresource3", "v1"),
			},
			expectedGVKs: []schema.GroupVersionKind{
				{
					Group:   "test.net",
					Version: "v1alpha1",
					Kind:    "testresource",
				},
				{
					Group:   "test.net",
					Version: "v1alpha2",
					Kind:    "testresource",
				},
				{
					Group:   "test.net",
					Version: "v1beta1",
					Kind:    "testresource2",
				},
				{
					Group:   "test.net",
					Version: "v1",
					Kind:    "testresource3",
				},
			},
		},
		{
			name: "Structured CRDs",
			inputCRDs: []runtime.Object{
				newTestCRDv1("test", "test.net", "testresource", []string{"v1alpha1", "v1alpha2"}),
				newTestCRDv1beta1("test2", "test.net", "testresource2", "v1beta1"),
				newTestCRDv1beta1("test3", "test.net", "testresource3", "v1"),
			},
			expectedGVKs: []schema.GroupVersionKind{
				{
					Group:   "test.net",
					Version: "v1alpha1",
					Kind:    "testresource",
				},
				{
					Group:   "test.net",
					Version: "v1alpha2",
					Kind:    "testresource",
				},
				{
					Group:   "test.net",
					Version: "v1beta1",
					Kind:    "testresource2",
				},
				{
					Group:   "test.net",
					Version: "v1",
					Kind:    "testresource3",
				},
			},
		},
		{
			name: "Error. Wrong kind",
			inputCRDs: []runtime.Object{
				NewResource("apiextensions.k8s.io/v1alpha1", "CustomResourceDefinition", "", ""),
			},
			shouldError: true,
		},
		{
			name: "Error. Wrong type",
			inputCRDs: []runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			shouldError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			gotGVKs, err := ExtractGVKFromCRD(test.inputCRDs)
			if test.shouldError {
				assert.NotNil(t, err)
				assert.Nil(t, gotGVKs)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedGVKs, gotGVKs)
			}
		})
	}
}

func newTestCRDv1(name, group, resourceKind string, resourceVersions []string) *apiextv1.CustomResourceDefinition {
	crdVersions := []apiextv1.CustomResourceDefinitionVersion{}
	for _, v := range resourceVersions {
		crdVersions = append(crdVersions, apiextv1.CustomResourceDefinitionVersion{
			Name: v,
		})
	}

	return &apiextv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group:    group,
			Versions: crdVersions,
			Names: apiextv1.CustomResourceDefinitionNames{
				Kind: resourceKind,
			},
		},
	}
}

func newTestCRDv1beta1(name, group, resourceKind, resourceVersion string) *apiextv1beta1.CustomResourceDefinition {
	return &apiextv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group:   group,
			Version: resourceVersion,
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Kind: resourceKind,
			},
		},
	}
}
