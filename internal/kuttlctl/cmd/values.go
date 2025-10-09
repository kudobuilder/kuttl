package cmd

import (
	"fmt"

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
