//go:build integration

package testcase

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	"github.com/kudobuilder/kuttl/internal/report"
	"github.com/kudobuilder/kuttl/internal/step"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

// Create two test environments, ensure that the second environment is used when
// Kubeconfig is set on a Step.
func TestMultiClusterCase(t *testing.T) {
	testenv, err := kubernetes.StartTestEnvironment(false)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, testenv.Environment.Stop())
	})

	testenv2, err := kubernetes.StartTestEnvironment(false)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, testenv2.Environment.Stop())
	})

	podSpec := map[string]interface{}{
		"restartPolicy": "Never",
		"containers": []map[string]interface{}{
			{
				"name":  "nginx",
				"image": "nginx:1.7.9",
			},
		},
	}

	tmpfile, err := os.CreateTemp(t.TempDir(), "kubeconfig")
	require.NoError(t, err)
	require.NoError(t, kubernetes.Kubeconfig(testenv2.Config, tmpfile))

	c := NewCase("multicluster", "", true, "", 20, nil, nil,
		func(bool) (client.Client, error) {
			return testenv.Client, nil
		},
		func() (discovery.DiscoveryInterface, error) {
			return testenv.DiscoveryClient, nil
		},
	)
	c.SetLogger(testutils.NewTestLogger(t, ""))
	c.steps = []*step.Step{
		{
			Name:  "initialize-testenv",
			Index: 0,
			Apply: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), podSpec),
			},
			Asserts: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), podSpec),
			},
			Timeout: 2,
		},
		{
			Name:  "use-testenv2",
			Index: 1,
			Apply: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello2", ""), podSpec),
			},
			Asserts: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello2", ""), podSpec),
			},
			Errors: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), podSpec),
			},
			Timeout:    2,
			Kubeconfig: tmpfile.Name(),
		},
		{
			Name:  "verify-testenv-does-not-have-testenv2-resources",
			Index: 2,
			Asserts: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello", ""), podSpec),
			},
			Errors: []client.Object{
				kubernetes.WithSpec(t, kubernetes.NewPod("hello2", ""), podSpec),
			},
			Timeout: 2,
		},
	}

	c.Run(t, &noOpReporter{})
}

type noOpReporter struct{}

func (r *noOpReporter) Done() {}
func (r *noOpReporter) Step(string) report.StepReporter {
	return r
}
func (r *noOpReporter) AddAssertions(int)        {}
func (r *noOpReporter) Failure(string, ...error) {}
