package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromPath(t *testing.T) {

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected []string
		wantErr  bool
	}{
		{
			name:     `good path no extension`,
			path:     "testdata/path",
			pattern:  "",
			expected: []string{"testdata/path/skip.txt", "testdata/path/test1.yaml", "testdata/path/test2.yaml"},
			wantErr:  false,
		},
		{
			name:     `good path yaml extension`,
			path:     "testdata/path",
			pattern:  "*.yaml",
			expected: []string{"testdata/path/test1.yaml", "testdata/path/test2.yaml"},
			wantErr:  false,
		},
		{
			name:     `bad path`,
			path:     "testdata/badpath",
			pattern:  "",
			expected: []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			paths, err := FromPath(tt.path, tt.pattern)
			assert.Equal(t, tt.wantErr, err != nil, "expected error %v, but got %v", tt.wantErr, err)
			assert.ElementsMatch(t, paths, tt.expected)
		})
	}
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

func TestTrimExt(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{name: "standard tar", path: "foo.tar", expected: "foo"},
		{name: "tar in path", path: "in/a/path/foo.tar", expected: "in/a/path/foo"},
		{name: "tgz in path", path: "in/a/path/foo.tgz", expected: "in/a/path/foo"},
		{name: "non-supported tar.gz", path: "in/a/path/foo.tar.gz", expected: "in/a/path/foo.tar"}, // we don't support this file format
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			path := TrimExt(tt.path)
			assert.Equal(t, path, tt.expected)
		})
	}
}
