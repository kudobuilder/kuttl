package expressions

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/test"
)

func buildProgram(expr string, env *cel.Env) (cel.Program, error) {
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("type-check error: %s", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program construction error: %w", err)
	}

	return prg, nil
}

func buildEnv(resourceRefs []harness.TestResourceRef) (*cel.Env, error) {
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

// RunAssertExpressions evaluates a set of CEL expressions specified as AnyAllExpressions
func RunAssertExpressions(
	programs map[string]cel.Program,
	variables map[string]interface{},
	assertAny,
	assertAll []*harness.Assertion,
) []error {
	var errs []error
	if len(assertAny) == 0 && len(assertAll) == 0 {
		return errs
	}

	var anyExpressionsEvaluation, allExpressionsEvaluation []error
	for _, expr := range assertAny {
		prg, ok := programs[expr.CELExpression]
		if !ok {
			return []error{fmt.Errorf("couldn't find pre-built program for expression: %v", expr.CELExpression)}
		}
		out, _, err := prg.Eval(variables)
		if err != nil {
			return []error{fmt.Errorf("failed to evaluate program: %w", err)}
		}

		if out.Value() != true {
			anyExpressionsEvaluation = append(anyExpressionsEvaluation, fmt.Errorf("expression '%v' evaluated to '%v'", expr.CELExpression, out.Value()))
		}
	}

	for _, expr := range assertAll {
		prg, ok := programs[expr.CELExpression]
		if !ok {
			return []error{fmt.Errorf("couldn't find pre-built program for expression: %v", expr.CELExpression)}
		}
		out, _, err := prg.Eval(variables)
		if err != nil {
			return []error{fmt.Errorf("failed to evaluate program: %w", err)}
		}

		if out.Value() != true {
			allExpressionsEvaluation = append(allExpressionsEvaluation, fmt.Errorf("expression '%v' evaluated to '%v'", expr.CELExpression, out.Value()))
		}
	}

	if len(assertAny) != 0 && len(anyExpressionsEvaluation) == len(assertAny) {
		errs = append(errs, fmt.Errorf("no expression evaluated to true: %w", errors.Join(anyExpressionsEvaluation...)))
	}

	if len(allExpressionsEvaluation) != len(assertAll) {
		errs = append(errs, fmt.Errorf("not all expressions evaluated to true: %w", errors.Join(allExpressionsEvaluation...)))
	}

	return errs
}

func LoadPrograms(s *test.Step) error {
	var errs []error
	for _, resourceRef := range s.Assert.ResourceRefs {
		if err := resourceRef.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("validation failed for reference '%v': %w", resourceRef.String(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to load resource reference(s): %w", errors.Join(errs...))
	}

	var assertions []*harness.Assertion
	assertions = append(assertions, s.Assert.AssertAny...)
	assertions = append(assertions, s.Assert.AssertAll...)

	env, err := buildEnv(s.Assert.ResourceRefs)
	if err != nil {
		return fmt.Errorf("failed to build environment: %w", err)
	}

	if len(assertions) > 0 {
		s.Programs = make(map[string]cel.Program)
	}

	for _, assertion := range assertions {
		if prg, err := buildProgram(assertion.CELExpression, env); err != nil {
			errs = append(errs, err)
		} else {
			s.Programs[assertion.CELExpression] = prg
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to build program(s): %w", errors.Join(errs...))
	}

	return nil
}
