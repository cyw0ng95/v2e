package proc

import (
	"testing"
)

// BenchmarkGetMessageBySize compares the performance of the new tiered message pool
func BenchmarkGetMessageBySize(b *testing.B) {
	b.Run("SmallMessages", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			msg := GetMessageBySize(32)
			PutMessageBySize(msg, 32)
		}
	})

	b.Run("LargeMessages", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			msg := GetMessageBySize(2048)
			PutMessageBySize(msg, 2048)
		}
	})
}

// BenchmarkFastMarshal compares the performance of the new fast marshal
func BenchmarkFastMarshal(b *testing.B) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeRequest
	msg.ID = "test-id"
	msg.Payload = []byte(`{"test":"data","value":123}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.FastMarshal()
	}
}

// BenchmarkCompareMarshalMethods compares different marshal methods
func BenchmarkCompareMarshalMethods(b *testing.B) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeRequest
	msg.ID = "test-id"
	msg.Payload = []byte(`{"test":"data","value":123}`)

	b.Run("OptimizedMarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = msg.OptimizedMarshal()
		}
	})

	b.Run("FastMarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = msg.FastMarshal()
		}
	})
}
