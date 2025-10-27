package testcase

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

type clientWithPath struct {
	client.Client
	kubeConfigPath string
	logger         testutils.Logger
}

// Logf behaves like logger.Logf, but potentially appends a note about which kubeconfig is being used.
// See also getKubeConfigInfo.
func (cl clientWithPath) Logf(format string, args ...any) {
	cl.logger.Log(fmt.Sprintf(format, args...) + getKubeConfigInfo(cl.kubeConfigPath))
}

// getKubeConfigInfo returns a note about kubeConfig (with a space prepended), unless empty.
func getKubeConfigInfo(kubeConfigPath string) string {
	if kubeConfigPath == "" {
		return ""
	}
	return fmt.Sprintf(" (using kubeconfig %q)", kubeConfigPath)
}
