package cmd

import (
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/labels"
)

type labelSetValue labels.Set

func (v *labelSetValue) String() string {
	return labels.Set(*v).String()
}

func (v *labelSetValue) Set(s string) error {
	l, err := labels.ConvertSelectorToLabelsMap(s)
	if err != nil {
		return fmt.Errorf("cannot parse label set: %w", err)
	}
	*v = labelSetValue(l)
	return nil
}

func (v *labelSetValue) Type() string {
	return "labelSet"
}

func (v *labelSetValue) AsLabelSet() labels.Set {
	return labels.Set(*v)
}

func parseVars(varsStrings map[string]string) (map[string]any, error) {
	parsedVars := map[string]any{}
	for k, v := range varsStrings {
		var val any
		err := yaml.Unmarshal([]byte(v), &val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value of %q as YAML: %w", k, err)
		}
		parsedVars[k] = val
	}
	return parsedVars, nil
}
