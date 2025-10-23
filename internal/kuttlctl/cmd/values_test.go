package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVars(t *testing.T) {
	tests := map[string]struct {
		input         map[string]string
		expected      map[string]any
		expectError   bool
		errorContains string
	}{
		"empty input": {
			input:    map[string]string{},
			expected: map[string]any{},
		},
		"simple string values": {
			input: map[string]string{
				"name":        "test-name",
				"description": "A test description",
				"empty":       "",
			},
			expected: map[string]any{
				"name":        "test-name",
				"description": "A test description",
				"empty":       nil,
			},
		},
		"numeric values": {
			input: map[string]string{
				"integer":    "42",
				"float":      "3.14159",
				"negative":   "-100",
				"zero":       "0",
				"scientific": "1.23e-4",
			},
			expected: map[string]any{
				"integer":    42,
				"float":      3.14159,
				"negative":   -100,
				"zero":       0,
				"scientific": 1.23e-4,
			},
		},
		"boolean values": {
			input: map[string]string{
				"true_lower":  "true",
				"false_lower": "false",
				"true_title":  "True",
				"false_title": "False",
			},
			expected: map[string]any{
				"true_lower":  true,
				"false_lower": false,
				"true_title":  true,
				"false_title": false,
			},
		},
		"null values": {
			input: map[string]string{
				"null_lower": "null",
				"null_title": "Null",
				"tilde":      "~",
			},
			expected: map[string]any{
				"null_lower": nil,
				"null_title": nil,
				"tilde":      nil,
			},
		},
		"array values": {
			input: map[string]string{
				"simple_array": "[1, 2, 3]",
				"string_array": `["apple", "banana", "cherry"]`,
				"mixed_array":  `[1, "two", true, null]`,
				"empty_array":  "[]",
			},
			expected: map[string]any{
				"simple_array": []interface{}{1, 2, 3},
				"string_array": []interface{}{"apple", "banana", "cherry"},
				"mixed_array":  []interface{}{1, "two", true, nil},
				"empty_array":  []interface{}{},
			},
		},
		"object values": {
			input: map[string]string{
				"simple_object": `{"name": "test", "value": 42}`,
				"empty_object":  "{}",
			},
			expected: map[string]any{
				"simple_object": map[string]interface{}{"name": "test", "value": 42},
				"empty_object":  map[string]interface{}{},
			},
		},
		"YAML flow syntax": {
			input: map[string]string{
				"flow_array":  "[a, b, c]",
				"flow_object": "{name: test, count: 5, enabled: true}",
			},
			expected: map[string]any{
				"flow_array":  []interface{}{"a", "b", "c"},
				"flow_object": map[string]interface{}{"name": "test", "count": 5, "enabled": true},
			},
		},
		"multiline YAML": {
			input: map[string]string{
				"multiline": "name: test\nversion: 0.1\ntags:\n  - v1\n  - stable",
			},
			expected: map[string]any{
				"multiline": map[string]interface{}{
					"name":    "test",
					"version": 0.1,
					"tags":    []interface{}{"v1", "stable"},
				},
			},
		},
		"unicode and special characters": {
			input: map[string]string{
				"unicode":        "测试",
				"emoji":          "🚀",
				"special_chars":  "name@domain.com",
				"with_quotes":    `"quoted string"`,
				"with_backslash": `path\to\file`,
			},
			expected: map[string]any{
				"unicode":        "测试",
				"emoji":          "🚀",
				"special_chars":  "name@domain.com",
				"with_quotes":    "quoted string",
				"with_backslash": `path\to\file`,
			},
		},
		"unclosed bracket": {
			input: map[string]string{
				"broken": "[1, 2, 3",
			},
			expectError:   true,
			errorContains: `failed to parse value of "broken" as YAML`,
		},
		"unclosed brace": {
			input: map[string]string{
				"broken": "{name: test",
			},
			expectError:   true,
			errorContains: `failed to parse value of "broken" as YAML`,
		},
		"invalid YAML syntax": {
			input: map[string]string{
				"broken": "name: test\n  invalid: : value",
			},
			expectError:   true,
			errorContains: `failed to parse value of "broken" as YAML`,
		},
		"mixed valid and invalid": {
			input: map[string]string{
				"valid":   "42",
				"invalid": "[1, 2, 3",
			},
			expectError:   true,
			errorContains: `failed to parse value of "invalid" as YAML`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := parseVars(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
