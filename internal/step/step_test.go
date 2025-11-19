package step

import (
	"testing"
	"time"

	kfile "github.com/kudobuilder/kuttl/internal/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	k8sfake "github.com/kudobuilder/kuttl/internal/kubernetes/fake"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

const (
	testNamespace = "world"
)

// Verify the test state as loaded from disk.
// Each test provides a path to a set of test steps and their rendered result.
func TestStepClean(t *testing.T) {
	pod := kubernetes.NewPod("hello", "")

	podWithNamespace := kubernetes.WithNamespace(pod, testNamespace)
	pod2WithNamespace := kubernetes.NewPod("hello2", testNamespace)
	pod2WithDiffNamespace := kubernetes.NewPod("hello2", "different-namespace")

	cl := fake.NewClientBuilder().WithObjects(pod, pod2WithNamespace, pod2WithDiffNamespace).WithScheme(scheme.Scheme).Build()

	step := Step{
		Apply: []client.Object{
			pod, pod2WithDiffNamespace, kubernetes.NewPod("does-not-exist", ""),
		},
		Client:          func(bool) (client.Client, error) { return cl, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return k8sfake.DiscoveryClient(), nil },
	}

	require.NoError(t, step.Clean(testNamespace))

	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(podWithNamespace), podWithNamespace)))
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod2WithNamespace), pod2WithNamespace))
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(pod2WithDiffNamespace), pod2WithDiffNamespace)))
}

// Verify the test state as loaded from disk.
// Each test provides a path to a set of test steps and their rendered result.
func TestStepCreate(t *testing.T) {
	pod := kubernetes.NewPod("hello", "default")
	podWithNamespace := kubernetes.NewPod("hello2", "different-namespace")
	clusterScopedResource := kubernetes.NewResource("v1", "Namespace", "my-namespace", "default")
	podToUpdate := kubernetes.NewPod("update-me", "default")
	specToApply := map[string]interface{}{
		"containers":    nil,
		"restartPolicy": "OnFailure",
	}

	updateToApply := kubernetes.WithSpec(t, podToUpdate, specToApply)

	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(kubernetes.WithNamespace(podToUpdate, testNamespace)).Build()

	step := Step{
		Logger: testutils.NewTestLogger(t, ""),
		Apply: []client.Object{
			pod.DeepCopy(), podWithNamespace.DeepCopy(), clusterScopedResource, updateToApply,
		},
		Client:          func(bool) (client.Client, error) { return cl, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return k8sfake.DiscoveryClient(), nil },
	}

	assert.Equal(t, []error{}, step.Create(t, testNamespace))

	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod))
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(clusterScopedResource), clusterScopedResource))

	updatedPod := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "v1", "kind": "Pod"}}
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podToUpdate), updatedPod))
	assert.Equal(t, specToApply, updatedPod.Object["spec"])

	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podWithNamespace), podWithNamespace))
	actual := kubernetes.NewPod("hello2", testNamespace)
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(actual), actual)))
}

// Verify that the DeleteExisting method properly cleans up resources during a test step.
func TestStepDeleteExisting(t *testing.T) {
	podToDelete := kubernetes.NewPod("delete-me", testNamespace)
	podToDeleteDefaultNS := kubernetes.NewPod("also-delete-me", "default")
	podToKeep := kubernetes.NewPod("keep-me", testNamespace)

	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(podToDelete, podToKeep, podToDeleteDefaultNS).Build()

	step := Step{
		Logger: testutils.NewTestLogger(t, ""),
		Step: &harness.TestStep{
			Delete: []harness.ObjectReference{
				{
					ObjectReference: corev1.ObjectReference{
						Kind:       "Pod",
						APIVersion: "v1",
						Name:       "delete-me",
					},
				},
				{
					ObjectReference: corev1.ObjectReference{
						Kind:       "Pod",
						APIVersion: "v1",
						Name:       "also-delete-me",
						Namespace:  "default",
					},
				},
			},
		},
		Client:          func(bool) (client.Client, error) { return cl, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return k8sfake.DiscoveryClient(), nil },
	}

	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podToKeep), podToKeep))
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podToDelete), podToDelete))
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podToDeleteDefaultNS), podToDeleteDefaultNS))

	require.NoError(t, step.DeleteExisting(testNamespace))

	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(podToKeep), podToKeep))
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(podToDelete), podToDelete)))
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(podToDeleteDefaultNS), podToDeleteDefaultNS)))
}

func TestCheckResource(t *testing.T) {
	for _, test := range []struct {
		testName    string
		actual      []runtime.Object
		expected    runtime.Object
		shouldError bool
	}{
		{
			testName: "resource matches",
			actual: []runtime.Object{
				kubernetes.NewPod("hello", ""),
			},
			expected: kubernetes.NewPod("hello", ""),
		},
		{
			testName: "resource matches with labels",
			actual: []runtime.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("deploy-8b2d", ""),
					map[string]interface{}{
						"containers":         nil,
						"serviceAccountName": "invalid",
					}),
				kubernetes.WithSpec(
					t,
					kubernetes.WithLabels(
						t,
						kubernetes.NewPod("deploy-8c2z", ""),
						map[string]string{"label": "my-label"},
					),
					map[string]interface{}{
						"containers":         nil,
						"serviceAccountName": "valid",
					},
				),
			},

			expected: kubernetes.WithSpec(
				t,
				kubernetes.WithLabels(
					t,
					kubernetes.NewPod("", ""),
					map[string]string{"label": "my-label"},
				),
				map[string]interface{}{
					"containers":         nil,
					"serviceAccountName": "valid",
				},
			),
		},
		{
			testName:    "resource mis-match",
			actual:      []runtime.Object{kubernetes.NewPod("hello", "")},
			expected:    kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), map[string]interface{}{"invalid": "key"}),
			shouldError: true,
		},
		{
			testName: "resource subset match",
			actual: []runtime.Object{kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), map[string]interface{}{
				"containers":    nil,
				"restartPolicy": "OnFailure",
			})},
			expected: kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), map[string]interface{}{
				"restartPolicy": "OnFailure",
			}),
		},
		{
			testName:    "resource does not exist",
			actual:      []runtime.Object{kubernetes.NewPod("other", "")},
			expected:    kubernetes.NewPod("hello", ""),
			shouldError: true,
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			fakeDiscovery := k8sfake.DiscoveryClient()
			namespace := testNamespace

			for _, actualObj := range test.actual {
				_, _, err := kubernetes.Namespaced(fakeDiscovery, actualObj, namespace)
				require.NoError(t, err)
			}

			step := Step{
				Logger: testutils.NewTestLogger(t, ""),
				Client: func(bool) (client.Client, error) {
					return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(test.actual...).Build(), nil
				},
				DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return fakeDiscovery, nil },
			}

			errors := step.CheckResource(test.expected, namespace)
			if test.shouldError {
				assert.NotEqual(t, []error{}, errors)
			} else {
				assert.Equal(t, []error{}, errors)
			}
		})
	}
}

func TestCheckResourceAbsent(t *testing.T) {
	for _, test := range []struct {
		name        string
		actual      []runtime.Object
		expected    runtime.Object
		shouldError bool
		expectedErr string
	}{
		{
			name:        "resource matches",
			actual:      []runtime.Object{kubernetes.NewPod("hello", "")},
			expected:    kubernetes.NewPod("hello", ""),
			shouldError: true,
		},
		{
			name: "one of more resources matches",
			actual: []runtime.Object{
				kubernetes.NewV1Pod("pod1", "", "val1"),
				kubernetes.NewV1Pod("pod2", "", "val2"),
			},
			expected:    kubernetes.WithSpec(t, kubernetes.NewPod("", ""), map[string]interface{}{"serviceAccountName": "val1"}),
			shouldError: true,
			expectedErr: "resource /v1, Kind=Pod pod1 matched error assertion",
		},
		{
			name: "multiple of more resources matches",
			actual: []runtime.Object{
				kubernetes.NewV1Pod("pod1", "", "val1"),
				kubernetes.NewV1Pod("pod2", "", "val1"),
				kubernetes.NewV1Pod("pod3", "", "val2"),
			},
			expected:    kubernetes.WithSpec(t, kubernetes.NewPod("", ""), map[string]interface{}{"serviceAccountName": "val1"}),
			shouldError: true,
			expectedErr: "resource /v1, Kind=Pod pod1 (and 1 other resources) matched error assertion",
		},
		{
			name:     "resource mis-match",
			actual:   []runtime.Object{kubernetes.NewPod("hello", "")},
			expected: kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), map[string]interface{}{"invalid": "key"}),
		},
		{
			name:     "resource does not exist",
			actual:   []runtime.Object{kubernetes.NewPod("other", "")},
			expected: kubernetes.NewPod("hello", ""),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fakeDiscovery := k8sfake.DiscoveryClient()

			for _, object := range test.actual {
				_, _, err := kubernetes.Namespaced(fakeDiscovery, object, testNamespace)
				assert.NoError(t, err)
			}

			step := Step{
				Logger: testutils.NewTestLogger(t, ""),
				Client: func(bool) (client.Client, error) {
					return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(test.actual...).Build(), nil
				},
				DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return fakeDiscovery, nil },
			}

			err := step.CheckResourceAbsent(test.expected, testNamespace)

			if test.shouldError {
				assert.Error(t, err)
				if test.expectedErr != "" {
					assert.EqualError(t, err, test.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	for _, test := range []struct {
		testName     string
		shouldError  bool
		Step         Step
		updateMethod func(*testing.T, client.Client)
	}{
		{
			testName: "successful run", Step: Step{
				Apply: []client.Object{
					kubernetes.NewPod("hello", ""),
				},
				Asserts: []client.Object{
					kubernetes.NewPod("hello", ""),
				},
			},
		},
		{
			"failed run", true, Step{
				Apply: []client.Object{
					kubernetes.NewPod("hello", ""),
				},
				Asserts: []client.Object{
					kubernetes.WithStatus(t, kubernetes.NewPod("hello", ""), map[string]interface{}{
						"phase": "Ready",
					}),
				},
			}, nil,
		},
		{
			"delayed run", false, Step{
				Apply: []client.Object{
					kubernetes.NewPod("hello", ""),
				},
				Asserts: []client.Object{
					kubernetes.WithStatus(t, kubernetes.NewPod("hello", ""), map[string]interface{}{
						"phase": "Ready",
					}),
				},
			}, func(t *testing.T, client client.Client) {
				pod := kubernetes.NewPod("hello", testNamespace)
				require.NoError(t, client.Get(t.Context(), types.NamespacedName{Namespace: testNamespace, Name: "hello"}, pod))

				// mock kubelet to set the pod status
				require.NoError(t, client.Status().Update(t.Context(), kubernetes.WithStatus(t, pod, map[string]interface{}{
					"phase": "Ready",
				})))
			},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			test.Step.Assert = &harness.TestAssert{
				Timeout: 1,
			}

			cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

			test.Step.Client = func(bool) (client.Client, error) { return cl, nil }
			test.Step.DiscoveryClient = func() (discovery.DiscoveryInterface, error) { return k8sfake.DiscoveryClient(), nil }
			test.Step.Logger = testutils.NewTestLogger(t, "")

			if test.updateMethod != nil {
				test.Step.Assert.Timeout = 10

				go func() {
					time.Sleep(time.Second * 2)
					test.updateMethod(t, cl)
				}()
			}

			errors := test.Step.Run(t, testNamespace)

			if test.shouldError {
				assert.NotEqual(t, []error{}, errors)
			} else {
				assert.Equal(t, []error{}, errors)
			}
		})
	}
}

func TestPopulateObjectsByFileName(t *testing.T) {
	for _, tt := range []struct {
		fileName                   string
		isAssert, isError, isApply bool
		name                       string
		errExp                     bool
	}{
		{"00-assert.yaml", true, false, false, "", false},
		{"00-errors.yaml", false, true, false, "", false},
		{"00-foo.yaml", false, false, true, "foo", false},
		{"123-assert.yaml", true, false, false, "", false},
		{"123-errors.yaml", false, true, false, "", false},
		{"123-foo.yaml", false, false, true, "foo", false},
		{"00-assert-bar.yaml", true, false, false, "", false},
		{"00-errors-bar.yaml", false, true, false, "", false},
		{"00-foo-bar.yaml", false, false, true, "foo-bar", false},
		{"00-foo-bar-baz.yaml", false, false, true, "foo-bar-baz", false},
	} {
		t.Run(tt.fileName, func(t *testing.T) {
			step := &Step{}
			err := step.populateObjectsByType(kfile.Parse(tt.fileName), []client.Object{kubernetes.NewPod("foo", "")})
			require.NoError(t, err)
			assert.Equal(t, tt.isAssert, len(step.Asserts) != 0)
			assert.Equal(t, tt.isError, len(step.Errors) != 0)
			assert.Equal(t, tt.isApply, len(step.Apply) != 0)
			if tt.isApply && len(step.Apply) != 0 {
				assert.Equal(t, tt.name, step.Name)
			}
		})
	}
}
