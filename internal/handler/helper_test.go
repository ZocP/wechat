package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestParsePagination_DefaultAndCustom(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	p, ps, off := parsePagination(c)
	assert.Equal(t, 1, p)
	assert.Equal(t, 20, ps)
	assert.Equal(t, 0, off)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/?page=2&page_size=200", nil)
	p2, ps2, off2 := parsePagination(c2)
	assert.Equal(t, 2, p2)
	assert.Equal(t, 100, ps2)
	assert.Equal(t, 100, off2)

	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest("GET", "/?page=-1&page_size=abc", nil)
	p3, ps3, off3 := parsePagination(c3)
	assert.Equal(t, 1, p3)
	assert.Equal(t, 20, ps3)
	assert.Equal(t, 0, off3)
}
