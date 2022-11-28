//go:build integration

package test

import (
	"os"
	"testing"

	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyverno/kuttl/pkg/report"
	testutils "github.com/kyverno/kuttl/pkg/test/utils"
)

// Create two test environments, ensure that the second environment is used when
// Kubeconfig is set on a Step.
func TestMultiClusterCase(t *testing.T) {
	testenv, err := testutils.StartTestEnvironment(testutils.APIServerDefaultArgs, false)
	if err != nil {
		t.Error(err)
		return
	}
	t.Cleanup(func() {
		if err := testenv.Environment.Stop(); err != nil {
			t.Error(err)
		}
	})

	testenv2, err := testutils.StartTestEnvironment(testutils.APIServerDefaultArgs, false)
	if err != nil {
		t.Error(err)
		return
	}
	t.Cleanup(func() {
		if err := testenv2.Environment.Stop(); err != nil {
			t.Error(err)
		}
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

	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		t.Error(err)
		return
	}
	t.Cleanup(func() {
		os.Remove(tmpfile.Name())
	})
	if err := testutils.Kubeconfig(testenv2.Config, tmpfile); err != nil {
		t.Error(err)
		return
	}

	c := Case{
		Logger: testutils.NewTestLogger(t, ""),
		Steps: []*Step{
			{
				Name:  "initialize-testenv",
				Index: 0,
				Apply: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Asserts: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Timeout: 2,
			},
			{
				Name:  "use-testenv2",
				Index: 1,
				Apply: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello2", ""), podSpec),
				},
				Asserts: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello2", ""), podSpec),
				},
				Errors: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Timeout:    2,
				Kubeconfig: tmpfile.Name(),
			},
			{
				Name:  "verify-testenv-does-not-have-testenv2-resources",
				Index: 2,
				Asserts: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Errors: []client.Object{
					testutils.WithSpec(t, testutils.NewPod("hello2", ""), podSpec),
				},
				Timeout: 2,
			},
		},
		Client: func(bool) (client.Client, error) {
			return testenv.Client, nil
		},
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) {
			return testenv.DiscoveryClient, nil
		},
	}

	c.Run(t, &report.Testcase{})
}
