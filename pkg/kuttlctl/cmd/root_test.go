package cmd

import (
	"reflect"
	"testing"
)

func TestNormalizationFuncGlobalExistence(t *testing.T) {
	root := NewKuttlCmd()

	if root.Parent() != nil {
		t.Fatal("We expect the root command to be returned")
	}

	sub := root
	for sub.HasSubCommands() {
		sub = sub.Commands()[0]
	}

	// In case of failure of this test check this PR: spf13/cobra#110
	if reflect.ValueOf(sub.GlobalNormalizationFunc()).Pointer() != reflect.ValueOf(root.GlobalNormalizationFunc()).Pointer() {
		t.Fatal("child and root commands should have the same normalization functions")
	}
}
