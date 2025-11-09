package files

import (
	"os"
	"path/filepath"

	kfile "github.com/kudobuilder/kuttl/internal/file"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

// CollectTestStepFiles collects a map of test steps and their associated files
// from a directory.
func CollectTestStepFiles(dir string, logger testutils.Logger) (map[int64][]string, error) {
	testStepFiles := map[int64][]string{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f := kfile.Parse(file.Name())
		if f.Type == kfile.TypeUnknown {
			logger.Logf("Ignoring %q: %v.", file.Name(), f.Error)
			continue
		}
		if !f.HasIndex {
			logger.Logf("Ignoring %q: does not begin with a number followed by a dash.", file.Name())
			continue
		}

		var names []string
		testStepPath := filepath.Join(dir, file.Name())

		if file.IsDir() {
			testStepDir, err := os.ReadDir(testStepPath)
			if err != nil {
				return nil, err
			}

			for _, testStepFile := range testStepDir {
				names = append(names, filepath.Join(testStepPath, testStepFile.Name()))
			}
		} else {
			names = append(names, testStepPath)
		}
		testStepFiles[f.Index] = append(testStepFiles[f.Index], names...)
	}

	return testStepFiles, nil
}
