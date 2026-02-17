package scheduler

import "testing"

func TestProvide(t *testing.T) {
	op := Provide()
	if op == nil {
		t.Fatal("Provide() returned nil")
	}
}
