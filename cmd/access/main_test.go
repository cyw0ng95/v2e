package main

import (
	"testing"
)

func TestFindModuleRoot(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("Expected to find module root, got error: %v", err)
	}

	if root == "" {
		t.Error("Expected non-empty module root path")
	}

	t.Logf("Module root: %s", root)
}
