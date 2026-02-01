package notes

import (
	"context"

	"gorm.io/gorm"
)

// BookmarkServiceInterface defines the interface for the bookmark service
type BookmarkServiceInterface interface {
	CreateBookmark(ctx context.Context, globalItemID, itemType, itemID, title, description string) (*BookmarkModel, *MemoryCardModel, error)
	GetBookmarkByID(ctx context.Context, id uint) (*BookmarkModel, error)
	GetBookmarksByGlobalItemID(ctx context.Context, globalItemID string) ([]*BookmarkModel, error)
	UpdateBookmark(ctx context.Context, bookmark *BookmarkModel) error
	UpdateLearningState(ctx context.Context, bookmarkID uint, newState LearningState) error
	GetBookmarksByLearningState(ctx context.Context, state LearningState) ([]*BookmarkModel, error)
	ListBookmarks(ctx context.Context, state string, offset, limit int) ([]*BookmarkModel, int64, error)
	DeleteBookmark(ctx context.Context, id uint) error
	UpdateBookmarkStats(ctx context.Context, bookmarkID uint, viewIncrement int, studyIncrement int) error
	GetBookmarkStats(ctx context.Context, bookmarkID uint) (map[string]interface{}, error)
}

// NoteServiceInterface defines the interface for the note service
type NoteServiceInterface interface {
	AddNote(ctx context.Context, bookmarkID uint, content string, author *string, isPrivate bool) (*NoteModel, error)
	GetNoteByID(ctx context.Context, id uint) (*NoteModel, error)
	GetNotesByBookmarkID(ctx context.Context, bookmarkID uint) ([]*NoteModel, error)
	UpdateNote(ctx context.Context, note *NoteModel) error
	DeleteNote(ctx context.Context, id uint) error
}

// MemoryCardServiceInterface defines the interface for the memory card service
type MemoryCardServiceInterface interface {
	CreateMemoryCard(ctx context.Context, bookmarkID uint, front, back string) (*MemoryCardModel, error)
	GetMemoryCardByID(ctx context.Context, id uint) (*MemoryCardModel, error)
	GetMemoryCardsByBookmarkID(ctx context.Context, bookmarkID uint) ([]*MemoryCardModel, error)
	GetCardsForReview(ctx context.Context) ([]*MemoryCardModel, error)
	GetCardsByLearningState(ctx context.Context, state LearningState) ([]*MemoryCardModel, error)
	UpdateCardAfterReview(ctx context.Context, cardID uint, rating CardRating) error
	UpdateMemoryCardFields(ctx context.Context, fields map[string]any) (*MemoryCardModel, error)
	DeleteMemoryCard(ctx context.Context, id uint) error
	TransitionCardStatus(ctx context.Context, cardID uint, expectedVersion *int, next CardStatus) error
	ListMemoryCards(ctx context.Context, bookmarkID *uint, learningState *string, author *string, isPrivate *bool, offset, limit int) ([]*MemoryCardModel, int64, error)
}

// CrossReferenceServiceInterface defines the interface for the cross-reference service
type CrossReferenceServiceInterface interface {
	CreateCrossReference(ctx context.Context, sourceItemID, targetItemID, sourceType, targetType, relationshipType string, strength float32, description *string) (*CrossReferenceModel, error)
	GetCrossReferencesBySource(ctx context.Context, sourceItemID string) ([]*CrossReferenceModel, error)
	GetCrossReferencesByTarget(ctx context.Context, targetItemID string) ([]*CrossReferenceModel, error)
	GetCrossReferencesByType(ctx context.Context, relationshipType RelationshipType) ([]*CrossReferenceModel, error)
	GetBidirectionalCrossReferences(ctx context.Context, itemID1, itemID2 string) ([]*CrossReferenceModel, error)
}

// HistoryServiceInterface defines the interface for the history service
type HistoryServiceInterface interface {
	GetHistoryByBookmarkID(ctx context.Context, bookmarkID uint) ([]*BookmarkHistoryModel, error)
	RevertBookmarkState(ctx context.Context, bookmarkID uint, timestamp interface{}) error
}

// ServiceContainer holds all the service implementations
type ServiceContainer struct {
	BookmarkService       BookmarkServiceInterface
	NoteService           NoteServiceInterface
	MemoryCardService     MemoryCardServiceInterface
	CrossReferenceService CrossReferenceServiceInterface
	HistoryService        HistoryServiceInterface
}

// NewServiceContainer creates a new service container with local implementations
func NewServiceContainer(db *gorm.DB) *ServiceContainer {
	bookmarkService := NewBookmarkService(db)
	noteService := NewNoteService(db)
	memoryCardService := NewMemoryCardService(db)
	crossRefService := NewCrossReferenceService(db)
	historyService := NewHistoryService(db)

	return &ServiceContainer{
		BookmarkService:       bookmarkService,
		NoteService:           noteService,
		MemoryCardService:     memoryCardService,
		CrossReferenceService: crossRefService,
		HistoryService:        historyService,
	}
}

// NewRPCServiceContainer creates a new service container with RPC implementations
func NewRPCServiceContainer(client *RPCClient) *ServiceContainer {
	bookmarkService, noteService, memoryCardService, crossRefService, historyService := client.GetRPCClients()

	return &ServiceContainer{
		BookmarkService:       bookmarkService,
		NoteService:           noteService,
		MemoryCardService:     memoryCardService,
		CrossReferenceService: crossRefService,
		HistoryService:        historyService,
	}
}
