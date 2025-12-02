package testcase

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	"github.com/kudobuilder/kuttl/internal/step"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

// Verify the test state as loaded from disk.
// Each test provides a path to a set of test steps and their rendered result.
func TestLoadTestSteps(t *testing.T) {
	for _, tt := range []struct {
		path      string
		runLabels labels.Set
		testSteps []step.Step
	}{
		{
			"test_data/with-overrides",
			labels.Set{},
			[]step.Step{
				{
					Name:  "with-test-step-name-override",
					Index: 0,
					Step: &harness.TestStep{
						ObjectMeta: metav1.ObjectMeta{
							Name: "with-test-step-name-override",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "TestStep",
							APIVersion: "kuttl.dev/v1beta1",
						},
						Index: 0,
					},
					Apply: []client.Object{
						kubernetes.WithSpec(t, kubernetes.NewPod("test", ""), map[string]interface{}{
							"restartPolicy": "Never",
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
					},
					Asserts: []client.Object{
						kubernetes.WithStatus(t, kubernetes.NewPod("test", ""), map[string]interface{}{
							"qosClass": "BestEffort",
						}),
					},
					Errors:        []client.Object{},
					TestRunLabels: labels.Set{},
				},
				{
					Name:  "test-assert",
					Index: 1,
					Step: &harness.TestStep{
						TypeMeta: metav1.TypeMeta{
							Kind:       "TestStep",
							APIVersion: "kuttl.dev/v1beta1",
						},
						Index: 1,
						Delete: []harness.ObjectReference{
							{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "v1",
									Kind:       "Pod",
									Name:       "test",
								},
							},
						},
					},
					Assert: &harness.TestAssert{
						TypeMeta: metav1.TypeMeta{
							Kind:       "TestAssert",
							APIVersion: "kuttl.dev/v1beta1",
						},
						Timeout: 20,
					},
					Apply: []client.Object{
						kubernetes.WithSpec(t, kubernetes.NewPod("test2", ""), map[string]interface{}{
							"restartPolicy": "Never",
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
					},
					Asserts: []client.Object{
						kubernetes.WithStatus(t, kubernetes.NewPod("test2", ""), map[string]interface{}{
							"qosClass": "BestEffort",
						}),
					},
					Errors:        []client.Object{},
					TestRunLabels: labels.Set{},
				},
				{
					Name:  "pod",
					Index: 2,
					Apply: []client.Object{
						kubernetes.WithSpec(t, kubernetes.NewPod("test4", ""), map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
						kubernetes.WithSpec(t, kubernetes.NewPod("test3", ""), map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
					},
					Asserts: []client.Object{
						kubernetes.WithStatus(t, kubernetes.NewPod("test3", ""), map[string]interface{}{
							"qosClass": "BestEffort",
						}),
					},
					Errors:        []client.Object{},
					TestRunLabels: labels.Set{},
				},
				{
					Name:  "name-overridden",
					Index: 3,
					Step: &harness.TestStep{
						ObjectMeta: metav1.ObjectMeta{
							Name: "name-overridden",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "TestStep",
							APIVersion: "kuttl.dev/v1beta1",
						},
						Index: 3,
					},
					Apply: []client.Object{
						kubernetes.WithSpec(t, kubernetes.NewPod("test6", ""), map[string]interface{}{
							"restartPolicy": "Never",
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
						kubernetes.WithSpec(t, kubernetes.NewPod("test5", ""), map[string]interface{}{
							"restartPolicy": "Never",
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
					},
					Asserts: []client.Object{
						kubernetes.WithSpec(t, kubernetes.NewPod("test5", ""), map[string]interface{}{
							"restartPolicy": "Never",
						}),
					},
					Errors:        []client.Object{},
					TestRunLabels: labels.Set{},
				},
			},
		},
		{
			"test_data/list-pods",
			labels.Set{},
			[]step.Step{
				{
					Name:  "pod",
					Index: 0,
					Apply: []client.Object{
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Pod",
								"metadata": map[string]interface{}{
									"name": "pod-1",
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
						},
					},
					Asserts: []client.Object{
						&unstructured.Unstructured{
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
						},
					},
					Errors:        []client.Object{},
					TestRunLabels: labels.Set{},
				},
			},
		},
		{
			"test_data/test-run-labels",
			labels.Set{},
			[]step.Step{
				{
					Name:          "",
					Index:         1,
					TestRunLabels: labels.Set{},
					Apply:         []client.Object{},
					Asserts:       []client.Object{},
					Errors:        []client.Object{},
				},
			},
		},
		{
			"test_data/test-run-labels",
			labels.Set{"flavor": "a"},
			[]step.Step{
				{
					Name:          "create-a",
					Index:         1,
					TestRunLabels: labels.Set{"flavor": "a"},
					Apply: []client.Object{
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "test",
								},
								"data": map[string]interface{}{
									"flavor": "a",
								},
							},
						},
					},
					Asserts: []client.Object{
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "test",
								},
								"data": map[string]interface{}{
									"flavor": "a",
								},
							},
						},
					},
					Errors: []client.Object{},
				},
			},
		},
		{
			"test_data/test-run-labels",
			labels.Set{"flavor": "b"},
			[]step.Step{
				{
					Name:          "create-b",
					Index:         1,
					TestRunLabels: labels.Set{"flavor": "b"},
					Apply: []client.Object{
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "test",
								},
								"data": map[string]interface{}{
									"flavor": "b",
								},
							},
						},
					},
					Asserts: []client.Object{
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "test",
								},
								"data": map[string]interface{}{
									"flavor": "b",
								},
							},
						},
					},
					Errors: []client.Object{},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s/%s", tt.path, tt.runLabels), func(t *testing.T) {
			test := &Case{dir: tt.path, logger: testutils.NewTestLogger(t, tt.path), runLabels: tt.runLabels}

			err := test.LoadTestSteps()
			require.NoError(t, err)

			testStepsVal := []step.Step{}
			for _, testStep := range test.steps {
				testStepsVal = append(testStepsVal, *testStep)
			}

			assert.Equal(t, len(tt.testSteps), len(testStepsVal))
			for index := range tt.testSteps {
				tt.testSteps[index].Dir = tt.path
				assert.Equal(t, tt.testSteps[index].Apply, testStepsVal[index].Apply, "apply objects need to match")
				assert.Equal(t, tt.testSteps[index].Asserts, testStepsVal[index].Asserts, "assert objects need to match")
				assert.Equal(t, tt.testSteps[index].Errors, testStepsVal[index].Errors, "error objects need to match")
				assert.Equal(t, tt.testSteps[index].Step, testStepsVal[index].Step, "step object needs to match")
				assert.Equal(t, tt.testSteps[index].Dir, testStepsVal[index].Dir, "dir needs to match")
				assert.Equal(t, tt.testSteps[index], testStepsVal[index])
			}
		})
	}
}

// testMock is an object useful for unit-testing Case.createNamespace().
type testMock struct {
	cleanup   func()
	testError []any
}

func (t *testMock) Context() context.Context {
	return context.Background()
}

func (t *testMock) Cleanup(f func()) {
	t.cleanup = f
}

func (t *testMock) Error(args ...any) {
	t.testError = args
}

func TestCase_createNamespace(t *testing.T) {
	tests := map[string]struct {
		options               []CaseOption
		cl                    func(*testing.T, string) client.Client
		wantErr               error
		expectedCleanupErrors int
		getNsBeforeCleanup    func(*testing.T, error)
		getNsAfterCleanup     func(*testing.T, error)
	}{
		"user-supplied exists": {
			options: []CaseOption{WithNamespace("foo")},
			cl:      newClientWithExistingNs,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"user-supplied absent": {
			options: []CaseOption{WithNamespace("foo")},
			cl:      newClientWithAbsentNs,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be deleted after cleanup, but client returned %v", err)
			},
		},
		"user-supplied absent and no write permission": {
			options: []CaseOption{WithNamespace("foo")},
			cl:      newClientWithAbsentNsNoWritePerm,
			wantErr: errCreationForbidden,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be missing before cleanup, but client returned %v", err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be missing after cleanup, but client returned %v", err)
			},
		},
		"user-supplied exists and no write permission": {
			options: []CaseOption{WithNamespace("foo")},
			cl:      newClientWithExistingNsNoWritePerm,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"user-supplied exists and no permissions at all": {
			options: []CaseOption{WithNamespace("foo")},
			cl:      newClientWithExistingNsNoPerms,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"ephemeral exists": {
			cl: newClientWithExistingNs,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be deleted after cleanup, but client returned %v", err)
			},
		},
		"ephemeral absent": {
			cl: newClientWithAbsentNs,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be deleted after cleanup, but client returned %v", err)
			},
		},
		"ephemeral absent and no write permission": {
			cl:      newClientWithAbsentNsNoWritePerm,
			wantErr: errCreationForbidden,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be missing before cleanup, but client returned %v", err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				assert.True(t, k8serrors.IsNotFound(err), "expected namespace to be missing after cleanup, but client returned %v", err)
			},
		},
		"ephemeral exists and no write permission": {
			cl:                    newClientWithExistingNsNoWritePerm,
			expectedCleanupErrors: 1,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"ephemeral exists and no permissions at all": {
			cl:                    newClientWithExistingNsNoPerms,
			expectedCleanupErrors: 1,
			getNsBeforeCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			getNsAfterCleanup: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCase(name, "", tt.options...)
			tm := &testMock{}
			cl := tt.cl(t, c.ns.name)
			if npc, ok := cl.(*noPermClient); ok {
				npc.t = t
			}
			clk := clientWithKubeConfig{
				Client:         cl,
				kubeConfigPath: "kubeconfig/path",
				logger:         testutils.NewTestLogger(t, ""),
			}

			gotErr := c.createNamespace(tm, clk)
			if tt.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.ErrorIs(t, gotErr, tt.wantErr)
			}

			baseClient := cl
			if npc, ok := cl.(*noPermClient); ok {
				baseClient = npc.Client
			}

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: c.ns.name,
				},
			}
			err := baseClient.Get(t.Context(), kubernetes.ObjectKey(ns), ns)
			tt.getNsBeforeCleanup(t, err)

			if tm.cleanup != nil {
				tm.cleanup()
			}

			assert.Len(t, tm.testError, tt.expectedCleanupErrors)
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: c.ns.name,
				},
			}
			err = baseClient.Get(t.Context(), kubernetes.ObjectKey(ns), ns)

			tt.getNsAfterCleanup(t, err)
		})
	}
}

// noPermClient wraps a client and returns forbidden errors for Create/Delete operations.
// Optionally it also refuses Get operations.
type noPermClient struct {
	client.Client
	forbidGet bool
	t         *testing.T
}

var errCreationForbidden = k8serrors.NewForbidden(schema.GroupResource{Group: "", Resource: "namespaces"}, "foo", fmt.Errorf("forbidden: User cannot create resource \"namespaces\""))

func (c *noPermClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	c.t.Logf("Create object %v refused", obj.GetObjectKind().GroupVersionKind())
	return errCreationForbidden
}

func (c *noPermClient) Delete(_ context.Context, obj client.Object, _ ...client.DeleteOption) error {
	c.t.Logf("Delete object %v refused", obj.GetObjectKind().GroupVersionKind())
	return k8serrors.NewForbidden(schema.GroupResource{Group: "", Resource: "namespaces"}, obj.GetName(), fmt.Errorf("forbidden: User cannot delete resource \"namespaces\""))
}

func (c *noPermClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.forbidGet {
		c.t.Logf("Get object %v refused", key)
		return k8serrors.NewForbidden(schema.GroupResource{Group: "", Resource: "namespaces"}, obj.GetName(), fmt.Errorf("forbidden: User cannot get resource \"namespaces\""))
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func newClientWithExistingNsNoWritePerm(t *testing.T, nsName string) client.Client {
	return &noPermClient{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsName,
			},
		}).Build(),
		forbidGet: false,
		t:         t,
	}
}
func newClientWithAbsentNsNoWritePerm(t *testing.T, _ string) client.Client {
	return &noPermClient{
		Client:    fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
		forbidGet: false,
		t:         t,
	}
}
func newClientWithExistingNsNoPerms(t *testing.T, nsName string) client.Client {
	return &noPermClient{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsName,
			},
		}).Build(),
		forbidGet: true,
		t:         t,
	}
}

func newClientWithAbsentNs(*testing.T, string) client.Client {
	return fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
}

func newClientWithExistingNs(_ *testing.T, nsName string) client.Client {
	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}).Build()
}
