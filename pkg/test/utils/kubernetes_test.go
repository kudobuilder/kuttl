package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/kubernetes"
)

func TestKubeconfigPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		override string
		expected string
	}{
		{name: "no-override", path: "foo", expected: "foo/kubeconfig"},
		{name: "override-relative", path: "foo", override: "bar/kubeconfig", expected: "foo/bar/kubeconfig"},
		{name: "override-abs", path: "foo", override: "/bar/kubeconfig", expected: "/bar/kubeconfig"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := kubeconfigPath(tt.path, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesKind(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.yaml")
	assert.Nil(t, err)
	defer tmpfile.Close()

	err = os.WriteFile(tmpfile.Name(), []byte(`
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

	objs, err := kubernetes.LoadYAMLFromFile(tmpfile.Name())
	assert.Nil(t, err)

	crd := kubernetes.NewResource("apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", "", "")
	pod := kubernetes.NewResource("v1", "Pod", "", "")
	svc := kubernetes.NewResource("v1", "Service", "", "")

	assert.False(t, kubernetes.MatchesKind(objs[0], crd))
	assert.True(t, kubernetes.MatchesKind(objs[0], pod))
	assert.True(t, kubernetes.MatchesKind(objs[0], pod, crd))
	assert.True(t, kubernetes.MatchesKind(objs[0], crd, pod))
	assert.False(t, kubernetes.MatchesKind(objs[0], crd, svc))

	assert.True(t, kubernetes.MatchesKind(objs[1], crd))
	assert.False(t, kubernetes.MatchesKind(objs[1], pod))
	assert.True(t, kubernetes.MatchesKind(objs[1], pod, crd))
	assert.True(t, kubernetes.MatchesKind(objs[1], crd, pod))
	assert.False(t, kubernetes.MatchesKind(objs[1], svc, pod))
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
			_, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")

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
