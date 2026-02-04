package notes

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
	"time"
)

// MockDBConnection creates a mock database connection for testing
// In real tests, this would connect to a test database
func TestBookmarkModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBookmarkModel", nil, func(t *testing.T, tx *gorm.DB) {
		// Test basic struct properties
		b := &BookmarkModel{
			GlobalItemID:  "global-item-123",
			ItemType:      "CVE",
			ItemID:        "CVE-2021-1234",
			Title:         "Test Vulnerability",
			Description:   "A test vulnerability for bookmarking",
			LearningState: string(LearningStateToReview),
			MasteryLevel:  0.5,
		}

		// Verify basic properties
		if b.GlobalItemID != "global-item-123" {
			t.Errorf("Expected GlobalItemID to be 'global-item-123', got '%s'", b.GlobalItemID)
		}

		if b.ItemType != "CVE" {
			t.Errorf("Expected ItemType to be 'CVE', got '%s'", b.ItemType)
		}

		if b.LearningState != string(LearningStateToReview) {
			t.Errorf("Expected LearningState to be '%s', got '%s'", string(LearningStateToReview), b.LearningState)
		}

		if b.MasteryLevel != 0.5 {
			t.Errorf("Expected MasteryLevel to be 0.5, got %f", b.MasteryLevel)
		}
	})

}

func TestBookmarkHistoryModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBookmarkHistoryModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		h := &BookmarkHistoryModel{
			BookmarkID: 1,
			Action:     string(BookmarkActionCreated),
			OldValue:   "",
			NewValue:   string(LearningStateToReview),
			Timestamp:  now,
		}

		if h.BookmarkID != 1 {
			t.Errorf("Expected BookmarkID to be 1, got %d", h.BookmarkID)
		}

		if h.Action != string(BookmarkActionCreated) {
			t.Errorf("Expected Action to be '%s', got '%s'", string(BookmarkActionCreated), h.Action)
		}

		if h.Timestamp.IsZero() {
			t.Errorf("Expected Timestamp to be set, but it's zero")
		}
	})

}

func TestNoteModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNoteModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		n := &NoteModel{
			BookmarkID: 1,
			Content:    "This is a test note",
			CreatedAt:  now,
			IsPrivate:  false,
		}

		if n.BookmarkID != 1 {
			t.Errorf("Expected BookmarkID to be 1, got %d", n.BookmarkID)
		}

		if n.Content != "This is a test note" {
			t.Errorf("Expected Content to be 'This is a test note', got '%s'", n.Content)
		}

		if n.IsPrivate != false {
			t.Errorf("Expected IsPrivate to be false, got %t", n.IsPrivate)
		}
	})

}

func TestMemoryCardModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMemoryCardModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		c := &MemoryCardModel{
			BookmarkID: 1,
			Front:      "What is CVE-2021-1234?",
			Back:       "A test vulnerability in a software component",
			EaseFactor: 2.5,
			Interval:   1,
			Repetition: 0,
			CreatedAt:  now,
		}

		if c.BookmarkID != 1 {
			t.Errorf("Expected BookmarkID to be 1, got %d", c.BookmarkID)
		}

		if c.Front != "What is CVE-2021-1234?" {
			t.Errorf("Expected Front to be 'What is CVE-2021-1234?', got '%s'", c.Front)
		}

		if c.Back != "A test vulnerability in a software component" {
			t.Errorf("Expected Back to be 'A test vulnerability in a software component', got '%s'", c.Back)
		}

		if c.EaseFactor != 2.5 {
			t.Errorf("Expected EaseFactor to be 2.5, got %f", c.EaseFactor)
		}

		if c.Interval != 1 {
			t.Errorf("Expected Interval to be 1, got %d", c.Interval)
		}

		if c.Repetition != 0 {
			t.Errorf("Expected Repetition to be 0, got %d", c.Repetition)
		}
	})

}

func TestLearningSessionModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLearningSessionModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		s := &LearningSessionModel{
			SessionStart:  now,
			CardsReviewed: 5,
			CardsCorrect:  4,
		}

		if s.CardsReviewed != 5 {
			t.Errorf("Expected CardsReviewed to be 5, got %d", s.CardsReviewed)
		}

		if s.CardsCorrect != 4 {
			t.Errorf("Expected CardsCorrect to be 4, got %d", s.CardsCorrect)
		}

		if s.SessionStart.IsZero() {
			t.Errorf("Expected SessionStart to be set, but it's zero")
		}
	})

}

func TestCrossReferenceModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCrossReferenceModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		cr := &CrossReferenceModel{
			SourceItemID:     "global-source-123",
			TargetItemID:     "global-target-456",
			SourceType:       "CVE",
			TargetType:       "CWE",
			RelationshipType: string(RelationshipTypeExploits),
			Strength:         0.8,
			CreatedAt:        now,
		}

		if cr.SourceItemID != "global-source-123" {
			t.Errorf("Expected SourceItemID to be 'global-source-123', got '%s'", cr.SourceItemID)
		}

		if cr.TargetItemID != "global-target-456" {
			t.Errorf("Expected TargetItemID to be 'global-target-456', got '%s'", cr.TargetItemID)
		}

		if cr.SourceType != "CVE" {
			t.Errorf("Expected SourceType to be 'CVE', got '%s'", cr.SourceType)
		}

		if cr.TargetType != "CWE" {
			t.Errorf("Expected TargetType to be 'CWE', got '%s'", cr.TargetType)
		}

		if cr.RelationshipType != string(RelationshipTypeExploits) {
			t.Errorf("Expected RelationshipType to be '%s', got '%s'", string(RelationshipTypeExploits), cr.RelationshipType)
		}

		if cr.Strength != 0.8 {
			t.Errorf("Expected Strength to be 0.8, got %f", cr.Strength)
		}
	})

}

func TestGlobalItemModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGlobalItemModel", nil, func(t *testing.T, tx *gorm.DB) {
		now := time.Now()
		gi := &GlobalItemModel{
			ID:        "global-item-789",
			ItemType:  "CAPEC",
			SourceID:  "CAPEC-123",
			Title:     "Test Attack Pattern",
			Source:    "MITRE_CAPEC",
			CreatedAt: now,
		}

		if gi.ID != "global-item-789" {
			t.Errorf("Expected ID to be 'global-item-789', got '%s'", gi.ID)
		}

		if gi.ItemType != "CAPEC" {
			t.Errorf("Expected ItemType to be 'CAPEC', got '%s'", gi.ItemType)
		}

		if gi.SourceID != "CAPEC-123" {
			t.Errorf("Expected SourceID to be 'CAPEC-123', got '%s'", gi.SourceID)
		}

		if gi.Title != "Test Attack Pattern" {
			t.Errorf("Expected Title to be 'Test Attack Pattern', got '%s'", gi.Title)
		}

		if gi.Source != "MITRE_CAPEC" {
			t.Errorf("Expected Source to be 'MITRE_CAPEC', got '%s'", gi.Source)
		}
	})

}

func TestLearningStateConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLearningStateConstants", nil, func(t *testing.T, tx *gorm.DB) {
		states := []LearningState{
			LearningStateToReview,
			LearningStateLearning,
			LearningStateMastered,
			LearningStateArchived,
		}

		expectedValues := []string{
			"to-review",
			"learning",
			"mastered",
			"archived",
		}

		for i, state := range states {
			if string(state) != expectedValues[i] {
				t.Errorf("Expected state %d to be '%s', got '%s'", i, expectedValues[i], state)
			}
		}
	})

}

func TestCardRatingConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCardRatingConstants", nil, func(t *testing.T, tx *gorm.DB) {
		ratings := []CardRating{
			CardRatingAgain,
			CardRatingHard,
			CardRatingGood,
			CardRatingEasy,
		}

		expectedValues := []string{
			"again",
			"hard",
			"good",
			"easy",
		}

		for i, rating := range ratings {
			if string(rating) != expectedValues[i] {
				t.Errorf("Expected rating %d to be '%s', got '%s'", i, expectedValues[i], rating)
			}
		}
	})

}

func TestRelationshipTypeConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRelationshipTypeConstants", nil, func(t *testing.T, tx *gorm.DB) {
		relationships := []RelationshipType{
			RelationshipTypeRelatedTo,
			RelationshipTypeExploits,
			RelationshipTypeMitigates,
			RelationshipTypeSimilarTo,
			RelationshipTypePartOf,
			RelationshipTypeCausedBy,
		}

		expectedValues := []string{
			"related-to",
			"exploits",
			"mitigates",
			"similar-to",
			"part-of",
			"caused-by",
		}

		for i, rel := range relationships {
			if string(rel) != expectedValues[i] {
				t.Errorf("Expected relationship %d to be '%s', got '%s'", i, expectedValues[i], rel)
			}
		}
	})

}

func TestBookmarkActionConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBookmarkActionConstants", nil, func(t *testing.T, tx *gorm.DB) {
		actions := []BookmarkAction{
			BookmarkActionCreated,
			BookmarkActionUpdated,
			BookmarkActionLearningStateChanged,
			BookmarkActionNoteAdded,
			BookmarkActionDeleted,
			BookmarkActionReviewed,
		}

		expectedValues := []string{
			"created",
			"updated",
			"learning_state_changed",
			"note_added",
			"deleted",
			"reviewed",
		}

		for i, action := range actions {
			if string(action) != expectedValues[i] {
				t.Errorf("Expected action %d to be '%s', got '%s'", i, expectedValues[i], action)
			}
		}
	})

}

func TestItemTypeConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestItemTypeConstants", nil, func(t *testing.T, tx *gorm.DB) {
		types := []ItemType{
			ItemTypeCVE,
			ItemTypeCWE,
			ItemTypeCAPEC,
			ItemTypeAttack,
		}

		expectedValues := []string{
			"CVE",
			"CWE",
			"CAPEC",
			"ATT&CK",
		}

		for i, itemType := range types {
			if string(itemType) != expectedValues[i] {
				t.Errorf("Expected item type %d to be '%s', got '%s'", i, expectedValues[i], itemType)
			}
		}
	})

}
