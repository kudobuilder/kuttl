package utils //nolint:revive,nolintlint // apparently nolintlint is confused

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSubset(t *testing.T) {
	require.NoError(t, IsSubset(map[string]any{
		"hello": "world",
	}, map[string]any{
		"hello": "world",
		"bye":   "moon",
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": "moon",
	}, map[string]any{
		"hello": "world",
		"bye":   "moon",
	}))

	require.NoError(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": "world",
		},
	}, map[string]any{
		"hello": map[string]any{
			"hello": "world",
			"bye":   "moon",
		},
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": "moon",
		},
	}, map[string]any{
		"hello": map[string]any{
			"hello": "world",
			"bye":   "moon",
		},
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": "moon",
		},
	}, map[string]any{
		"hello": "world",
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": "world",
	}, map[string]any{}))

	require.NoError(t, IsSubset(map[string]any{
		"hello": []int{
			1, 2, 3,
		},
	}, map[string]any{
		"hello": []int{
			1, 2, 3,
		},
	}))

	require.NoError(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "hello",
					"bye":   "moon",
				},
			},
		},
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "hello",
					"bye":   "moon",
				},
				{
					"bye": "moon",
				},
			},
		},
	}))

	assert.NotNil(t, IsSubset(map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]any{
		"hello": map[string]any{
			"hello": []map[string]any{
				{
					"image": "world",
				},
			},
		},
	}))
}
