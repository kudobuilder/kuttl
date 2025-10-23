package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Type int

const (
	// TypeUnknown means it was not possible to determine the type of file, e.g. unrecognized name pattern.
	TypeUnknown Type = iota
	// TypeApply denotes files with resources to apply on the cluster.
	TypeApply
	// TypeAssert denotes assertion files. Must match for the step to pass.
	TypeAssert
	// TypeError denotes negative assertion files. Must not match for the step to pass.
	TypeError
)

type Info struct {
	Type Type
	// Error is set for TypeUnknown objects and describes the reason.
	Error    error
	BaseName string
	FullName string
	// HasIndex is true when the file name starts with a valid index.
	HasIndex bool
	Index    int64
	StepName string
}

// fileNameRegex is used to parse both:
// - names of entries (files or directories) contained directly in a test case directory, and
// - name of files contained in a test step directory
// Apart from the extension separated with a dot, the groups are separated by dashes and are:
//   - optional numeric prefix - required only for entries directly in the test case directory
//   - first (or only) name component, it is special in that if it's equal to "assert" or "error" it denotes
//     the file as an assert or error file, respectively.
//   - optional additional components separated by dashes
var fileNameRegex = regexp.MustCompile(`^(\d+-)?([^-.]+)(-[^.]+)?(?:\.yaml)?$`)

// fileNamePattern is a human-readable representation of fileNameRegex.
const fileNamePattern = "(<number>-)<name>(-<name>)(.yaml)"

func Parse(fullName string) Info {
	name := filepath.Base(fullName)
	matches := fileNameRegex.FindStringSubmatch(name)
	if len(matches) < 3 {
		return Info{
			Type:     TypeUnknown,
			BaseName: name,
			FullName: fullName,
			Error:    fmt.Errorf("name does not follow pattern %q", fileNamePattern),
		}
	}

	var i int64
	var hasIndex bool
	if matches[1] != "" {
		var err error
		i, err = strconv.ParseInt(strings.TrimSuffix(matches[1], "-"), 10, 32)
		if err != nil {
			return Info{
				Type:     TypeUnknown,
				BaseName: name,
				FullName: fullName,
				Error:    fmt.Errorf("parsing index failed: %w", err),
			}
		}
		hasIndex = true
	}

	switch fname := strings.ToLower(matches[2]); fname {
	case "assert":
		return Info{
			Type:     TypeAssert,
			BaseName: name,
			FullName: fullName,
			HasIndex: hasIndex,
			Index:    i,
		}
	case "errors":
		return Info{
			Type:     TypeError,
			BaseName: name,
			FullName: fullName,
			HasIndex: hasIndex,
			Index:    i,
		}
	default:
		var stepName string
		if len(matches) > 3 {
			// The second matching group will already have a hyphen prefix.
			stepName = matches[2] + matches[3]
		} else {
			stepName = matches[2]
		}
		return Info{
			Type:     TypeApply,
			BaseName: name,
			FullName: fullName,
			HasIndex: hasIndex,
			Index:    i,
			StepName: stepName,
		}
	}
}
