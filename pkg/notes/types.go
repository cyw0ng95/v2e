package notes

// LearningState represents the learning state of a bookmarked item
type LearningState string

const (
	// LearningStateToReview indicates the item is queued for review
	LearningStateToReview LearningState = "to-review"
	// LearningStateLearning indicates the item is currently being studied
	LearningStateLearning LearningState = "learning"
	// LearningStateMastered indicates the item has been mastered
	LearningStateMastered LearningState = "mastered"
	// LearningStateArchived indicates the item has been archived
	LearningStateArchived LearningState = "archived"
)

// CardRating represents the rating given after reviewing a memory card
type CardRating string

const (
	// CardRatingAgain means the user couldn't recall the information
	CardRatingAgain CardRating = "again"
	// CardRatingHard means the user recalled with difficulty
	CardRatingHard CardRating = "hard"
	// CardRatingGood means the user recalled with some hesitation
	CardRatingGood CardRating = "good"
	// CardRatingEasy means the user recalled easily
	CardRatingEasy CardRating = "easy"
)

// RelationshipType represents the type of relationship between items
type RelationshipType string

const (
	// RelationshipTypeRelatedTo indicates items are generally related
	RelationshipTypeRelatedTo RelationshipType = "related-to"
	// RelationshipTypeExploits indicates the source exploits the target
	RelationshipTypeExploits RelationshipType = "exploits"
	// RelationshipTypeMitigates indicates the source mitigates the target
	RelationshipTypeMitigates RelationshipType = "mitigates"
	// RelationshipTypeSimilarTo indicates items are similar in nature
	RelationshipTypeSimilarTo RelationshipType = "similar-to"
	// RelationshipTypePartOf indicates the source is part of the target
	RelationshipTypePartOf RelationshipType = "part-of"
	// RelationshipTypeCausedBy indicates the source is caused by the target
	RelationshipTypeCausedBy RelationshipType = "caused-by"
)

// BookmarkAction represents the type of action performed on a bookmark
type BookmarkAction string

const (
	// BookmarkActionCreated indicates a bookmark was created
	BookmarkActionCreated BookmarkAction = "created"
	// BookmarkActionUpdated indicates a bookmark was updated
	BookmarkActionUpdated BookmarkAction = "updated"
	// BookmarkActionLearningStateChanged indicates the learning state was changed
	BookmarkActionLearningStateChanged BookmarkAction = "learning_state_changed"
	// BookmarkActionNoteAdded indicates a note was added to the bookmark
	BookmarkActionNoteAdded BookmarkAction = "note_added"
	// BookmarkActionDeleted indicates a bookmark was deleted
	BookmarkActionDeleted BookmarkAction = "deleted"
	// BookmarkActionReviewed indicates a bookmark was reviewed in learning mode
	BookmarkActionReviewed BookmarkAction = "reviewed"
)

// ItemType represents the type of security item
type ItemType string

const (
	// ItemTypeCVE indicates a CVE item
	ItemTypeCVE ItemType = "CVE"
	// ItemTypeCWE indicates a CWE item
	ItemTypeCWE ItemType = "CWE"
	// ItemTypeCAPEC indicates a CAPEC item
	ItemTypeCAPEC ItemType = "CAPEC"
	// ItemTypeAttack indicates an ATT&CK item
	ItemTypeAttack ItemType = "ATT&CK"
)