//go:build !CONFIG_FLOW_ASSERTIONS
// +build !CONFIG_FLOW_ASSERTIONS

package assert

func Assert(checker func() bool, message string) {
	_ = checker
	_ = message
}

func AssertMsg(condition bool, message string) {
	_ = condition
	_ = message
}

func Assertf(checker func() bool, format string, args ...interface{}) {
	_ = checker
	_ = format
	_ = args
}
