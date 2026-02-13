package assert

import (
	"testing"
)

// TestAssertMsg_Noop tests that AssertMsg is a no-op without the build tag
func TestAssertMsg_Noop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AssertMsg should not panic without CONFIG_FLOW_ASSERTIONS: %v", r)
		}
	}()

	AssertMsg(false, "This should not panic")
	AssertMsg(true, "This should not panic either")
}

// TestAssert_Noop tests that Assert is a no-op without the build tag
func TestAssert_Noop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Assert should not panic without CONFIG_FLOW_ASSERTIONS: %v", r)
		}
	}()

	Assert(func() bool { return false }, "This should not panic")
	Assert(func() bool { return true }, "This should not panic either")
}

// TestAssertf_Noop tests that Assertf is a no-op without the build tag
func TestAssertf_Noop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Assertf should not panic without CONFIG_FLOW_ASSERTIONS: %v", r)
		}
	}()

	Assertf(func() bool { return false }, "test %d", 123)
	Assertf(func() bool { return true }, "test %s", "hello")
}
