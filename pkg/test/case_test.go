package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// Verify the test state as loaded from disk.
// Each test provides a path to a set of test steps and their rendered result.
func TestLoadTestSteps(t *testing.T) {
	for _, tt := range []struct {
		path      string
		runLabels labels.Set
		testSteps []Step
	}{
		{
			"test_data/with-overrides",
			labels.Set{},
			[]Step{
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
						testutils.WithSpec(t, testutils.NewPod("test", ""), map[string]interface{}{
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
						testutils.WithStatus(t, testutils.NewPod("test", ""), map[string]interface{}{
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
						testutils.WithSpec(t, testutils.NewPod("test2", ""), map[string]interface{}{
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
						testutils.WithStatus(t, testutils.NewPod("test2", ""), map[string]interface{}{
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
						testutils.WithSpec(t, testutils.NewPod("test4", ""), map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
						testutils.WithSpec(t, testutils.NewPod("test3", ""), map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
					},
					Asserts: []client.Object{
						testutils.WithStatus(t, testutils.NewPod("test3", ""), map[string]interface{}{
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
						testutils.WithSpec(t, testutils.NewPod("test6", ""), map[string]interface{}{
							"restartPolicy": "Never",
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						}),
						testutils.WithSpec(t, testutils.NewPod("test5", ""), map[string]interface{}{
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
						testutils.WithSpec(t, testutils.NewPod("test5", ""), map[string]interface{}{
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
			[]Step{
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
			[]Step{
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
			[]Step{
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
			[]Step{
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
		tt := tt

		t.Run(fmt.Sprintf("%s/%s", tt.path, tt.runLabels), func(t *testing.T) {
			test := &Case{Dir: tt.path, Logger: testutils.NewTestLogger(t, tt.path), RunLabels: tt.runLabels}

			err := test.LoadTestSteps()
			assert.Nil(t, err)

			testStepsVal := []Step{}
			for _, testStep := range test.Steps {
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

func TestCollectTestStepFiles(t *testing.T) {
	for _, tt := range []struct {
		path     string
		expected map[int64][]string
	}{
		{
			"test_data/with-overrides",
			map[int64][]string{
				int64(0): {
					"test_data/with-overrides/00-assert.yaml",
					"test_data/with-overrides/00-test-step.yaml",
				},
				int64(1): {
					"test_data/with-overrides/01-assert.yaml",
					"test_data/with-overrides/01-test-assert.yaml",
				},
				int64(2): {
					"test_data/with-overrides/02-directory/assert.yaml",
					"test_data/with-overrides/02-directory/pod.yaml",
					"test_data/with-overrides/02-directory/pod2.yaml",
				},
				int64(3): {
					"test_data/with-overrides/03-assert.yaml",
					"test_data/with-overrides/03-pod.yaml",
					"test_data/with-overrides/03-pod2.yaml",
				},
			},
		},
		{
			"test_data/list-pods",
			map[int64][]string{
				int64(0): {
					"test_data/list-pods/00-assert.yaml",
					"test_data/list-pods/00-pod.yaml",
				},
			},
		},
	} {
		tt := tt

		t.Run(tt.path, func(t *testing.T) {
			test := &Case{Dir: tt.path, Logger: testutils.NewTestLogger(t, tt.path)}
			testStepFiles, err := test.CollectTestStepFiles()
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, testStepFiles)
		})
	}
}

func TestGetIndexFromFile(t *testing.T) {
	for _, tt := range []struct {
		fileName string
		indexExp int64
	}{
		{"00-foo.yaml", 0},
		{"01-foo.yaml", 1},
		{"1-foo.yaml", 1},
		{"01-foo", 1},
		{"01234-foo.yaml", 1234},
		{"1-foo-bar.yaml", 1},
		{"01.yaml", -1},
		{"foo-01.yaml", -1},
	} {
		tt := tt

		t.Run(tt.fileName, func(t *testing.T) {
			index, err := getIndexFromFile(tt.fileName)
			assert.Nil(t, err)
			assert.Equal(t, tt.indexExp, index)
		})
	}
}
