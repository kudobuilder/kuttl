package file

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUntarInPlace(t *testing.T) {
	tarfile := "testdata/tar-test.tgz"

	sandbox := t.TempDir()

	testFile := path.Join(sandbox, "tar-test.tgz")
	data, err := os.ReadFile(tarfile)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(testFile, data, 0600))

	err = UntarInPlace(testFile)
	assert.NoError(t, err)

	folder := path.Join(sandbox, "tar-test")
	fi, err := os.Stat(folder)
	assert.NoError(t, err)
	assert.True(t, fi.IsDir())
}
