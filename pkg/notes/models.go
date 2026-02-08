package notes

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// BookmarkModel stores bookmarked items with type, global ID, title, learning state, and metadata
type BookmarkModel struct {
	ID           uint   `gorm:"primaryKey"`
	GlobalItemID string `gorm:"index;not null"` // Links to GlobalItemModel
	ItemType     string `gorm:"index;not null"` // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	ItemID       string `gorm:"not null"`       // Original ID from source (e.g., CVE-2021-1234)
	URN          string `gorm:"index"`          // URN reference (e.g., v2e::nvd::cve::CVE-2021-1234)
	Title        string `gorm:"not null"`
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Learning state fields
	LearningState string `gorm:"default:'to-review'"` // to-review, learning, mastered, archived
	LastReviewed  *time.Time
	NextReview    *time.Time
	MasteryLevel  float32 `gorm:"default:0.0"` // 0.0 to 1.0

	// Metadata for storing stats and additional information
	Metadata map[string]interface{} `gorm:"serializer:json"`

	// Relationships
	Notes   []NoteModel            `gorm:"foreignKey:BookmarkID"`
	History []BookmarkHistoryModel `gorm:"foreignKey:BookmarkID"`
}

// BookmarkHistoryModel tracks state changes over time
type BookmarkHistoryModel struct {
	ID         uint   `gorm:"primaryKey"`
	BookmarkID uint   `gorm:"index;not null"`
	Action     string `gorm:"not null"` // e.g., "created", "updated", "learning_state_changed", "note_added", "deleted"
	OldValue   string // Previous state/value
	NewValue   string // New state/value
	Timestamp  time.Time
	UserID     *string // Optional user identifier if multi-user support is added later
}

// NoteModel stores user notes associated with bookmarks
type NoteModel struct {
	ID         uint   `gorm:"primaryKey"`
	BookmarkID uint   `gorm:"index;not null"`
	URN        string `gorm:"uniqueIndex;index"` // Unique URN: v2e::note::<id>
	Content    string `gorm:"type:text;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Author     *string // Optional author field
	IsPrivate  bool    `gorm:"default:false"` // Whether the note is private to the user

	// MemoryFSM state fields (embedded for GORM)
	FSMState        string `gorm:"column:fsm_state;default:'draft'"`
	FSMStateHistory string `gorm:"column:fsm_state_history;type:json"`
	FSMCreatedAt    int64  `gorm:"column:fsm_created_at;autoCreateTime:millisecond"`
	FSMUpdatedAt    int64  `gorm:"column:fsm_updated_at;autoUpdateTime:millisecond"`
}

// MemoryCardModel stores card-specific learning data
type MemoryCardModel struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	BookmarkID uint       `gorm:"index;not null" json:"bookmark_id"`
	URN        string     `gorm:"uniqueIndex;index" json:"urn"`                   // Unique URN: v2e::card::<id>
	Front      string     `gorm:"type:text;not null" json:"front_content"`        // Question/content on front of card
	Back       string     `gorm:"type:text;not null" json:"back_content"`         // Answer/explanation on back of card
	MajorClass string     `gorm:"type:varchar(64);default:''" json:"major_class"` // Major class/category
	MinorClass string     `gorm:"type:varchar(64);default:''" json:"minor_class"` // Minor class/category
	Status     string     `gorm:"type:varchar(32);default:''" json:"status"`      // Status (e.g., active, archived)
	Version    int        `gorm:"default:1" json:"version"`
	Content    string     `gorm:"type:json;not null" json:"content"` // TipTap JSON content
	EaseFactor float32    `gorm:"default:2.5" json:"ease_factor"`    // For spaced repetition algorithm
	Interval   int        `gorm:"default:1" json:"interval"`         // Days until next review
	Repetition int        `gorm:"default:0" json:"repetition"`       // Number of successful reviews
	NextReview *time.Time `json:"next_review_at"`                    // When to review this card next
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Additional fields for card metadata
	CardType  string                 `gorm:"type:varchar(32);default:'basic'" json:"card_type"` // Card type: basic, cloze, reverse
	Author    string                 `gorm:"type:varchar(128);default:''" json:"author"`        // Card creator/author
	IsPrivate bool                   `gorm:"default:false" json:"is_private"`                   // Whether card is private
	Metadata  map[string]interface{} `gorm:"serializer:json" json:"metadata,omitempty"`         // Additional metadata

	// MemoryFSM state fields (embedded for GORM)
	FSMState        string `gorm:"column:fsm_state;default:'new'"`
	FSMStateHistory string `gorm:"column:fsm_state_history;type:json"`
	FSMCreatedAt    int64  `gorm:"column:fsm_created_at;autoCreateTime:millisecond"`
	FSMUpdatedAt    int64  `gorm:"column:fsm_updated_at;autoUpdateTime:millisecond"`
}

// LearningSessionModel tracks learning sessions and progress
type LearningSessionModel struct {
	ID            uint    `gorm:"primaryKey"`
	UserID        *string // Optional user identifier
	SessionStart  time.Time
	SessionEnd    *time.Time
	CardsReviewed int     // Number of cards reviewed in this session
	CardsCorrect  int     // Number of cards answered correctly
	SessionNotes  *string // Optional notes about the session
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CrossReferenceModel stores relationships between items from different databases
type CrossReferenceModel struct {
	ID               uint    `gorm:"primaryKey"`
	SourceItemID     string  `gorm:"index;not null"` // Global ID of source item
	TargetItemID     string  `gorm:"index;not null"` // Global ID of target item
	SourceType       string  `gorm:"not null"`       // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	TargetType       string  `gorm:"not null"`       // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	RelationshipType string  `gorm:"not null"`       // e.g., "related-to", "exploits", "mitigates", "similar-to"
	Strength         float32 `gorm:"default:1.0"`    // Confidence/strength of relationship (0.0 to 1.0)
	Description      *string // Optional description of the relationship
	CreatedAt        time.Time
}

// GlobalItemModel stores unified identifiers for items across sources
type GlobalItemModel struct {
	ID          string `gorm:"primaryKey"`     // Globally unique identifier
	ItemType    string `gorm:"index;not null"` // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	SourceID    string `gorm:"not null"`       // Original ID from source (e.g., CVE-2021-1234)
	URN         string `gorm:"index"`          // URN reference (e.g., v2e::nvd::cve::CVE-2021-1234)
	Title       string `gorm:"not null"`
	Source      string `gorm:"not null"` // e.g., "NVD", "MITRE", "CAPEC", "MITRE_ATT&CK"
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description *string
	// Relationships
	Bookmarks []BookmarkModel `gorm:"foreignKey:GlobalItemID"`
}

// GetURN returns the URN string for this model, computing it if not already set
func (m *BookmarkModel) GetURN() string {
	if m.URN != "" {
		return m.URN
	}
	// Generate URN from ItemType and ItemID
	return GenerateURN(m.ItemType, m.ItemID, "")
}

// GetURN returns the URN string for this model, computing it if not already set
func (m *GlobalItemModel) GetURN() string {
	if m.URN != "" {
		return m.URN
	}
	// Generate URN from ItemType and SourceID
	return GenerateURN(m.ItemType, m.SourceID, m.Source)
}

// GenerateURN creates a URN string from item type, ID, and optional source
// Format: v2e::<provider>::<type>::<atomic_id>
// Examples:
//   - CVE: v2e::nvd::cve::CVE-2024-12233
//   - CWE: v2e::mitre::cwe::CWE-79
//   - CAPEC: v2e::mitre::capec::CAPEC-66
//   - ATT&CK: v2e::mitre::attack::T1566
func GenerateURN(itemType, itemID, source string) string {
	provider := "mitre" // default provider
	if itemType == "CVE" {
		provider = "nvd"
	}
	// Override provider if source is provided
	if source != "" {
		if source == "NVD" {
			provider = "nvd"
		} else if source == "SSG" || source == "ssg" {
			provider = "ssg"
		}
	}

	resourceType := "cve"
	switch itemType {
	case "CVE":
		resourceType = "cve"
	case "CWE":
		resourceType = "cwe"
	case "CAPEC":
		resourceType = "capec"
	case "ATT&CK":
		resourceType = "attack"
	}

	return fmt.Sprintf("v2e::%s::%s::%s", provider, resourceType, itemID)
}

// GetNoteURN generates or returns the URN for a note
func GetNoteURN(id uint) string {
	return fmt.Sprintf("v2e::note::%d", id)
}

// GetCardURN generates or returns the URN for a memory card
func GetCardURN(id uint) string {
	return fmt.Sprintf("v2e::card::%d", id)
}

// GetURN returns the URN for the note
func (m *NoteModel) GetURN() string {
	if m.URN != "" {
		return m.URN
	}
	if m.ID > 0 {
		return GetNoteURN(m.ID)
	}
	return ""
}

// GetMemoryFSMState returns the current FSM state for the note
func (m *NoteModel) GetMemoryFSMState() string {
	return m.FSMState
}

// SetMemoryFSMState sets the FSM state for the note
func (m *NoteModel) SetMemoryFSMState(state string) error {
	m.FSMState = state
	return nil
}

// GetURN returns the URN for the memory card
func (m *MemoryCardModel) GetURN() string {
	if m.URN != "" {
		return m.URN
	}
	if m.ID > 0 {
		return GetCardURN(m.ID)
	}
	return ""
}

// GetMemoryFSMState returns the current FSM state for the card
func (m *MemoryCardModel) GetMemoryFSMState() string {
	return m.FSMState
}

// SetMemoryFSMState sets the FSM state for the card
func (m *MemoryCardModel) SetMemoryFSMState(state string) error {
	m.FSMState = state
	return nil
}
