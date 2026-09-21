package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	harnessapi "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

// TestReportFormat verifies Report against already-normalized formats: it writes
// the requested format, writes nothing for an empty format, and errors on any
// non-canonical value (case normalization happens at the command-line boundary,
// not here). See issue #449.
func TestReportFormat(t *testing.T) {
	tests := map[string]struct {
		ftype       harnessapi.ReportType
		wantXMLFile bool
		wantJSON    bool
		wantErr     bool
	}{
		"XML":                  {ftype: harnessapi.ReportTypeXML, wantXMLFile: true},
		"JSON":                 {ftype: harnessapi.ReportTypeJSON, wantJSON: true},
		"empty writes nothing": {ftype: harnessapi.ReportTypeNil},
		"lowercase is unknown": {ftype: "xml", wantErr: true},
		"unknown is an error":  {ftype: "foobar", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Use a not-yet-existing subdirectory to verify Report creates it.
			dir := filepath.Join(t.TempDir(), "artifacts")
			suites := &Testsuites{}

			err := suites.Report(dir, "kuttl-report", tt.ftype)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			_, xmlErr := os.Stat(filepath.Join(dir, "kuttl-report.xml"))
			_, jsonErr := os.Stat(filepath.Join(dir, "kuttl-report.json"))

			assert.Equal(t, tt.wantXMLFile, xmlErr == nil, "XML report presence")
			assert.Equal(t, tt.wantJSON, jsonErr == nil, "JSON report presence")

			// The artifacts subdirectory is created whenever a report is attempted
			// (i.e. for any non-empty format), and not for the empty "no report" case.
			_, dirErr := os.Stat(dir)
			assert.Equal(t, tt.ftype != harnessapi.ReportTypeNil, dirErr == nil, "subdirectory creation")
		})
	}
}
