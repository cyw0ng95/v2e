package common

import "time"

// Timeout defaults for RPC and operations
const (
	// DefaultRPCTimeout is the standard timeout for RPC requests between services
	// Used by: sysmon, access, meta (should use this instead of 60s)
	DefaultRPCTimeout = 30 * time.Second

	// DefaultLongOperationTimeout is for operations that take longer (imports, bulk operations)
	// Used by: meta for CWE/CAPEC/ATT&CK imports
	DefaultLongOperationTimeout = 120 * time.Second

	// DefaultShutdownTimeout is the graceful shutdown timeout
	DefaultShutdownTimeout = 10 * time.Second
)

// Database path defaults
const (
	DefaultSessionDBPath  = "session.db"
	DefaultCVEDBPath      = "cve.db"
	DefaultCWEDBPath      = "cwe.db"
	DefaultCAPECDBPath    = "capec.db"
	DefaultATTACKDBPath   = "attack.db"
	DefaultBookmarkDBPath = "bookmark.db"
)

// Pagination and buffer size defaults
const (
	DefaultPageSize    = 100
	MaxPageSize        = 1000
	DefaultWorkerCount = 100 // For taskflow job executor
)

// Asset path defaults
const (
	DefaultCWEAssetPath    = "assets/cwe-raw.json"
	DefaultCAPECAssetPath  = "assets/capec_contents_latest.xml"
	DefaultATTACKAssetPath = "assets/enterprise-attack.xlsx"
)
