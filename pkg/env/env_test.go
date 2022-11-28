package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandWithMap(t *testing.T) {
	os.Setenv("KUTTL_TEST_123", "hello")
	t.Cleanup(func() {
		os.Unsetenv("KUTTL_TEST_123")
	})
	assert.Equal(t, "hello $  world", ExpandWithMap("$KUTTL_TEST_123 $$ $DOES_NOT_EXIST_1234 ${EXPAND_ME}", map[string]string{
		"EXPAND_ME": "world",
	}))
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: `do not expand $$`,
			in:   "test $$",
			want: "test $",
		},
		{
			name: `expand os`,
			in:   "$KUTTL_TEST_123 $$",
			want: "hello $",
		},
		{
			name: `expansions with no values, have no output`,
			in:   "$KUTTL_TEST_123 $$ ${NOT_PROVIDED}",
			want: "hello $ ",
		},
	}

	os.Setenv("KUTTL_TEST_123", "hello")
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := Expand(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %q; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
