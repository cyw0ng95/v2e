package meta

import "time"

// MemoryCard represents a spaced repetition card linked to a bookmark.
type MemoryCard struct {
	ID         string // Unique identifier
	BookmarkID string // Associated bookmark
	Content    string // Card content
	CreatedAt  int64  // Unix timestamp
}

// CreateMemoryCard creates a memory card for a given bookmark.
func CreateMemoryCard(bookmarkID, content string) *MemoryCard {
	return &MemoryCard{
		ID:         bookmarkID + "-card", // Simple ID scheme
		BookmarkID: bookmarkID,
		Content:    content,
		CreatedAt:  time.Now().Unix(),
	}
}
