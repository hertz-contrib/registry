package test_registry

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) == 3 {
		t.Log("add succeed")
	}
}
