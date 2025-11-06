package file

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []Info{
		{
			Type:     TypeAssert,
			FullName: "00-assert.yaml",
			BaseName: "00-assert.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "",
		},
		{
			Type:     TypeError,
			FullName: "00-errors.yaml",
			BaseName: "00-errors.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "",
		},
		{
			Type:     TypeApply,
			FullName: "00-foo.yaml",
			BaseName: "00-foo.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "foo",
		},
		{
			Type:     TypeApply,
			FullName: "01-foo.yaml",
			BaseName: "01-foo.yaml",
			HasIndex: true,
			Index:    1,
			StepName: "foo",
		},
		{
			Type:     TypeApply,
			FullName: "01234-foo.yaml",
			BaseName: "01234-foo.yaml",
			HasIndex: true,
			Index:    1234,
			StepName: "foo",
		},
		{
			Type:     TypeApply,
			FullName: "01-foo",
			BaseName: "01-foo",
			HasIndex: true,
			Index:    1,
			StepName: "foo",
		},
		{
			Type:     TypeApply,
			FullName: "1-foo.yaml",
			BaseName: "1-foo.yaml",
			HasIndex: true,
			Index:    1,
			StepName: "foo",
		},
		{
			Type:     TypeApply,
			FullName: "1-foo",
			BaseName: "1-foo",
			HasIndex: true,
			Index:    1,
			StepName: "foo",
		},
		{
			Type:     TypeAssert,
			FullName: "123-assert.yaml",
			BaseName: "123-assert.yaml",
			HasIndex: true,
			Index:    123,
			StepName: "",
		},
		{
			Type:     TypeError,
			FullName: "123-errors.yaml",
			BaseName: "123-errors.yaml",
			HasIndex: true,
			Index:    123,
			StepName: "",
		},
		{
			Type:     TypeApply,
			FullName: "123-foo.yaml",
			BaseName: "123-foo.yaml",
			HasIndex: true,
			Index:    123,
			StepName: "foo",
		},
		{
			Type:     TypeAssert,
			FullName: "00-assert-bar.yaml",
			BaseName: "00-assert-bar.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "",
		},
		{
			Type:     TypeError,
			FullName: "00-errors-bar.yaml",
			BaseName: "00-errors-bar.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "",
		},
		{
			Type:     TypeApply,
			FullName: "00-foo-bar.yaml",
			BaseName: "00-foo-bar.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "foo-bar",
		},
		{
			Type:     TypeApply,
			FullName: "1-foo-bar.yaml",
			BaseName: "1-foo-bar.yaml",
			HasIndex: true,
			Index:    1,
			StepName: "foo-bar",
		},
		{
			Type:     TypeApply,
			FullName: "00-foo-bar-baz.yaml",
			BaseName: "00-foo-bar-baz.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "foo-bar-baz",
		},
		{
			Type:     TypeApply,
			FullName: "01.yaml",
			BaseName: "01.yaml",
			HasIndex: false,
			StepName: "01",
		},
		{
			Type:     TypeApply,
			FullName: "foo-01.yaml",
			BaseName: "foo-01.yaml",
			HasIndex: false,
			StepName: "foo-01",
		},
		{
			Type:     TypeApply,
			FullName: "foo",
			BaseName: "foo",
			HasIndex: false,
			StepName: "foo",
		},
		{
			Type:     TypeAssert,
			FullName: "some/dir/00-assert.yaml",
			BaseName: "00-assert.yaml",
			HasIndex: true,
			Index:    0,
			StepName: "",
		},
		{
			Type:     TypeUnknown,
			FullName: "1.foo.yaml",
			BaseName: "1.foo.yaml",
			Error:    fmt.Errorf("name does not follow pattern %q", fileNamePattern),
		},
		{
			Type:     TypeUnknown,
			FullName: "99999999999999999999999999999999-foo.yaml",
			BaseName: "99999999999999999999999999999999-foo.yaml",
			Error: fmt.Errorf("parsing index failed: %w", &strconv.NumError{
				Func: "ParseInt",
				Num:  "99999999999999999999999999999999",
				Err:  errors.New("value out of range"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.FullName, func(t *testing.T) {
			assert.Equalf(t, tt, Parse(tt.FullName), "Parse(%v)", tt.BaseName)
		})
	}
}
