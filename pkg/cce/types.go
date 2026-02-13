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

// ToModel converts CCE to CCEModel for database storage
func (c *CCE) ToModel() CCEModel {
	return CCEModel{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Owner:       c.Owner,
		Status:      c.Status,
		Type:        c.Type,
		Reference:   c.Reference,
		Metadata:    c.Metadata,
	}
}

// ToModelSlice converts a slice of CCE to a slice of CCEModel
func ToModelSlice(entries []CCE) []CCEModel {
	models := make([]CCEModel, len(entries))
	for i, entry := range entries {
		models[i] = entry.ToModel()
	}
	return models
}

// ToModelPointer converts CCE pointer to CCEModel pointer
func (c *CCE) ToModelPointer() *CCEModel {
	if c == nil {
		return nil
	}
	m := c.ToModel()
	return &m
}

// ToCCE converts CCEModel to CCE
func (m *CCEModel) ToCCE() CCE {
	if m == nil {
		return CCE{}
	}
	return CCE{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Owner:       m.Owner,
		Status:      m.Status,
		Type:        m.Type,
		Reference:   m.Reference,
		Metadata:    m.Metadata,
	}
}

// ToCCESlice converts a slice of CCEModel to a slice of CCE
func ToCCESlice(models []CCEModel) []CCE {
	entries := make([]CCE, len(models))
	for i, model := range models {
		entries[i] = model.ToCCE()
	}
	return entries
}

// CCEBatch represents a batch of CCE entries for import
type CCEBatch struct {
	Entries []CCE `json:"entries"`
	Total   int   `json:"total"`
}
