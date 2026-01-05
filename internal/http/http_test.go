package http //nolint:revive

import (
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			"path to folder",
			"/opt/foo",
			false,
		},
		{
			"path to file",
			"/opt/foo.txt",
			false,
		},
		{
			"http to file",
			"http://kuttl.dev/foo.txt",
			true,
		},
		{
			"https to file",
			"https://kuttl.dev/foo.txt",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.path); got != tt.want {
				t.Errorf("IsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
