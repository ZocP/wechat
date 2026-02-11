package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUintParam_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected uint
	}{
		{"1", 1},
		{"0", 0},
		{"100", 100},
		{"4294967295", 4294967295}, // Max uint32
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseUintParam(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseUintParam_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"negative", "-1"},
		{"non-numeric", "abc"},
		{"float", "1.5"},
		{"special chars", "!@#"},
		{"spaces", " 1 "},
		{"overflow uint32", "99999999999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseUintParam(tc.input)
			assert.Error(t, err)
		})
	}
}
