package cce

// CCE represents a Common Configuration Enumeration
type CCE struct {
	ID          string `json:"id"`          // CCE identifier (e.g., "CCE-00000-0")
	Title       string `json:"title"`       // Human-readable title
	Description string `json:"description"` // Detailed description
	Owner       string `json:"owner"`       // Responsible organization (e.g., "DISA", "NIST")
	Status      string `json:"status"`      // Status of the entry (e.g., "ACTIVE", "DEPRECATED")
	Type        string `json:"type"`        // Type of the CCE
	Reference   string `json:"reference"`   // Reference to related documentation
	Metadata    string `json:"metadata"`    // Additional metadata as JSON string
}

// CCEBatch represents a batch of CCE entries for import
type CCEBatch struct {
	Entries []CCE `json:"entries"`
	Total   int   `json:"total"`
}
