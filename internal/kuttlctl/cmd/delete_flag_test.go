package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	harnessApi "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

// TestDeleteFlagValidation exercises the --delete flag validation logic.
func TestDeleteFlagValidation(t *testing.T) {
	tests := map[string]struct {
		args           []string
		expectedErr    string
		expectedPolicy harnessApi.DeletePolicy
	}{
		"--delete all is valid": {
			args:           []string{"test", "--delete", "all", "./testdir"},
			expectedPolicy: harnessApi.DeleteAll,
		},
		"--delete success is valid": {
			args:           []string{"test", "--delete", "success", "./testdir"},
			expectedPolicy: harnessApi.DeleteSuccess,
		},
		"--delete none is valid": {
			args:           []string{"test", "--delete", "none", "./testdir"},
			expectedPolicy: harnessApi.DeleteNone,
		},
		"--delete invalid returns error": {
			args:        []string{"test", "--delete", "bogus", "./testdir"},
			expectedErr: `invalid --delete value "bogus": must be one of all, success, none`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := NewKuttlCmd()
			root.SetArgs(tt.args)

			// Capture the options set during PreRunE by running only PreRunE.
			// We find the "test" sub-command and invoke its PreRunE directly.
			var testCmd = root
			for _, sub := range root.Commands() {
				if sub.Use == "test [flags]... [test suite]..." {
					testCmd = sub
					break
				}
			}
			require.NotNil(t, testCmd)

			// Parse the flags so PreRunE can read them.
			require.NoError(t, testCmd.ParseFlags(tt.args[1:]))

			err := testCmd.PreRunE(testCmd, testCmd.Flags().Args())

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
