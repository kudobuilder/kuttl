package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimLeftChar(t *testing.T) {
	testMap := map[string]string{
		"":           "",
		"$":          "",
		"$NAMESPACE": "NAMESPACE",
		"NAMESPACE":  "AMESPACE",
	}
	for k, v := range testMap {
		assert.Equal(t, v, TrimLeftChar(k))
	}
}
