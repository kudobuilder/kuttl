package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSubset(t *testing.T) {
	require.NoError(t, IsSubset(map[string]interface{}{
		"hello": "world",
	}, map[string]interface{}{
		"hello": "world",
		"bye":   "moon",
	}))

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": "moon",
	}, map[string]interface{}{
		"hello": "world",
		"bye":   "moon",
	}))

	require.NoError(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": "world",
		},
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": "world",
			"bye":   "moon",
		},
	}))

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": "moon",
		},
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": "world",
			"bye":   "moon",
		},
	}))

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": "moon",
		},
	}, map[string]interface{}{
		"hello": "world",
	}))

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": "world",
	}, map[string]interface{}{}))

	require.NoError(t, IsSubset(map[string]interface{}{
		"hello": []int{
			1, 2, 3,
		},
	}, map[string]interface{}{
		"hello": []int{
			1, 2, 3,
		},
	}))

	require.NoError(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
				{
					"image": "hello",
					"bye":   "moon",
				},
			},
		},
	}))

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
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

	assert.NotNil(t, IsSubset(map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
				{
					"image": "hello",
				},
			},
		},
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"hello": []map[string]interface{}{
				{
					"image": "world",
				},
			},
		},
	}))
}
