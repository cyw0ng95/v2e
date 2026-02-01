package notes

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&MemoryCardModel{}, &BookmarkModel{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCanTransition(t *testing.T) {
	if !CanTransition(StatusNew, StatusLearning) {
		t.Fatalf("expected new->learning allowed")
	}
	if CanTransition(StatusArchived, StatusNew) {
		t.Fatalf("expected archived->new disallowed")
	}
}

func TestUpdateMemoryCardFields_StatusTransition(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMemoryCardService(db)

	// create bookmark and card
	bm := &BookmarkModel{GlobalItemID: "g1", ItemType: "test", ItemID: "i1", Title: "t"}
	if err := db.Create(bm).Error; err != nil {
		t.Fatal(err)
	}
	card := &MemoryCardModel{BookmarkID: bm.ID, Front: "q", Back: "a", Status: string(StatusNew), Content: "{}"}
	if err := db.Create(card).Error; err != nil {
		t.Fatal(err)
	}

	// valid transition
	fields := map[string]any{"id": float64(card.ID), "status": "learning"}
	updated, err := svc.UpdateMemoryCardFields(context.Background(), fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != string(StatusLearning) {
		t.Fatalf("expected status learning, got %s", updated.Status)
	}

	// invalid transition
	fields2 := map[string]any{"id": float64(card.ID), "status": "new"}
	_, err = svc.UpdateMemoryCardFields(context.Background(), fields2)
	if err == nil {
		t.Fatalf("expected error for invalid transition")
	}
}

func TestUpdateCardAfterReview_Transition(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMemoryCardService(db)

	bm := &BookmarkModel{GlobalItemID: "g2", ItemType: "test", ItemID: "i2", Title: "t2"}
	if err := db.Create(bm).Error; err != nil {
		t.Fatal(err)
	}
	card := &MemoryCardModel{BookmarkID: bm.ID, Front: "q2", Back: "a2", Status: string(StatusLearning), Content: "{}", Repetition: 5}
	if err := db.Create(card).Error; err != nil {
		t.Fatal(err)
	}

	// call review
	if err := svc.UpdateCardAfterReview(context.Background(), card.ID, CardRatingGood); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var updated MemoryCardModel
	if err := db.First(&updated, card.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Status != string(StatusMastered) {
		t.Fatalf("expected mastered, got %s", updated.Status)
	}
	if updated.NextReview == nil || updated.NextReview.Before(time.Now()) {
		t.Fatalf("expected next review set in future")
	}
}
