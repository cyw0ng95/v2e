package main

const (
	// Log messages
	LogMsgFailedToSetupLogging = "[remote] Failed to setup logging: %v"
	LogMsgServiceStarted       = "[remote] CVE remote service started"

	// Error messages
	ErrMsgFailedDownloadArchive = "failed to download archive: %v"
	ErrMsgUnexpectedHTTPStatus  = "unexpected HTTP status: %s"
	ErrMsgFailedReadBody        = "failed to read archive body: %v"
	ErrMsgFailedOpenZip         = "failed to open zip archive: %v"
	ErrMsgFailedMarshalResp     = "failed to marshal response: %v"
	ErrMsgFailedParseReq        = "failed to parse request: %v"
	ErrMsgCVEIDRequired         = "cve_id is required"
	ErrMsgNVDRateLimited        = "NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)"
	ErrMsgFailedFetchCVE        = "failed to fetch CVE: %v"
	ErrMsgFailedFetchCount      = "failed to fetch CVE count: %v"
	ErrMsgFailedMarshalResult   = "failed to marshal result: %v"
	ErrMsgFailedFetchCVEs       = "failed to fetch CVEs: %v"
)
