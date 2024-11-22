package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/google/cel-go/cel"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

// RunAssertExpressions evaluates a set of CEL expressions specified as AnyAllExpressions
func RunAssertExpressions(
	ctx context.Context,
	logger Logger,
	resourceRefs []harness.TestResourceRef,
	expressions harness.AnyAllExpressions,
	kubeconfigOverride string,
) []error {
	errs := []error{}

	actualDir, err := os.Getwd()
	if err != nil {
		return []error{fmt.Errorf("failed to get current working director: %w", err)}
	}

	kubeconfig := kubeconfigPath(actualDir, kubeconfigOverride)
	cl, err := NewClient(kubeconfig, "")(false)
	if err != nil {
		return []error{fmt.Errorf("failed to construct client: %w", err)}
	}

	variables := make(map[string]interface{})
	for _, resourceRef := range resourceRefs {
		gvk := constructGVK(resourceRef.ApiVersion, resourceRef.Kind)
		referencedResource := &unstructured.Unstructured{}
		referencedResource.SetGroupVersionKind(gvk)

		if err := cl.Get(
			ctx,
			types.NamespacedName{Namespace: resourceRef.Namespace, Name: resourceRef.Name},
			referencedResource,
		); err != nil {
			return []error{fmt.Errorf("failed to get referenced resource '%v': %w", gvk, err)}
		}

		variables[resourceRef.Id] = referencedResource.Object
	}

	env, err := cel.NewEnv()
	if err != nil {
		return []error{fmt.Errorf("failed to create environment: %w", err)}
	}

	for k := range variables {
		env, err = env.Extend(cel.Variable(k, cel.DynType))
		if err != nil {
			return []error{fmt.Errorf("failed to add resource parameter '%v' to environment: %w", k, err)}
		}
	}

	for _, expr := range expressions.Any {
		ast, issues := env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return []error{fmt.Errorf("type-check error: %s", issues.Err())}
		}

		prg, err := env.Program(ast)
		if err != nil {
			return []error{fmt.Errorf("program constuction error: %w", err)}
		}

		out, _, err := prg.Eval(variables)
		if err != nil {
			return []error{fmt.Errorf("failed to evaluate program: %w", err)}
		}

		logger.Logf("expression '%v' evaluated to '%v'", expr, out.Value())
		if out.Value() != true {
			errs = append(errs, fmt.Errorf("failed validation, expression '%v' evaluated to '%v'", expr, out.Value()))
		}
	}

	return errs
}
