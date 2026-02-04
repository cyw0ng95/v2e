package capec

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/xml"
	"testing"
)

func TestCAPECAttackPattern_Unmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCAPECAttackPattern_Unmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		data := `<?xml version="1.0"?>
	    <Attack_Pattern ID="123" Name="Test Pattern" Abstraction="Detailed" Status="Stable">
	      <Summary>Short summary</Summary>
	      <Description>Full description here</Description>
	      <Likelihood_Of_Attack>High</Likelihood_Of_Attack>
	      <Typical_Severity>High</Typical_Severity>
	      <Related_Weaknesses>
	        <Related_Weakness CWE_ID="CWE-79" />
	      </Related_Weaknesses>
	      <Example_Instances>
	        <Example><p>example text</p></Example>
	      </Example_Instances>
	      <Mitigations>
	        <Mitigation>Do X</Mitigation>
	      </Mitigations>
	      <References>
	        <Reference External_Reference_ID="REF-1" />
	      </References>
	    </Attack_Pattern>`

		var ap CAPECAttackPattern
		if err := xml.Unmarshal([]byte(data), &ap); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if ap.ID != 123 {
			t.Fatalf("expected ID 123 got %d", ap.ID)
		}
		if ap.Name != "Test Pattern" {
			t.Fatalf("expected Name, got %q", ap.Name)
		}
		if ap.Abstraction != "Detailed" {
			t.Fatalf("expected Abstraction, got %q", ap.Abstraction)
		}
		if ap.Status != "Stable" {
			t.Fatalf("expected Status, got %q", ap.Status)
		}
		if ap.Summary != "Short summary" {
			t.Fatalf("expected Summary, got %q", ap.Summary)
		}
		if len(ap.RelatedWeaknesses) != 1 || ap.RelatedWeaknesses[0].CWEID != "CWE-79" {
			t.Fatalf("related weaknesses not parsed correctly: %+v", ap.RelatedWeaknesses)
		}
		if len(ap.Examples) != 1 {
			t.Fatalf("examples not parsed: %d", len(ap.Examples))
		}
	})

}
