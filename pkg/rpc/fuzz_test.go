package rpc

import (
	"encoding/json"
	"testing"
)

// FuzzFetchCVEsParams tests JSON unmarshaling of FetchCVEsParams with random inputs
func FuzzFetchCVEsParams(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"start_index":0,"results_per_page":100}`))
	f.Add([]byte(`{"start_index":-1,"results_per_page":0}`))
	f.Add([]byte(`{"start_index":999999,"results_per_page":1}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"start_index":"invalid"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var params FetchCVEsParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzImportParams tests JSON unmarshaling of ImportParams with random inputs
func FuzzImportParams(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"path":"/tmp/test.xml","xsd":"/tmp/schema.xsd","force":true}`))
	f.Add([]byte(`{"path":"","xsd":"","force":false}`))
	f.Add([]byte(`{"path":"../../../etc/passwd"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"path":null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var params ImportParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzGetByIDParams tests JSON unmarshaling of GetByIDParams with random inputs
func FuzzGetByIDParams(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"id":"CVE-2021-44228"}`))
	f.Add([]byte(`{"id":""}`))
	f.Add([]byte(`{"id":"<script>alert('xss')</script>"}`))
	f.Add([]byte(`{"id":"'; DROP TABLE users--"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var params GetByIDParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzCVEIDParams tests JSON unmarshaling of CVEIDParams with random inputs
func FuzzCVEIDParams(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"cve_id":"CVE-2021-44228"}`))
	f.Add([]byte(`{"cve_id":""}`))
	f.Add([]byte(`{"cve_id":"NOT-A-CVE"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"cve_id":12345}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var params CVEIDParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzListParams tests JSON unmarshaling of ListParams with random inputs
func FuzzListParams(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"offset":0,"limit":10}`))
	f.Add([]byte(`{"offset":-100,"limit":-1}`))
	f.Add([]byte(`{"offset":999999999,"limit":999999999}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"offset":"invalid","limit":"bad"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var params ListParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}
