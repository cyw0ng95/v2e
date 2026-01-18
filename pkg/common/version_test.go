package common

import "testing"

func TestVersionConst(t *testing.T) {
	v := Version()
	if v == "" {
		t.Fatal("Version() should not be empty")
	}
}
