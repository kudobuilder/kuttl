package testcase

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

type clientWithKubeConfig struct {
	client.Client
	kubeConfigPath string
	logger         testutils.Logger
}

// Logf behaves like logger.Logf, but potentially appends a note about which kubeconfig is being used.
// See also getKubeConfigInfo.
func (cl clientWithKubeConfig) Logf(format string, args ...any) {
	cl.logger.Log(fmt.Sprintf(format, args...) + getKubeConfigInfo(cl.kubeConfigPath))
}

// Wrapf returns an error based on format and args, potentially with the addition of a note about
// which kubeconfig is being used, and wrapping err.
// Note: if err is nil, returns nil.
func (cl clientWithKubeConfig) Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s%s: %w", msg, getKubeConfigInfo(cl.kubeConfigPath), err)
}

// getKubeConfigInfo returns a note about kubeConfig (with a space prepended), unless empty.
func getKubeConfigInfo(kubeConfigPath string) string {
	if kubeConfigPath == "" {
		return ""
	}
	return fmt.Sprintf(" (using kubeconfig %q)", kubeConfigPath)
}
