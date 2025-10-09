package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntarInPlace(t *testing.T) {
	tarfile := "testdata/tar-test.tgz"
	err := UntarInPlace(tarfile)
	assert.NoError(t, err)

	folder := "testdata/tar-test"
	defer os.RemoveAll(folder)

	fi, err := os.Stat(folder)
	assert.NoError(t, err)
	assert.True(t, fi.IsDir())
}
