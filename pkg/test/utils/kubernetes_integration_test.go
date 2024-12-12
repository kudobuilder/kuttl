//go:build integration

package utils

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

func TestRunCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hcmd := harness.Command{
		Command: "echo 'hello'",
	}

	logger := NewTestLogger(t, "")
	// assert foreground cmd returns nil
	cmd, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.NoError(t, err)
	assert.Nil(t, cmd)
	// foreground processes should have stdout
	assert.NotEmpty(t, stdout)

	hcmd.Background = true
	stdout = &bytes.Buffer{}

	// assert background cmd returns process
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	// no stdout for background processes
	assert.Empty(t, strings.TrimSpace(stdout.String()))

	stdout = &bytes.Buffer{}
	hcmd.Background = false
	hcmd.Command = "sleep 42"

	// assert foreground cmd times out
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 2, "")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "timeout"))
	assert.Nil(t, cmd)

	stdout = &bytes.Buffer{}
	hcmd.Background = false
	hcmd.Command = "sleep 42"
	hcmd.Timeout = 2

	// assert foreground cmd times out with command timeout
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "timeout"))
	assert.Nil(t, cmd)
}

func TestRunCommandIgnoreErrors(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hcmd := harness.Command{
		Command:       "sleep -u",
		IgnoreFailure: true,
	}

	logger := NewTestLogger(t, "")
	// assert foreground cmd returns nil
	cmd, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.NoError(t, err)
	assert.Nil(t, cmd)

	hcmd.IgnoreFailure = false
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.Error(t, err)
	assert.Nil(t, cmd)

	// bad commands have errors regardless of ignore setting
	hcmd = harness.Command{
		Command:       "bad-command",
		IgnoreFailure: true,
	}
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestRunCommandSkipLogOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hcmd := harness.Command{
		Command: "echo 'test'",
	}

	logger := NewTestLogger(t, "")
	// test there is a stdout
	cmd, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.NoError(t, err)
	assert.Nil(t, cmd)
	assert.True(t, stdout.Len() > 0)

	hcmd.SkipLogOutput = true
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	// test there is no stdout
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, logger, 0, "")
	assert.NoError(t, err)
	assert.Nil(t, cmd)
	assert.True(t, stdout.Len() == 0)
}
