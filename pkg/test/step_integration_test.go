//go:build integration

package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

var testenv testutils.TestEnvironment

func TestMain(m *testing.M) {
	var err error

	testenv, err = testutils.StartTestEnvironment(false)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()
	testenv.Environment.Stop()
	os.Exit(exitCode)
}

func TestCheckResourceIntegration(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for _, test := range []struct {
		testName    string
		actual      []client.Object
		expected    client.Object
		shouldError bool
	}{
		{
			testName: "match object by labels, first in list matches",
			actual: []client.Object{
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("labels-match-pod", ""), map[string]string{
					"app": "nginx",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "nginx:1.7.9",
							"name":  "nginx",
						},
					},
				}),
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("bb", ""), map[string]string{
					"app": "not-match",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "nginx:1.7.9",
							"name":  "nginx",
						},
					},
				}),
			},
			expected: &unstructured.Unstructured{
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
		{
			testName: "match object by labels, last in list matches",
			actual: []client.Object{
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("last-in-list", ""), map[string]string{
					"app": "not-match",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "nginx:1.7.9",
							"name":  "nginx",
						},
					},
				}),
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("bb", ""), map[string]string{
					"app": "nginx",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "nginx:1.7.9",
							"name":  "nginx",
						},
					},
				}),
			},
			expected: &unstructured.Unstructured{
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
		{
			testName: "match object by labels, does not exist",
			actual: []client.Object{
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("hello", ""), map[string]string{
					"app": "NOT-A-MATCH",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "nginx:1.7.9",
							"name":  "nginx",
						},
					},
				}),
			},
			expected: &unstructured.Unstructured{
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
			shouldError: true,
		},
		{
			testName: "match object by labels, field mismatch",
			actual: []client.Object{
				testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("hello", ""), map[string]string{
					"app": "nginx",
				}), map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "otherimage:latest",
							"name":  "nginx",
						},
					},
				}),
			},
			expected: &unstructured.Unstructured{
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
			shouldError: true,
		},
		{
			testName: "step should fail if there are no objects of the same type in the namespace",
			actual:   []client.Object{},
			expected: &unstructured.Unstructured{
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
			shouldError: true,
		},
	} {
		test := test
		t.Run(test.testName, func(t *testing.T) {
			namespace := fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-"))

			err := testenv.Client.Create(context.TODO(), testutils.NewResource("v1", "Namespace", namespace, ""))
			if !k8serrors.IsAlreadyExists(err) {
				// we are ignoring already exists here because in tests we by default use retry client so this can happen
				assert.Nil(t, err)
			}
			for _, actual := range test.actual {
				_, _, err := testutils.Namespaced(testenv.DiscoveryClient, actual, namespace)
				assert.Nil(t, err)

				assert.Nil(t, testenv.Client.Create(context.TODO(), actual))
			}

			step := Step{
				Logger:          testutils.NewTestLogger(t, ""),
				Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
				DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
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

// Verify that the DeleteExisting method properly cleans up resources that are matched on labels during a test step.
func TestStepDeleteExistingLabelMatch(t *testing.T) {
	namespace := "world"

	podSpec := map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"image": "otherimage:latest",
				"name":  "nginx",
			},
		},
	}

	podToDelete := testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("aa-delete-me", "world"), map[string]string{
		"hello": "world",
	}), podSpec)

	podToKeep := testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("bb-dont-delete-me", "world"), map[string]string{
		"bye": "moon",
	}), podSpec)

	podToDelete2 := testutils.WithSpec(t, testutils.WithLabels(t, testutils.NewPod("cc-delete-me", "world"), map[string]string{
		"hello": "world",
	}), podSpec)

	step := Step{
		Logger:  testutils.NewTestLogger(t, ""),
		Timeout: 60,
		Step: &harness.TestStep{
			Delete: []harness.ObjectReference{
				{
					ObjectReference: corev1.ObjectReference{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					Labels: map[string]string{
						"hello": "world",
					},
				},
			},
		},
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	namespaceObj := testutils.NewResource("v1", "Namespace", namespace, "default")

	assert.Nil(t, testenv.Client.Create(context.TODO(), namespaceObj))
	assert.Nil(t, testenv.Client.Create(context.TODO(), podToKeep))
	assert.Nil(t, testenv.Client.Create(context.TODO(), podToDelete))
	assert.Nil(t, testenv.Client.Create(context.TODO(), podToDelete2))

	assert.Nil(t, testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToKeep), podToKeep))
	assert.Nil(t, testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToDelete), podToDelete))
	assert.Nil(t, testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToDelete2), podToDelete2))

	assert.Nil(t, step.DeleteExisting(namespace))

	assert.Nil(t, testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToKeep), podToKeep))
	assert.True(t, k8serrors.IsNotFound(testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToDelete), podToDelete)))
	assert.True(t, k8serrors.IsNotFound(testenv.Client.Get(context.TODO(), testutils.ObjectKey(podToDelete2), podToDelete2)))
}

func TestCheckedTypeAssertions(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
	}{
		{"assert", "TestAssert"},
		{"apply", "TestStep"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			step := Step{}
			path := fmt.Sprintf("step_integration_test_data/error_detect/00-%s.yaml", test.name)
			assert.EqualError(t, step.LoadYAML(path),
				fmt.Sprintf("failed to load %s object from %s: it contains an object of type *unstructured.Unstructured",
					test.typeName, path))
		})
	}
}

func TestApplyExpansion(t *testing.T) {
	os.Setenv("TEST_FOO", "test")
	t.Cleanup(func() {
		os.Unsetenv("TEST_FOO")
	})

	step := Step{Dir: "step_integration_test_data/assert_expand/"}
	path := "step_integration_test_data/assert_expand/00-step1.yaml"
	err := step.LoadYAML(path)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(step.Apply))
}

func TestOverriddenKubeconfigPathResolution(t *testing.T) {
	os.Setenv("SUBPATH", "test")
	t.Cleanup(func() {
		os.Unsetenv("SUBPATH")
	})
	stepRelativePath := &Step{Dir: "step_integration_test_data/kubeconfig_path_resolution/"}
	err := stepRelativePath.LoadYAML("step_integration_test_data/kubeconfig_path_resolution/00-step1.yaml")
	assert.NoError(t, err)
	assert.Equal(t, "step_integration_test_data/kubeconfig_path_resolution/kubeconfig-test.yaml", stepRelativePath.Kubeconfig)

	stepAbsPath := &Step{Dir: "step_integration_test_data/kubeconfig_path_resolution/"}
	err = stepAbsPath.LoadYAML("step_integration_test_data/kubeconfig_path_resolution/00-step2.yaml")
	assert.NoError(t, err)
	assert.Equal(t, "/absolute/kubeconfig-test.yaml", stepAbsPath.Kubeconfig)
}

func TestTwoTestStepping(t *testing.T) {
	apply := []client.Object{}
	step := &Step{
		Name:  "twostepping",
		Index: 0,
		Apply: apply,
	}

	// 2 apply files in 1 step
	err := step.LoadYAML("step_integration_test_data/two_step/00-step1.yaml")
	assert.NoError(t, err)
	err = step.LoadYAML("step_integration_test_data/two_step/00-step2.yaml")
	assert.Error(t, err, "more than 1 TestStep not allowed in step \"twostepping\"")

	// 2 teststeps in 1 file in 1 step
	step = &Step{
		Name:  "twostepping",
		Index: 0,
		Apply: apply,
	}
	err = step.LoadYAML("step_integration_test_data/two_step/01-step1.yaml")
	assert.Error(t, err, "more than 1 TestStep not allowed in step \"twostepping\"")
}

// intentional testing that a test failure captures the test errors and does not have a segfault
// driving by issue: https://github.com/kudobuilder/kuttl/issues/154
func TestStepFailure(t *testing.T) {
	// an assert without setup
	var expected client.Object = &unstructured.Unstructured{
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
	}

	namespace := fmt.Sprintf("kuttl-test-%s", petname.Generate(2, "-"))

	err := testenv.Client.Create(context.TODO(), testutils.NewResource("v1", "Namespace", namespace, ""))
	if !k8serrors.IsAlreadyExists(err) {
		// we are ignoring already exists here because in tests we by default use retry client so this can happen
		assert.Nil(t, err)
	}

	asserts := []client.Object{expected}
	step := Step{
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
		Asserts:         asserts,
		Timeout:         1,
	}

	errs := step.Run(t, namespace)
	assert.Equal(t, len(errs), 1)
}

func TestAssertCommandsValidCommandRunsOk(t *testing.T) {
	step := &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	// Load test that has an echo command, so it should run ok, and don't return any errors
	err := step.LoadYAML("step_integration_test_data/assert_commands/valid_command/00-assert.yaml")
	assert.NoError(t, err)

	errors := step.Run(t, "irrelevant")
	assert.Equal(t, len(errors), 0)
}

func TestAssertCommandsMultipleCommandRunsOk(t *testing.T) {
	step := &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	// Load test that has an echo command, so it should run ok, and don't return any errors
	err := step.LoadYAML("step_integration_test_data/assert_commands/multiple_commands/00-assert.yaml")
	assert.NoError(t, err)

	errors := step.Run(t, "irrelevant")
	assert.Equal(t, len(errors), 0)
}

func TestAssertCommandsMissingCommandFails(t *testing.T) {
	step := &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	// Load test that has an command that is not present (thiscommanddoesnotexist), so it should return an error
	err := step.LoadYAML("step_integration_test_data/assert_commands/command_does_not_exist/00-assert.yaml")
	assert.NoError(t, err)

	errors := step.Run(t, "irrelevant")
	assert.Equal(t, len(errors), 1)
}

func TestAssertCommandsFailingCommandFails(t *testing.T) {
	step := &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	// Load test that has an command that is present but will allways fail (false), so we should get back the error.
	err := step.LoadYAML("step_integration_test_data/assert_commands/failing_comand/00-assert.yaml")
	assert.NoError(t, err)

	errors := step.Run(t, "irrelevant")
	assert.Equal(t, len(errors), 1)
}

func TestAssertCommandsShouldTimeout(t *testing.T) {
	step := &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, ""),
		Client:          func(bool) (client.Client, error) { return testenv.Client, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) { return testenv.DiscoveryClient, nil },
	}

	// Load test that has an command that sleeps for 5 seconds, while the timeout for the step is 1,
	// so we should get back the error, and the test should run in less slightly more than 1 seconds.
	err := step.LoadYAML("step_integration_test_data/assert_commands/timingout_command/00-assert.yaml")
	assert.NoError(t, err)

	start := time.Now()
	errors := step.Run(t, "irrelevant")
	duration := time.Since(start).Seconds()
	assert.Greater(t, duration, float64(1))
	assert.Less(t, duration, float64(5))
	assert.Equal(t, len(errors), 1)
}
