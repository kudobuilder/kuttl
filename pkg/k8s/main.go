package k8s

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var ImpersonateAs = ""

func GetConfig() (*rest.Config, error) {
	cfg, err := config.GetConfig()
	
	if err != nil {
		return &rest.Config{}, err
	}

	if ImpersonateAs != "" {
		cfg.Impersonate.UserName = ImpersonateAs
	}

	return cfg, nil
}
