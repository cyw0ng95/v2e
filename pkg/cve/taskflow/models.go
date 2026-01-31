package taskflow

import "time"

// DataType represents the type of data being populated
type DataType string

const (
	DataTypeCVE    DataType = "cve"
	DataTypeCWE    DataType = "cwe"
	DataTypeCAPEC  DataType = "capec"
	DataTypeATTACK DataType = "attack"
)

// DataProgress tracks progress for each data type
type DataProgress struct {
	TotalCount     int64     `json:"total_count"`
	ProcessedCount int64     `json:"processed_count"`
	ErrorCount     int64     `json:"error_count"`
	StartTime      time.Time `json:"start_time"`
	LastUpdate     time.Time `json:"last_update"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// JobRun represents a single job execution instance with full state
type JobRun struct {
	ID              string    `json:"id"`
	State           JobState  `json:"state"`
	DataType        DataType  `json:"data_type"`
	StartIndex      int       `json:"start_index"`
	ResultsPerBatch int       `json:"results_per_batch"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// Overall progress
	FetchedCount int64  `json:"fetched_count"`
	StoredCount  int64  `json:"stored_count"`
	ErrorCount   int64  `json:"error_count"`
	ErrorMessage string `json:"error_message,omitempty"`
	// Type-specific progress
	Progress map[DataType]DataProgress `json:"progress"`
	// Configuration parameters
	Params map[string]interface{} `json:"params,omitempty"`
}
