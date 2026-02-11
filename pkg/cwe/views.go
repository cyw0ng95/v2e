package cwe

import "encoding/json"

// CWEView represents a view resource from the CWE OpenAPI (V views).
type CWEView struct {
	ID             string           `json:"ID"`
	Name           string           `json:"Name,omitempty"`
	Type           string           `json:"Type,omitempty"`
	Status         string           `json:"Status,omitempty"`
	Objective      string           `json:"Objective,omitempty"`
	Audience       []Stakeholder    `json:"Audience,omitempty"`
	Members        []ViewMember     `json:"Members,omitempty"`
	References     []Reference      `json:"References,omitempty"`
	Notes          []Note           `json:"Notes,omitempty"`
	ContentHistory []ContentHistory `json:"Content_History,omitempty"`
	Raw            json.RawMessage  `json:"Raw,omitempty"`
}

type Stakeholder struct {
	Type        string `json:"Type"`
	Description string `json:"Description,omitempty"`
}

type ViewMember struct {
	CWEID string `json:"CweID"`
	Role  string `json:"Role,omitempty"`
}
