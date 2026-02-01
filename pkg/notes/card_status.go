package notes

import (
	"fmt"
	"strings"
)

// CardStatus represents the lifecycle status of a memory card.
type CardStatus string

const (
	StatusNew      CardStatus = "new"
	StatusLearning CardStatus = "learning"
	StatusDue      CardStatus = "due"
	StatusReviewed CardStatus = "reviewed"
	StatusMastered CardStatus = "mastered"
	StatusArchived CardStatus = "archived"
)

// ErrInvalidTransition returned when a requested state transition is not allowed.
var ErrInvalidTransition = fmt.Errorf("invalid card status transition")

// allowedTransitions defines permissible next states for a given current state.
// Keep this intentionally small and explicit so business rules are easy to review.
var allowedTransitions = map[CardStatus]map[CardStatus]bool{
	StatusNew: {
		StatusLearning: true,
		StatusArchived: true,
	},
	StatusLearning: {
		StatusDue:      true,
		StatusArchived: true,
	},
	StatusDue: {
		StatusReviewed: true,
		StatusArchived: true,
	},
	StatusReviewed: {
		StatusLearning: true,
		StatusMastered: true,
		StatusArchived: true,
	},
	StatusMastered: {
		StatusArchived: true,
	},
	StatusArchived: {},
}

// CanTransition returns true if a transition from 'from' to 'to' is allowed.
func CanTransition(from, to CardStatus) bool {
	if from == to {
		return true
	}
	nexts, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return nexts[to]
}

// ParseCardStatus parses an input string into a CardStatus, normalizing case.
func ParseCardStatus(s string) (CardStatus, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "new":
		return StatusNew, nil
	case "learning", "in-progress":
		return StatusLearning, nil
	case "due":
		return StatusDue, nil
	case "reviewed":
		return StatusReviewed, nil
	case "mastered":
		return StatusMastered, nil
	case "archived", "archive":
		return StatusArchived, nil
	default:
		return "", fmt.Errorf("unknown card status: %q", s)
	}
}
