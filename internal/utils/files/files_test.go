package files

import (
	"testing"

	"github.com/stretchr/testify/assert"

	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

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
			testStepFiles, err := CollectTestStepFiles(tt.path, testutils.NewTestLogger(t, tt.path))
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, testStepFiles)
		})
	}
}

func TestGetIndexFromFile(t *testing.T) {
	for _, tt := range []struct {
		fileName string
		indexExp int64
	}{
		{"00-foo.yaml", 0},
		{"01-foo.yaml", 1},
		{"1-foo.yaml", 1},
		{"01-foo", 1},
		{"01234-foo.yaml", 1234},
		{"1-foo-bar.yaml", 1},
		{"01.yaml", -1},
		{"foo-01.yaml", -1},
	} {
		t.Run(tt.fileName, func(t *testing.T) {
			index, err := getIndexFromFile(tt.fileName)
			assert.Nil(t, err)
			assert.Equal(t, tt.indexExp, index)
		})
	}
}
