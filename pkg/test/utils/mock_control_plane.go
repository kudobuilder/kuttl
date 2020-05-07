package utils

import (
	"bufio"
	"os"
)

// ReadMockControllerConfig reads the mock control plane config file
func ReadMockControllerConfig(filename string) (config []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		config = append(config, scanner.Text())
	}
	err = scanner.Err()
	return
}
