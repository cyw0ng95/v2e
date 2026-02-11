package testutils

import (
	"encoding/json"
)

func MakeCVEResponseJSON(cveID string, total int) []byte {
	resp := map[string]interface{}{
		"resultsPerPage":  1,
		"startIndex":      0,
		"totalResults":    total,
		"format":          "",
		"version":         "",
		"vulnerabilities": []interface{}{},
	}
	if cveID != "" {
		resp["vulnerabilities"] = []interface{}{
			map[string]interface{}{
				"cve": map[string]interface{}{
					"id":           cveID,
					"descriptions": []map[string]string{{"lang": "en", "value": "test description"}},
				},
			},
		}
	}
	b, _ := json.Marshal(resp)
	return b
}

func MakeCVEListResponseJSON(total int) []byte {
	return MakeCVEResponseJSON("", total)
}
