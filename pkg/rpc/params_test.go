package rpc

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/json"
	"testing"
)

func TestFetchCVEsParams_JSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFetchCVEsParams_JSON", nil, func(t *testing.T, tx *gorm.DB) {
		orig := FetchCVEsParams{StartIndex: 5, ResultsPerPage: 10}
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded map[string]int
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded["start_index"] != 5 || decoded["results_per_page"] != 10 {
			t.Fatalf("unexpected json fields: %+v", decoded)
		}
	})

}

func TestImportParams_JSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestImportParams_JSON", nil, func(t *testing.T, tx *gorm.DB) {
		params := ImportParams{Path: "file.xml", XSD: "schema.xsd", Force: true}
		data, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded["path"] != "file.xml" || decoded["xsd"] != "schema.xsd" || decoded["force"] != true {
			t.Fatalf("unexpected decoded: %+v", decoded)
		}
	})

}

func TestGetByIDParams_JSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetByIDParams_JSON", nil, func(t *testing.T, tx *gorm.DB) {
		params := GetByIDParams{ID: "abc"}
		data, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded map[string]string
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded["id"] != "abc" {
			t.Fatalf("unexpected id: %v", decoded["id"])
		}
	})

}
