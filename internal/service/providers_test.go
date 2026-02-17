package service

import (
	"testing"
)

func TestProvide_NotNil(t *testing.T) {
	op := Provide()
	if op == nil {
		t.Fatal("Provide returned nil")
	}
}
