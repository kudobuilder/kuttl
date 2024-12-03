package utils

import (
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

func BuildEnv(resourceRefs []v1beta1.TestResourceRef) (*cel.Env, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	for _, resourceRef := range resourceRefs {
		env, err = env.Extend(cel.Variable(resourceRef.Ref, cel.DynType))
		if err != nil {
			return nil, fmt.Errorf("failed to add resource parameter '%v' to environment: %w", resourceRef.Ref, err)
		}
	}

	return env, nil
}
