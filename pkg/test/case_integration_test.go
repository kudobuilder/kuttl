// +build integration

package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kudobuilder/kuttl/pkg/report"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// Create two test environments, ensure that the second environment is used when
// Kubeconfig is set on a Step.
func TestMultiClusterCase(t *testing.T) {
	testenv, err := testutils.StartTestEnvironment(testutils.APIServerDefaultArgs, false)
	if err != nil {
		t.Error(err)
		return
	}
	defer testenv.Environment.Stop()

	testenv2, err := testutils.StartTestEnvironment(testutils.APIServerDefaultArgs, false)
	if err != nil {
		t.Error(err)
		return
	}
	defer testenv2.Environment.Stop()

	podSpec := map[string]interface{}{
		"restartPolicy": "Never",
		"containers": []map[string]interface{}{
			{
				"name":  "nginx",
				"image": "nginx:1.7.9",
			},
		},
	}

	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tmpfile.Name())
	if err := testutils.Kubeconfig(testenv2.Config, tmpfile); err != nil {
		t.Error(err)
		return
	}

	c := Case{
		Logger: testutils.NewTestLogger(t, ""),
		Steps: []*Step{
			&Step{
				Name:  "initialize-testenv",
				Index: 0,
				Apply: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Asserts: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Timeout: 2,
			},
			&Step{
				Name:  "use-testenv2",
				Index: 1,
				Apply: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello2", ""), podSpec),
				},
				Asserts: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello2", ""), podSpec),
				},
				Errors: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Timeout:    2,
				Kubeconfig: tmpfile.Name(),
			},
			&Step{
				Name:  "verify-testenv-does-not-have-testenv2-resources",
				Index: 2,
				Asserts: []runtime.Object{
					testutils.WithSpec(t, testutils.NewPod("hello", ""), podSpec),
				},
				Errors: []runtime.Object{
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
