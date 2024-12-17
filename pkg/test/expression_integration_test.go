//go:build integration

package test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func buildTestStep(t *testing.T) *Step {
	codednsDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
	}
	metricServerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server-xyz-pqr",
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "metrics-server",
			},
		},
	}

	assert.NoError(t, testenv.Client.Create(context.TODO(), codednsDeployment))
	assert.NoError(t, testenv.Client.Create(context.TODO(), metricServerPod))

	return &Step{
		Name:            t.Name(),
		Index:           0,
		Logger:          testutils.NewTestLogger(t, t.Name()),
		Client:          testenv.Client,
		DiscoveryClient: testenv.DiscoveryClient,
	}
}

func TestAssertExpressions(t *testing.T) {
	testCases := []struct {
		name          string
		loadingFailed bool
		runFailed     bool
		errorMessage  string
	}{
		{
			name:          "invalid expression",
			loadingFailed: true,
			errorMessage:  "undeclared reference",
		},
		{
			name: "check deployment name",
		},
		{
			name:         "check incorrect deployment name",
			runFailed:    true,
			errorMessage: "not all expressions evaluated to true",
		},
		{
			name: "check multiple assert all",
		},
		{
			name:         "check multiple assert all with one failing",
			runFailed:    true,
			errorMessage: "not all expressions evaluated to true",
		},
		{
			name: "check multiple assert any",
		},
		{
			name:         "check multiple assert any with all failing",
			runFailed:    true,
			errorMessage: "no expression evaluated to true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			step := buildTestStep(t)

			fName := fmt.Sprintf(
				"step_integration_test_data/assert_expressions/%s/00-assert.yaml",
				strings.ReplaceAll(tc.name, " ", "_"),
			)

			// Load test that has an invalid expression
			err := step.LoadYAML(fName)
			if !tc.loadingFailed {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errorMessage)
				return
			}

			err = errors.Join(step.Run(t, "")...)
			if !tc.runFailed {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errorMessage)
			}
		})
	}
}
