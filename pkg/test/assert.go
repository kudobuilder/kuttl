package test

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// Assert
func Assert(assertFile, namespace string, timeout int) error {

	info, err := os.Stat(assertFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("the file %q does not exist", assertFile)
	}
	if info.IsDir() {
		return fmt.Errorf("%q is a directory and not a file", assertFile)
	}
	// feels like the wrong abstraction, need to do some refactoring
	s := &Step{
		Timeout:         0,
		Client:          Client,
		DiscoveryClient: DiscoveryClient,
	}

	objects, err := testutils.LoadYAMLFromFile(assertFile)
	if err != nil {
		return err
	}

	var testErrors []error
	for i := 0; i < timeout; i++ {
		// start fresh
		testErrors = []error{}
		for _, expected := range objects {
			testErrors = append(testErrors, s.CheckResource(expected, namespace)...)
		}

		if len(testErrors) == 0 {
			break
		}

		time.Sleep(time.Second)
	}

	if len(testErrors) == 0 {
		fmt.Printf("%q in the %q namespace is valid\n", assertFile, namespace)
		return nil
	}

	for _, testError := range testErrors {
		fmt.Println(testError)
	}
	return fmt.Errorf("%q in the %q namespace is not valid", assertFile, namespace)
}

func Client(forceNew bool) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}
	dclient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("fatal error getting discovery client: %v", err)
	}
	return dclient, nil
}
