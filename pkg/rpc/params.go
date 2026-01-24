package rpc

import "github.com/cyw0ng95/v2e/pkg/cve"

// FetchCVEsParams are the typed parameters for RPCFetchCVEs
type FetchCVEsParams struct {
	StartIndex     int `json:"start_index"`
	ResultsPerPage int `json:"results_per_page"`
}

// SaveCVEByIDParams are the typed parameters for RPCSaveCVEByID
type SaveCVEByIDParams struct {
	CVE cve.CVEItem `json:"cve"`
}

// GetByIDParams is a general typed param for operations by id
type GetByIDParams struct {
	ID string `json:"id"`
}

// CVEIDParams is used for RPCs that expect a cve_id field
type CVEIDParams struct {
	CVEID string `json:"cve_id"`
}
