package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvedDeletePolicy(t *testing.T) {
	tests := map[string]struct {
		suite    TestSuite
		expected DeletePolicy
	}{
		"default (nothing set) → all": {
			suite:    TestSuite{},
			expected: DeleteAll,
		},
		"Delete=all → all": {
			suite:    TestSuite{Delete: DeleteAll},
			expected: DeleteAll,
		},
		"Delete=success → success": {
			suite:    TestSuite{Delete: DeleteSuccess},
			expected: DeleteSuccess,
		},
		"Delete=none → none": {
			suite:    TestSuite{Delete: DeleteNone},
			expected: DeleteNone,
		},
		"SkipDelete=true (Delete unset) → none": {
			suite:    TestSuite{SkipDelete: true},
			expected: DeleteNone,
		},
		"SkipDelete=false (Delete unset) → all": {
			suite:    TestSuite{SkipDelete: false},
			expected: DeleteAll,
		},
		"Delete=success takes precedence over SkipDelete=true": {
			suite:    TestSuite{Delete: DeleteSuccess, SkipDelete: true},
			expected: DeleteSuccess,
		},
		"Delete=all takes precedence over SkipDelete=true": {
			suite:    TestSuite{Delete: DeleteAll, SkipDelete: true},
			expected: DeleteAll,
		},
		"invalid Delete value falls back to SkipDelete=true → none": {
			suite:    TestSuite{Delete: "unknown", SkipDelete: true},
			expected: DeleteNone,
		},
		"invalid Delete value falls back to SkipDelete=false → all": {
			suite:    TestSuite{Delete: "unknown", SkipDelete: false},
			expected: DeleteAll,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.suite.ResolvedDeletePolicy())
		})
	}
}
