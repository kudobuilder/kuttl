package files

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

// mockLogger is a simple logger that captures log messages for testing
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Log(args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprint(args...))
}

func (m *mockLogger) Logf(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) WithPrefix(_ string) testutils.Logger {
	return m
}

func (m *mockLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockLogger) Flush() {}

func (m *mockLogger) hasMessageContaining(substring string) bool {
	for _, msg := range m.messages {
		if strings.Contains(msg, substring) {
			return true
		}
	}
	return false
}

func TestCollectTestStepFiles(t *testing.T) {
	for _, tt := range []struct {
		path     string
		expected map[int64][]string
	}{
		{
			"test_data/with-overrides",
			map[int64][]string{
				int64(0): {
					"test_data/with-overrides/00-assert.yaml",
					"test_data/with-overrides/00-test-step.yaml",
				},
				int64(1): {
					"test_data/with-overrides/01-assert.yaml",
					"test_data/with-overrides/01-test-assert.yaml",
				},
				int64(2): {
					"test_data/with-overrides/02-directory/assert.yaml",
					"test_data/with-overrides/02-directory/pod.yaml",
					"test_data/with-overrides/02-directory/pod2.yaml",
				},
				int64(3): {
					"test_data/with-overrides/03-assert.yaml",
					"test_data/with-overrides/03-pod.yaml",
					"test_data/with-overrides/03-pod2.yaml",
				},
			},
		},
		{
			"test_data/list-pods",
			map[int64][]string{
				int64(0): {
					"test_data/list-pods/00-assert.yaml",
					"test_data/list-pods/00-pod.yaml",
				},
			},
		},
	} {
		t.Run(tt.path, func(t *testing.T) {
			testStepFiles, err := CollectTestStepFiles(tt.path, testutils.NewTestLogger(t, tt.path), nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, testStepFiles)
		})
	}
}

func TestCollectTestStepFilesWithIgnorePatterns(t *testing.T) {
	t.Run("default patterns ignore README files", func(t *testing.T) {
		logger := &mockLogger{}
		_, err := CollectTestStepFiles("test_data/with-overrides", logger, nil)
		require.NoError(t, err)

		assert.False(t, logger.hasMessageContaining("Ignoring \"README.md\""),
			"README.md should be silently ignored with default patterns")
	})

	t.Run("explicit patterns override defaults", func(t *testing.T) {
		logger := &mockLogger{}
		_, err := CollectTestStepFiles("test_data/with-overrides", logger, []string{})
		require.NoError(t, err)

		assert.True(t, logger.hasMessageContaining("Ignoring \"README.md\""),
			"README.md should generate warning when default patterns are overridden with empty list")
	})

	t.Run("custom patterns silently ignore matching files", func(t *testing.T) {
		logger := &mockLogger{}
		_, err := CollectTestStepFiles("test_data/with-overrides", logger, []string{"*.txt", "*.md"})
		require.NoError(t, err)

		assert.False(t, logger.hasMessageContaining("Ignoring \"README.md\""),
			"README.md should be silently ignored with matching custom pattern")
	})
}
