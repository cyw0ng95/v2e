package notes

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestCardStatus(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CanTransition_NewToLearning", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusNew, StatusLearning) {
			t.Errorf("Expected transition from New to Learning to be allowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_NewToArchived", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusNew, StatusArchived) {
			t.Errorf("Expected transition from New to Archived to be allowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_NewToMasteredNotAllowed", nil, func(t *testing.T, tx *gorm.DB) {
		if CanTransition(StatusNew, StatusMastered) {
			t.Errorf("Expected transition from New to Mastered to be disallowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_LearningToReviewed", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusLearning, StatusReviewed) {
			t.Errorf("Expected transition from Learning to Reviewed to be allowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_LearningToMastered", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusLearning, StatusMastered) {
			t.Errorf("Expected transition from Learning to Mastered to be allowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_MasteredToArchived", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusMastered, StatusArchived) {
			t.Errorf("Expected transition from Mastered to Archived to be allowed")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_ArchivedNoTransitions", nil, func(t *testing.T, tx *gorm.DB) {
		if CanTransition(StatusArchived, StatusLearning) {
			t.Errorf("Expected no transitions from Archived state")
		}
		if CanTransition(StatusArchived, StatusReviewed) {
			t.Errorf("Expected no transitions from Archived state")
		}
	})

	testutils.Run(t, testutils.Level1, "CanTransition_SameStateAllowed", nil, func(t *testing.T, tx *gorm.DB) {
		if !CanTransition(StatusNew, StatusNew) {
			t.Errorf("Expected same state transition to be allowed")
		}
		if !CanTransition(StatusLearning, StatusLearning) {
			t.Errorf("Expected same state transition to be allowed")
		}
	})
}

func TestParseCardStatus(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ParseCardStatus_New", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("new")
		if err != nil {
			t.Errorf("Expected no error parsing 'new', got %v", err)
		}
		if status != StatusNew {
			t.Errorf("Expected status to be StatusNew, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Learning", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("learning")
		if err != nil {
			t.Errorf("Expected no error parsing 'learning', got %v", err)
		}
		if status != StatusLearning {
			t.Errorf("Expected status to be StatusLearning, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_InProgress", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("in-progress")
		if err != nil {
			t.Errorf("Expected no error parsing 'in-progress', got %v", err)
		}
		if status != StatusLearning {
			t.Errorf("Expected status to be StatusLearning, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Due", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("due")
		if err != nil {
			t.Errorf("Expected no error parsing 'due', got %v", err)
		}
		if status != StatusDue {
			t.Errorf("Expected status to be StatusDue, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Reviewed", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("reviewed")
		if err != nil {
			t.Errorf("Expected no error parsing 'reviewed', got %v", err)
		}
		if status != StatusReviewed {
			t.Errorf("Expected status to be StatusReviewed, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Mastered", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("mastered")
		if err != nil {
			t.Errorf("Expected no error parsing 'mastered', got %v", err)
		}
		if status != StatusMastered {
			t.Errorf("Expected status to be StatusMastered, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Archived", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("archived")
		if err != nil {
			t.Errorf("Expected no error parsing 'archived', got %v", err)
		}
		if status != StatusArchived {
			t.Errorf("Expected status to be StatusArchived, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Archive", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("archive")
		if err != nil {
			t.Errorf("Expected no error parsing 'archive', got %v", err)
		}
		if status != StatusArchived {
			t.Errorf("Expected status to be StatusArchived, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_CaseInsensitive", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("LEARNING")
		if err != nil {
			t.Errorf("Expected no error parsing 'LEARNING', got %v", err)
		}
		if status != StatusLearning {
			t.Errorf("Expected status to be StatusLearning, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_WhitespaceTrimmed", nil, func(t *testing.T, tx *gorm.DB) {
		status, err := ParseCardStatus("  learning  ")
		if err != nil {
			t.Errorf("Expected no error parsing '  learning  ', got %v", err)
		}
		if status != StatusLearning {
			t.Errorf("Expected status to be StatusLearning, got %v", status)
		}
	})

	testutils.Run(t, testutils.Level1, "ParseCardStatus_Invalid", nil, func(t *testing.T, tx *gorm.DB) {
		_, err := ParseCardStatus("invalid")
		if err == nil {
			t.Errorf("Expected error parsing 'invalid', got nil")
		}
	})
}
