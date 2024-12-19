//go:build integration

package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

func buildTestStep(t *testing.T) *Step {
	return &Step{
		Name:   t.Name(),
		Index:  0,
		Logger: testutils.NewTestLogger(t, t.Name()),
		Client: func(bool) (client.Client, error) {
			return testenv.Client, nil
		},
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) {
			return testenv.DiscoveryClient, nil
		},
	}
}

func TestAssertExpressions(t *testing.T) {
	codednsDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-dns"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"k8s-app": "kube-dns"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"k8s-app": "kube-dns"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "coredns",
							Image: "registry.k8s.io/coredns/coredns:v1.11.1",
						},
					},
				},
			},
		},
	}
	metricServerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server-xyz-pqr",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "metrics-server",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "metrics-server",
					Image: "registry.k8s.io/metrics-server/metrics-server:v0.7.2",
				},
			},
		},
	}

	assert.NoError(t, testenv.Client.Create(context.TODO(), codednsDeployment))
	assert.NoError(t, testenv.Client.Create(context.TODO(), metricServerPod))

	testCases := []struct {
		name                 string
		expectLoadFailure    bool
		expectRunFailure     bool
		expectedErrorMessage string
	}{
		{
			name:                 "invalid expression",
			expectLoadFailure:    true,
			expectedErrorMessage: "undeclared reference",
		},
		{
			name: "check deployment name",
		},
		{
			name:                 "check incorrect deployment name",
			expectRunFailure:     true,
			expectedErrorMessage: "not all assertAll expressions evaluated to true",
		},
		{
			name: "check multiple assert all",
		},
		{
			name:                 "check multiple assert all with one failing",
			expectRunFailure:     true,
			expectedErrorMessage: "not all assertAll expressions evaluated to true",
		},
		{
			name: "check multiple assert any",
		},
		{
			name:                 "check multiple assert any with all failing",
			expectRunFailure:     true,
			expectedErrorMessage: "no expression evaluated to true",
		},
		{
			name: "check expression for ephemeral namespace",
		},
	}

	const testNamespace = "kuttl-ephemeral-xyz"
	assert.NoError(t, testenv.Client.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
	}))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirName := fmt.Sprintf(
				"step_integration_test_data/assert_expressions/%s",
				strings.ReplaceAll(tc.name, " ", "_"),
			)

			files, err := os.ReadDir(dirName)
			assert.NoError(t, err)

			for i := 0; i < len(files)-1; i++ {
				fName := fmt.Sprintf("%s/%s", dirName, files[i].Name())
				step := buildTestStep(t)
				assert.NoError(t, step.LoadYAML(fName))
				assert.NoError(t, errors.Join(errors.Join(step.Run(t, testNamespace)...)))
			}

			step := buildTestStep(t)

			fName := fmt.Sprintf("%s/%s", dirName, files[len(files)-1].Name())

			err = step.LoadYAML(fName)
			if !tc.expectLoadFailure {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expectedErrorMessage)
				return
			}

			err = errors.Join(step.Run(t, testNamespace)...)
			if !tc.expectRunFailure {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expectedErrorMessage)
			}
		})
	}
}
