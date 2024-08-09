package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		name     string
		ver      string
		expected bool
	}{
		{
			name:     `current version`,
			ver:      "kind.sigs.k8s.io/v1alpha4",
			expected: true,
		},
		{
			name:     `early version`,
			ver:      "kind.sigs.k8s.io/v1alpha3",
			expected: false,
		},
		{
			name:     `newer version`,
			ver:      "kind.sigs.k8s.io/v1beta1",
			expected: true,
		},
		{
			name:     `wrong group`,
			ver:      "foo/v1alpha4",
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := IsMinVersion(tt.ver)
			assert.Equal(t, tt.expected, result)
		})
	}
}
