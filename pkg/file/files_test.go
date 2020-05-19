package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromPath(t *testing.T) {

	paths, err := FromPath("testdata/path")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(paths))
	assert.Equal(t, "testdata/path/test1.yaml", paths[0])

	_, err = FromPath("testdata/badpath")
	assert.Error(t, err, "file mode issue with stat testdata/badpath: no such file or directory")
}

func TestToRuntimeObjects(t *testing.T) {
	files := []string{"testdata/path/test1.yaml"}
	objs, err := ToRuntimeObjects(files)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(objs))
	assert.Equal(t, "Pod", objs[0].GetObjectKind().GroupVersionKind().Kind)

	files = append(files, "testdata/path/test2.yaml")
	_, err = ToRuntimeObjects(files)
	assert.Error(t, err, "file \"testdata/path/test2.yaml\" load yaml error")
}
