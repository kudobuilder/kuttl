package impersonation

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

var ImpersonateAs = ""

func Client(_ bool) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	if ImpersonateAs != "" {
		cfg.Impersonate.UserName = ImpersonateAs
		}

	client, err := testutils.NewRetryClient(cfg, client.Options{
		Scheme: testutils.Scheme(),
	})
	if err != nil {
		return nil, fmt.Errorf("fatal error getting client: %v", err)
	}
	return client, nil
}

func DiscoveryClient() (discovery.DiscoveryInterface, error) {
	cfg, err := config.GetConfig()

	if ImpersonateAs != "" {
		cfg.Impersonate.UserName = ImpersonateAs
	}

	if err != nil {
		return nil, err
	}
	dclient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("fatal error getting discovery client: %v", err)
	}
	return dclient, nil
}


