//go:build integration

package test

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

func TestHarnessRunIntegration(t *testing.T) {
	harness := Harness{
		TestSuite: harness.TestSuite{
			TestDirs: []string{
				"./test_data/",
			},
			StartControlPlane: true,
			CRDDir:            "./test_crds/",
		},
		T: t,
	}
	harness.Run()
}

func TestHarnessRunIntegrationWithConfig(t *testing.T) {
	testenv, err := testutils.StartTestEnvironment(nil, false)
	if err != nil {
		t.Fatalf("fatal error starting environment: %s", err)
	}
	harness := Harness{
		TestSuite: harness.TestSuite{
			TestDirs: []string{
				"./test_data/",
			},
			// set as true to skip service account check
			StartControlPlane: true,
			Config:            testenv.Config,
			CRDDir:            "./test_crds/",
		},
		T: t,
	}
	harness.Run()
	if err := testenv.Environment.Stop(); err != nil {
		t.Log("error tearing down mock control plane", err)
	}
}

// This test requires external KinD support to run thus is an integration test
func TestRunBackgroundCommands(t *testing.T) {
	h := Harness{
		T: t,
	}
	h.TestSuite.StartControlPlane = true
	commands := []harness.Command{{
		Command:    "sleep 1000000",
		Background: true,
	}}
	h.TestSuite.Commands = commands

	h.Setup()
	defer h.Stop()

	// setup creates bg processes
	assert.Equal(t, 1, len(h.bgProcesses))
	// process is alive
	assert.NoError(t, h.bgProcesses[0].Process.Signal(syscall.Signal(0)))

	// cleans up bg processes
	h.Stop()
	assert.Error(t, h.bgProcesses[0].Process.Signal(syscall.Signal(0)))
}
