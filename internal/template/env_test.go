package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_Clone(t *testing.T) {
	tests := map[string]Env{
		"empty": {},
		"just ns": {
			Namespace: "foo",
		},
		"vars": {
			Namespace: "foo",
			Vars: map[string]any{
				"foo":    "foo",
				"number": 1.1,
				"slice":  []any{float64(1), float64(2), float64(3), "bar"},
				"map": map[string]any{
					"a": float64(1),
					"b": float64(2),
					"c": "baz",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.Clone()
			require.NoError(t, err)
			assert.Equal(t, tt, got)
		})
	}
}
