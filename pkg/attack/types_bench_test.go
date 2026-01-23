package attack

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

// Helper function to create a test ATT&CK technique
func createTestTechnique(id string) *AttackTechnique {
	return &AttackTechnique{
		ID:          id,
		Name:        "Test Technique",
		Description: "This is a test description for the ATT&CK technique",
		Domain:      "enterprise-attack",
		Platform:    "Windows,Linux,macOS",
		Created:     "2024-01-01T00:00:00.000Z",
		Modified:    "2024-01-01T00:00:00.000Z",
		Revoked:     false,
		Deprecated:  false,
	}
}

// Helper function to create a test ATT&CK tactic
func createTestTactic(id string) *AttackTactic {
	return &AttackTactic{
		ID:          id,
		Name:        "Test Tactic",
		Description: "This is a test description for the ATT&CK tactic",
		Domain:      "enterprise-attack",
		Created:     "2024-01-01T00:00:00.000Z",
		Modified:    "2024-01-01T00:00:00.000Z",
	}
}

// Helper function to create a test ATT&CK mitigation
func createTestMitigation(id string) *AttackMitigation {
	return &AttackMitigation{
		ID:          id,
		Name:        "Test Mitigation",
		Description: "This is a test description for the ATT&CK mitigation",
		Domain:      "enterprise-attack",
		Created:     "2024-01-01T00:00:00.000Z",
		Modified:    "2024-01-01T00:00:00.000Z",
	}
}

// Helper function to create a test ATT&CK software
func createTestSoftware(id string) *AttackSoftware {
	return &AttackSoftware{
		ID:          id,
		Name:        "Test Software",
		Description: "This is a test description for the ATT&CK software",
		Type:        "malware",
		Domain:      "enterprise-attack",
		Created:     "2024-01-01T00:00:00.000Z",
		Modified:    "2024-01-01T00:00:00.000Z",
	}
}

// BenchmarkAttackTechniqueJSONMarshal benchmarks marshaling ATT&CK techniques to JSON
func BenchmarkAttackTechniqueJSONMarshal(b *testing.B) {
	technique := createTestTechnique("T1001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(technique)
		if err != nil {
			b.Fatalf("Failed to marshal technique: %v", err)
		}
	}
}

// BenchmarkAttackTechniqueJSONUnmarshal benchmarks unmarshaling ATT&CK techniques from JSON
func BenchmarkAttackTechniqueJSONUnmarshal(b *testing.B) {
	technique := createTestTechnique("T1001")
	data, _ := json.Marshal(technique)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result AttackTechnique
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal technique: %v", err)
		}
	}
}

// BenchmarkAttackTechniqueSonicMarshal benchmarks marshaling ATT&CK techniques using Sonic
func BenchmarkAttackTechniqueSonicMarshal(b *testing.B) {
	technique := createTestTechnique("T1001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(technique)
		if err != nil {
			b.Fatalf("Failed to marshal technique with sonic: %v", err)
		}
	}
}

// BenchmarkAttackTechniqueSonicUnmarshal benchmarks unmarshaling ATT&CK techniques using Sonic
func BenchmarkAttackTechniqueSonicUnmarshal(b *testing.B) {
	technique := createTestTechnique("T1001")
	data, _ := sonic.Marshal(technique)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result AttackTechnique
		err := sonic.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal technique with sonic: %v", err)
		}
	}
}

// BenchmarkAttackTacticJSONMarshal benchmarks marshaling ATT&CK tactics to JSON
func BenchmarkAttackTacticJSONMarshal(b *testing.B) {
	tactic := createTestTactic("TA0001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(tactic)
		if err != nil {
			b.Fatalf("Failed to marshal tactic: %v", err)
		}
	}
}

// BenchmarkAttackTacticJSONUnmarshal benchmarks unmarshaling ATT&CK tactics from JSON
func BenchmarkAttackTacticJSONUnmarshal(b *testing.B) {
	tactic := createTestTactic("TA0001")
	data, _ := json.Marshal(tactic)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result AttackTactic
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal tactic: %v", err)
		}
	}
}

// BenchmarkAttackMitigationJSONMarshal benchmarks marshaling ATT&CK mitigations to JSON
func BenchmarkAttackMitigationJSONMarshal(b *testing.B) {
	mitigation := createTestMitigation("M1001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(mitigation)
		if err != nil {
			b.Fatalf("Failed to marshal mitigation: %v", err)
		}
	}
}

// BenchmarkAttackMitigationJSONUnmarshal benchmarks unmarshaling ATT&CK mitigations from JSON
func BenchmarkAttackMitigationJSONUnmarshal(b *testing.B) {
	mitigation := createTestMitigation("M1001")
	data, _ := json.Marshal(mitigation)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result AttackMitigation
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal mitigation: %v", err)
		}
	}
}

// BenchmarkAttackSoftwareJSONMarshal benchmarks marshaling ATT&CK software to JSON
func BenchmarkAttackSoftwareJSONMarshal(b *testing.B) {
	software := createTestSoftware("S0001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(software)
		if err != nil {
			b.Fatalf("Failed to marshal software: %v", err)
		}
	}
}

// BenchmarkAttackSoftwareJSONUnmarshal benchmarks unmarshaling ATT&CK software from JSON
func BenchmarkAttackSoftwareJSONUnmarshal(b *testing.B) {
	software := createTestSoftware("S0001")
	data, _ := json.Marshal(software)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result AttackSoftware
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal software: %v", err)
		}
	}
}

// BenchmarkAttackSliceMarshal benchmarks marshaling slices of ATT&CK objects
func BenchmarkAttackSliceMarshal(b *testing.B) {
	techniques := make([]AttackTechnique, 100)
	for i := range techniques {
		techniques[i] = *createTestTechnique("T" + string(rune('1'+i%9)) + string(rune('0'+i/9%10)) + string(rune('0'+i/90%10)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(techniques)
		if err != nil {
			b.Fatalf("Failed to marshal slice: %v", err)
		}
	}
}

// BenchmarkAttackSliceUnmarshal benchmarks unmarshaling slices of ATT&CK objects
func BenchmarkAttackSliceUnmarshal(b *testing.B) {
	techniques := make([]AttackTechnique, 100)
	for i := range techniques {
		techniques[i] = *createTestTechnique("T" + string(rune('1'+i%9)) + string(rune('0'+i/9%10)) + string(rune('0'+i/90%10)))
	}
	data, _ := json.Marshal(techniques)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []AttackTechnique
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal slice: %v", err)
		}
	}
}