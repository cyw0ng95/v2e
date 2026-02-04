package urn

import (
	"testing"
)

func BenchmarkParse(b *testing.B) {
	input := "v2e::nvd::cve::CVE-2024-12233"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(ProviderNVD, TypeCVE, "CVE-2024-12233")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkURN_String(b *testing.B) {
	urn := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = urn.String()
	}
}

func BenchmarkURN_Key(b *testing.B) {
	urn := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = urn.Key()
	}
}

func BenchmarkURN_Equal(b *testing.B) {
	urn1 := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	urn2 := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = urn1.Equal(urn2)
	}
}

func BenchmarkMustParse(b *testing.B) {
	input := "v2e::nvd::cve::CVE-2024-12233"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustParse(input)
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	input := "v2e::nvd::cve::CVE-2024-12233"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		urn, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
		_ = urn.String()
	}
}
