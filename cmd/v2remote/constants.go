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

	// Service lifecycle messages
	LogMsgServiceReady            = "[remote] Remote service ready and accepting requests"
	LogMsgServiceShutdownStarting = "[remote] Remote service shutdown starting"
	LogMsgServiceShutdownComplete = "[remote] Remote service shutdown completed"

	// Process configuration messages
	LogMsgProcessIDConfigured    = "[remote] Process ID configured: %s"
	LogMsgBootstrapLoggerCreated = "[remote] Bootstrap logger created"
	LogMsgLoggingSetupComplete   = "[remote] Logging setup completed with level: %v"
	LogMsgSubprocessCreated      = "[remote] Subprocess created with ID: %s"

	// API configuration messages
	LogMsgAPIKeyDetected = "[remote] NVD API key detected in environment"
	LogMsgAPIKeyNotSet   = "[remote] NVD API key not set in environment"
	LogMsgFetcherCreated = "[remote] CVE fetcher created with API key: %t"

	// RPC handler messages
	LogMsgRPCHandlerRegistered = "[remote] RPC handler registered: %s"

	// CVE operation messages
	LogMsgStartingGetCVEByID = "[remote] Starting RPCGetCVEByID request for CVE ID: %s"
	LogMsgGetCVEByIDSuccess  = "[remote] RPCGetCVEByID completed successfully for CVE ID: %s"
	LogMsgGetCVEByIDFailed   = "[remote] RPCGetCVEByID failed for CVE ID: %s, error: %v"
	LogMsgStartingGetCVECnt  = "[remote] Starting RPCGetCVECnt request"
	LogMsgGetCVECntSuccess   = "[remote] RPCGetCVECnt completed successfully, count: %d"
	LogMsgGetCVECntFailed    = "[remote] RPCGetCVECnt failed, error: %v"
	LogMsgStartingFetchCVEs  = "[remote] Starting RPCFetchCVEs request with startIndex: %d, resultsPerPage: %d"
	LogMsgFetchCVEsSuccess   = "[remote] RPCFetchCVEs completed successfully, returned: %d CVEs"
	LogMsgFetchCVEsFailed    = "[remote] RPCFetchCVEs failed, error: %v"
	LogMsgStartingFetchViews = "[remote] Starting RPCFetchViews request with startIndex: %d, resultsPerPage: %d"
	LogMsgFetchViewsSuccess  = "[remote] RPCFetchViews completed successfully, returned: %d views"
	LogMsgFetchViewsFailed   = "[remote] RPCFetchViews failed, error: %v"

	// NVD API interaction messages
	LogMsgDownloadingArchive   = "[remote] Downloading archive from URL: %s"
	LogMsgDownloadSuccess      = "[remote] Archive downloaded successfully"
	LogMsgDownloadFailed       = "[remote] Archive download failed: %v"
	LogMsgHTTPStatusReceived   = "[remote] HTTP status received: %s"
	LogMsgReadingResponseBody  = "[remote] Reading response body"
	LogMsgResponseBodyRead     = "[remote] Response body read successfully, size: %d bytes"
	LogMsgOpeningZipArchive    = "[remote] Opening ZIP archive"
	LogMsgZipArchiveOpened     = "[remote] ZIP archive opened successfully, entries: %d"
	LogMsgZipArchiveOpenFailed = "[remote] ZIP archive open failed: %v"
	LogMsgProcessingZipEntry   = "[remote] Processing ZIP entry: %s"
	LogMsgSkippingZipEntry     = "[remote] Skipping ZIP entry: %s"
	LogMsgExtractingZipEntry   = "[remote] Extracting ZIP entry: %s"
	LogMsgZipEntryExtracted    = "[remote] ZIP entry extracted: %s"
	LogMsgUnmarshalingView     = "[remote] Unmarshaling view data for ID: %s"
	LogMsgViewUnmarshaled      = "[remote] View unmarshaled successfully: %s"
	LogMsgViewUnmarshalFailed  = "[remote] View unmarshal failed for entry: %s, error: %v"
	LogMsgPaginationApplied    = "[remote] Pagination applied: start=%d, pageSize=%d, total=%d, returned=%d"
	LogMsgViewFetchUrlSet      = "[remote] VIEW_FETCH_URL configured: %s"
	LogMsgViewFetchUrlDefault  = "[remote] VIEW_FETCH_URL using default value"
	LogMsgNVDRequestSent       = "[remote] NVD API request sent for CVE ID: %s"
	LogMsgNVDResponseReceived  = "[remote] NVD API response received for CVE ID: %s"
	LogMsgNVDFetchSuccess      = "[remote] NVD fetch succeeded for CVE ID: %s"
	LogMsgNVDFetchFailed       = "[remote] NVD fetch failed for CVE ID: %s, error: %v"
	LogMsgRateLimitEncountered = "[remote] Rate limit encountered, waiting before retry"
	LogMsgRateLimitRetry       = "[remote] Retrying after rate limit delay"
	LogMsgRateLimitExceeded    = "[remote] Rate limit exceeded permanently"

	// Caching messages
	LogMsgCacheHit   = "[remote] Cache hit for request: %s"
	LogMsgCacheMiss  = "[remote] Cache miss for request: %s"
	LogMsgCacheStore = "[remote] Storing response in cache: %s"
)
