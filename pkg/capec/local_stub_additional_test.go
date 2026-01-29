package capec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestImportFromXML_PersistsData(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "capec.db")
	store, err := NewLocalCAPECStore(dbPath)
	if err != nil {
		t.Fatalf("NewLocalCAPECStore: %v", err)
	}

	xmlContent := `<?xml version="1.0"?><Attack_Patterns><Attack_Pattern ID="1" Name="Test"><Description>Desc</Description><Likelihood_Of_Attack>Low</Likelihood_Of_Attack><Typical_Severity>High</Typical_Severity><Related_Weaknesses><Related_Weakness CWE_ID="CWE-1" /></Related_Weaknesses><Example_Instances><Example>Example text</Example></Example_Instances><Mitigations><Mitigation>Mitigation text</Mitigation></Mitigations><References><Reference External_Reference_ID="REF-1" /></References></Attack_Pattern></Attack_Patterns>`
	xmlPath := writeTempFile(t, dir, "capec.xml", xmlContent)

	if err := store.ImportFromXML(xmlPath, true); err != nil {
		t.Fatalf("ImportFromXML: %v", err)
	}

	item, err := store.GetByID(context.Background(), "CAPEC-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if item.Name != "Test" || item.Summary == "" || item.Likelihood != "Low" {
		t.Fatalf("unexpected item: %+v", item)
	}

	weak, _ := store.GetRelatedWeaknesses(context.Background(), 1)
	if len(weak) != 1 || weak[0].CWEID != "CWE-1" {
		t.Fatalf("unexpected weaknesses: %+v", weak)
	}
	examples, _ := store.GetExamples(context.Background(), 1)
	if len(examples) != 1 || examples[0].ExampleText == "" {
		t.Fatalf("expected example text")
	}
	mitigations, _ := store.GetMitigations(context.Background(), 1)
	if len(mitigations) != 1 || mitigations[0].MitigationText == "" {
		t.Fatalf("expected mitigation text")
	}
	refs, _ := store.GetReferences(context.Background(), 1)
	if len(refs) != 1 || refs[0].ExternalReference != "REF-1" {
		t.Fatalf("unexpected refs: %+v", refs)
	}
}

func TestListCAPECsPaginated_ReturnsTotalAndItems(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalCAPECStore(filepath.Join(dir, "capec.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	// Seed two items directly
	store.db.Create(&CAPECItemModel{CAPECID: 1, Name: "One"})
	store.db.Create(&CAPECItemModel{CAPECID: 2, Name: "Two"})

	items, total, err := store.ListCAPECsPaginated(context.Background(), 0, 1)
	if err != nil {
		t.Fatalf("ListCAPECsPaginated: %v", err)
	}
	if total != 2 || len(items) != 1 || items[0].CAPECID != 1 {
		t.Fatalf("unexpected pagination result: total=%d items=%v", total, items)
	}
}
