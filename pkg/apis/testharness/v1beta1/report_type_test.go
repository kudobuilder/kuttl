package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportTypeNormalize(t *testing.T) {
	tests := map[string]struct {
		in       ReportType
		expected ReportType
	}{
		"lowercase xml":          {in: "xml", expected: ReportTypeXML},
		"uppercase XML":          {in: "XML", expected: ReportTypeXML},
		"mixed case Xml":         {in: "Xml", expected: ReportTypeXML},
		"lowercase json":         {in: "json", expected: ReportTypeJSON},
		"uppercase JSON":         {in: "JSON", expected: ReportTypeJSON},
		"mixed case Json":        {in: "jSoN", expected: ReportTypeJSON},
		"spaces are not trimmed": {in: "  xml  ", expected: "  XML  "},
		"empty stays empty":      {in: "", expected: ReportTypeNone},
		"unknown value upper":    {in: "foo", expected: "FOO"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.in.Normalize())
		})
	}
}

func TestReportTypeValid(t *testing.T) {
	tests := map[string]struct {
		in       ReportType
		expected bool
	}{
		"empty is valid (no report)": {in: ReportTypeNone, expected: true},
		"XML is valid":               {in: ReportTypeXML, expected: true},
		"JSON is valid":              {in: ReportTypeJSON, expected: true},
		"lowercase xml is invalid":   {in: "xml", expected: false},
		"unknown is invalid":         {in: "FOO", expected: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.in.Valid())
		})
	}
}
