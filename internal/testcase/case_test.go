package testcase

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
