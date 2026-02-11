package notes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestRPCHandlersHandleCreateBookmark(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleCreateBookmark", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)
		ctx := context.Background()

		params := map[string]interface{}{
			"global_item_id": "CVE-2023-1234",
			"item_type":      "cve",
			"item_id":        "CVE-2023-1234",
			"title":          "Test Bookmark",
			"description":    "Test Description",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCCreateBookmark(ctx, msg)
		require.NoError(t, err)
		assert.Equal(t, subprocess.MessageTypeResponse, resp.Type)

		result := struct {
			Bookmark   *BookmarkModel   `json:"bookmark"`
			MemoryCard *MemoryCardModel `json:"memory_card,omitempty"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.NotNil(t, result.Bookmark)
		assert.Equal(t, "CVE-2023-1234", result.Bookmark.GlobalItemID)
		assert.Equal(t, "Test Bookmark", result.Bookmark.Title)
	})
}

func TestRPCHandlersHandleCreateBookmarkInvalidParams(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleCreateBookmarkInvalidParams", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)
		ctx := context.Background()

		tests := []struct {
			name   string
			params map[string]interface{}
		}{
			{"missing global_item_id", map[string]interface{}{"item_type": "cve", "item_id": "CVE-2023-1234", "title": "Test"}},
			{"missing item_type", map[string]interface{}{"global_item_id": "CVE-2023-1234", "item_id": "CVE-2023-1234", "title": "Test"}},
			{"missing item_id", map[string]interface{}{"global_item_id": "CVE-2023-1234", "item_type": "cve", "title": "Test"}},
			{"missing title", map[string]interface{}{"global_item_id": "CVE-2023-1234", "item_type": "cve", "item_id": "CVE-2023-1234"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := subprocess.MarshalFast(tt.params)
				require.NoError(t, err)

				msg := &subprocess.Message{
					Type:    subprocess.MessageTypeRequest,
					ID:      "test-id",
					Payload: payload,
				}

				resp, err := handlers.handleRPCCreateBookmark(ctx, msg)
				require.NoError(t, err)
				assert.Equal(t, subprocess.MessageTypeResponse, resp.Type)

				var result map[string]interface{}
				require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
				assert.Contains(t, result, "error")
			})
		}
	})
}

func TestRPCHandlersHandleGetBookmarkByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleGetBookmarkByID", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-5678", "cve", "CVE-2023-5678", "Test Bookmark", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"id": float64(bookmark.ID),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCGetBookmarkByID(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Bookmark *BookmarkModel `json:"bookmark"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Equal(t, bookmark.ID, result.Bookmark.ID)
		assert.Equal(t, "CVE-2023-5678", result.Bookmark.GlobalItemID)
	})
}

func TestRPCHandlersHandleGetBookmarkByIDNotFound(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleGetBookmarkByIDNotFound", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)
		ctx := context.Background()

		params := map[string]interface{}{
			"id": float64(99999),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCGetBookmarkByID(ctx, msg)
		require.NoError(t, err)

		var result map[string]interface{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Contains(t, result, "error")
	})
}

func TestRPCHandlersHandleUpdateBookmark(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleUpdateBookmark", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-9999", "cve", "CVE-2023-9999", "Original Title", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"id":          float64(bookmark.ID),
			"title":       "Updated Title",
			"description": "Updated Description",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCUpdateBookmark(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Bookmark *BookmarkModel `json:"bookmark"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Equal(t, "Updated Title", result.Bookmark.Title)
		assert.Equal(t, "Updated Description", result.Bookmark.Description)
	})
}

func TestRPCHandlersHandleDeleteBookmark(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleDeleteBookmark", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-7777", "cve", "CVE-2023-7777", "To Delete", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"id": float64(bookmark.ID),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCDeleteBookmark(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Success bool `json:"success"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.True(t, result.Success)

		_, err = container.BookmarkService.GetBookmarkByID(ctx, bookmark.ID)
		assert.Error(t, err)
	})
}

func TestRPCHandlersHandleAddNote(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleAddNote", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-1111", "cve", "CVE-2023-1111", "Test Bookmark", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"bookmark_id": float64(bookmark.ID),
			"content":     "Test note content",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCAddNote(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Note *NoteModel `json:"note"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.NotNil(t, result.Note)
		assert.Equal(t, "Test note content", result.Note.Content)
	})
}

func TestRPCHandlersHandleGetNotesByBookmarkID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleGetNotesByBookmarkID", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-2222", "cve", "CVE-2023-2222", "Test Bookmark", "")
		require.NoError(t, err)

		note1, err := container.NoteService.AddNote(ctx, bookmark.ID, "First note", nil, false)
		require.NoError(t, err)

		note2, err := container.NoteService.AddNote(ctx, bookmark.ID, "Second note", nil, false)
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"bookmark_id": float64(bookmark.ID),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCGetNotesByBookmarkID(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Notes []*NoteModel `json:"notes"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Len(t, result.Notes, 2)
		assert.Equal(t, note1.ID, result.Notes[0].ID)
		assert.Equal(t, note2.ID, result.Notes[1].ID)
	})
}

func TestRPCHandlersHandleUpdateNote(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleUpdateNote", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-3333", "cve", "CVE-2023-3333", "Test Bookmark", "")
		require.NoError(t, err)

		note, err := container.NoteService.AddNote(ctx, bookmark.ID, "Original content", nil, false)
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"id":      float64(note.ID),
			"content": "Updated content",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCUpdateNote(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Note *NoteModel `json:"note"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Equal(t, "Updated content", result.Note.Content)
	})
}

func TestRPCHandlersHandleDeleteNote(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleDeleteNote", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-4444", "cve", "CVE-2023-4444", "Test Bookmark", "")
		require.NoError(t, err)

		note, err := container.NoteService.AddNote(ctx, bookmark.ID, "To delete", nil, false)
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"id": float64(note.ID),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCDeleteNote(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Success bool `json:"success"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.True(t, result.Success)
	})
}

func TestRPCHandlersHandleCreateMemoryCard(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleCreateMemoryCard", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-5555", "cve", "CVE-2023-5555", "Test Bookmark", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"bookmark_id": float64(bookmark.ID),
			"front":       "Front content",
			"back":        "Back content",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCCreateMemoryCard(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Card *MemoryCardModel `json:"card"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.NotNil(t, result.Card)
		assert.Equal(t, "Front content", result.Card.Front)
		assert.Equal(t, "Back content", result.Card.Back)
	})
}

func TestRPCHandlersHandleUpdateCardAfterReview(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleUpdateCardAfterReview", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-6666", "cve", "CVE-2023-6666", "Test Bookmark", "")
		require.NoError(t, err)

		card, err := container.MemoryCardService.CreateMemoryCard(ctx, bookmark.ID, "Front", "Back")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"card_id": float64(card.ID),
			"rating":  "hard",
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCUpdateCardAfterReview(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Card *MemoryCardModel `json:"card"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.Equal(t, string(StatusReviewed), result.Card.Status)
	})
}

func TestRPCHandlersHandleGetCardsForReview(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleGetCardsForReview", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-7777", "cve", "CVE-2023-7777", "Test Bookmark", "")
		require.NoError(t, err)

		card1, err := container.MemoryCardService.CreateMemoryCard(ctx, bookmark.ID, "Front1", "Back1")
		require.NoError(t, err)
		// Set card to "due" status so it appears in review queue
		card1.Status = string(StatusDue)
		require.NoError(t, db.Save(card1).Error)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"limit": float64(10),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCGetCardsForReview(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Cards []*MemoryCardModel `json:"cards"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.GreaterOrEqual(t, len(result.Cards), 1)
	})
}

func TestRPCHandlersHandleCreateCrossReference(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleCreateCrossReference", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"source_item_id":    "CVE-2023-8888",
			"target_item_id":    "CVE-2023-9999",
			"source_type":       "cve",
			"target_type":       "cve",
			"relationship_type": "related-to",
			"strength":          float64(1.0),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCCreateCrossReference(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Ref *CrossReferenceModel `json:"cross_reference"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.NotNil(t, result.Ref)
		assert.Equal(t, "CVE-2023-8888", result.Ref.SourceItemID)
		assert.Equal(t, "CVE-2023-9999", result.Ref.TargetItemID)
		assert.Equal(t, "related-to", result.Ref.RelationshipType)
	})
}

func TestRPCHandlersHandleGetBookmarkStats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandlersHandleGetBookmarkStats", nil, func(t *testing.T, tx *gorm.DB) {
		// Create in-memory SQLite DB for this test
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Run migrations
		require.NoError(t, MigrateNotesTables(db))

		// Create test infrastructure
		logger := common.NewLogger(nil, "", common.InfoLevel)
		container := NewServiceContainer(db)
		ctx := context.Background()

		bookmark, _, err := container.BookmarkService.CreateBookmark(ctx, "CVE-2023-0000", "cve", "CVE-2023-0000", "Test Bookmark", "")
		require.NoError(t, err)

		sp := subprocess.New("test-notes-rpc")
		handlers := NewRPCHandlers(container, sp, logger)

		params := map[string]interface{}{
			"bookmark_id": float64(bookmark.ID),
		}
		payload, err := subprocess.MarshalFast(params)
		require.NoError(t, err)

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "test-id",
			Payload: payload,
		}

		resp, err := handlers.handleRPCGetBookmarkStats(ctx, msg)
		require.NoError(t, err)

		result := struct {
			Stats map[string]interface{} `json:"stats"`
		}{}
		require.NoError(t, subprocess.UnmarshalPayload(resp, &result))
		assert.NotNil(t, result.Stats)
		// Stats should contain bookmark_id
		assert.Equal(t, float64(bookmark.ID), result.Stats["bookmark_id"])
	})
}
