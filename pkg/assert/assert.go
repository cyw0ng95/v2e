//go:build CONFIG_FLOW_ASSERTIONS
// +build CONFIG_FLOW_ASSERTIONS

package assert

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
}

// Assert checks a condition and panics if the condition is false
// This is used for flow assertions in the FSM framework
func Assert(checker func() bool, message string) {
	if !checker() {
		stack := debug.Stack()
		log.Printf("FLOW ASSERTION FAILED: %s\n%s", message, stack)
		panic(fmt.Sprintf("FLOW ASSERTION FAILED: %s", message))
	}
}

// AssertMsg is a simpler version that takes a boolean directly
func AssertMsg(condition bool, message string) {
	if !condition {
		stack := debug.Stack()
		log.Printf("FLOW ASSERTION FAILED: %s\n%s", message, stack)
		panic(fmt.Sprintf("FLOW ASSERTION FAILED: %s", message))
	}
}

// Assertf checks a condition and panics with formatted message
func Assertf(checker func() bool, format string, args ...interface{}) {
	if !checker() {
		msg := fmt.Sprintf(format, args...)
		stack := debug.Stack()
		log.Printf("FLOW ASSERTION FAILED: %s\n%s", msg, stack)
		panic(fmt.Sprintf("FLOW ASSERTION FAILED: %s", msg))
	}
}
