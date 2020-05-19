package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// from a list of paths, returns an array of runtime objects
func ToRuntimeObjects(paths []string) ([]runtime.Object, error) {
	apply := []runtime.Object{}

	for _, path := range paths {
		objs, err := testutils.LoadYAML(path)
		if err != nil {
			return nil, fmt.Errorf("file %q load yaml error", path)
		}
		apply = append(apply, objs...)
	}

	return apply, nil
}

// From a file or dir path returns an array of flat file paths
func FromPath(path string) ([]string, error) {
	files := []string{}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file mode issue with %w", err)
	}
	if fi.IsDir() {
		fileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() {
				files = append(files, filepath.Join(path, fileInfo.Name()))
			}
		}
	} else {
		files = append(files, path)
	}

	return files, nil
}
