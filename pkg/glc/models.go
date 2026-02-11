package glc

import (
	"time"

	"gorm.io/gorm"
)

// GraphModel stores GLC graph data with nodes, edges, and metadata
type GraphModel struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	GraphID     string `gorm:"uniqueIndex;not null" json:"graph_id"` // UUID for graph identification
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	PresetID    string `gorm:"index;not null" json:"preset_id"` // References preset (built-in or user-defined)
	Tags        string `json:"tags"`                            // JSON array of tags
	Nodes       string `gorm:"type:text;not null" json:"nodes"` // JSON array of CADNode
	Edges       string `gorm:"type:text;not null" json:"edges"` // JSON array of CADEdge
	Viewport    string `json:"viewport"`                        // JSON object for viewport state
	Version     int    `gorm:"default:1" json:"version"`
	IsArchived  bool   `gorm:"default:false" json:"is_archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Versions []GraphVersionModel `gorm:"foreignKey:GraphID;constraint:OnDelete:CASCADE" json:"versions,omitempty"`
}

// GraphVersionModel stores version history for undo/restore functionality
type GraphVersionModel struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	GraphID   uint   `gorm:"index;not null" json:"graph_id"`
	Version   int    `gorm:"not null" json:"version"`
	Nodes     string `gorm:"type:text;not null" json:"nodes"`
	Edges     string `gorm:"type:text;not null" json:"edges"`
	Viewport  string `json:"viewport"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserPresetModel stores user-defined canvas presets
type UserPresetModel struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	PresetID    string `gorm:"uniqueIndex;not null" json:"preset_id"` // UUID for preset identification
	Name        string `gorm:"not null" json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Theme       string `gorm:"type:text;not null" json:"theme"`       // JSON object for CanvasPresetTheme
	Behavior    string `gorm:"type:text;not null" json:"behavior"`    // JSON object for CanvasPresetBehavior
	NodeTypes   string `gorm:"type:text;not null" json:"node_types"`  // JSON array of NodeTypeDefinition
	Relations   string `gorm:"type:text;not null" json:"relationships"` // JSON array of RelationshipDefinition
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ShareLinkModel stores share/embed links for graphs
type ShareLinkModel struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	LinkID    string `gorm:"uniqueIndex;not null" json:"link_id"` // Short unique identifier for URL
	GraphID   string `gorm:"index;not null" json:"graph_id"`      // References GraphModel.GraphID
	Password  string `json:"password,omitempty"`                  // Optional password protection (hashed)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	ViewCount int    `gorm:"default:0" json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides for GORM
func (GraphModel) TableName() string {
	return "glc_graphs"
}

func (GraphVersionModel) TableName() string {
	return "glc_graph_versions"
}

func (UserPresetModel) TableName() string {
	return "glc_user_presets"
}

func (ShareLinkModel) TableName() string {
	return "glc_share_links"
}
