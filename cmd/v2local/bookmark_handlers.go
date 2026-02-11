package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createBookmarkHandlers creates all bookmark-related RPC handlers
func createBookmarkHandlers(sp *subprocess.Subprocess, serviceContainer *notes.ServiceContainer, logger *common.Logger) {
	bookmarkService := serviceContainer.BookmarkService.(*notes.BookmarkService)
	noteService := serviceContainer.NoteService.(*notes.NoteService)

	sp.RegisterHandler("RPCCreateBookmark", createCreateBookmarkHandler(bookmarkService, logger))
	sp.RegisterHandler("RPCGetBookmark", createGetBookmarkHandler(bookmarkService, logger))
	sp.RegisterHandler("RPCUpdateBookmark", createUpdateBookmarkHandler(bookmarkService, logger))
	sp.RegisterHandler("RPCDeleteBookmark", createDeleteBookmarkHandler(bookmarkService, logger))
	sp.RegisterHandler("RPCListBookmarks", createListBookmarksHandler(bookmarkService, logger))
	sp.RegisterHandler("RPCAddNote", createAddNoteHandler(noteService, logger))
	sp.RegisterHandler("RPCGetNote", createGetNoteHandler(noteService, logger))
	sp.RegisterHandler("RPCUpdateNote", createUpdateNoteHandler(noteService, logger))
	sp.RegisterHandler("RPCDeleteNote", createDeleteNoteHandler(noteService, logger))
	sp.RegisterHandler("RPCGetNotesByBookmark", createGetNotesByBookmarkHandler(noteService, logger))
	logger.Info("Bookmark and Note handlers registered")
}

// createCreateBookmarkHandler creates a handler for RPCCreateBookmark
func createCreateBookmarkHandler(service *notes.BookmarkService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCCreateBookmark handler invoked")

		var params struct {
			GlobalItemID string `json:"global_item_id"`
			ItemType     string `json:"item_type"`
			ItemID       string `json:"item_id"`
			Title        string `json:"title"`
			Description  string `json:"description"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GlobalItemID == "" {
			logger.Warn("global_item_id is required")
			return subprocess.NewErrorResponse(msg, "global_item_id is required"), nil
		}
		if params.ItemType == "" {
			logger.Warn("item_type is required")
			return subprocess.NewErrorResponse(msg, "item_type is required"), nil
		}
		if params.ItemID == "" {
			logger.Warn("item_id is required")
			return subprocess.NewErrorResponse(msg, "item_id is required"), nil
		}
		if params.Title == "" {
			logger.Warn("title is required")
			return subprocess.NewErrorResponse(msg, "title is required"), nil
		}

		bookmark, memoryCard, err := service.CreateBookmark(ctx, params.GlobalItemID, params.ItemType, params.ItemID, params.Title, params.Description)
		if err != nil {
			logger.Warn("Failed to create bookmark: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create bookmark: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":     true,
			"bookmark":    bookmark,
			"memory_card": memoryCard,
		}

		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createGetBookmarkHandler creates a handler for RPCGetBookmark
func createGetBookmarkHandler(service *notes.BookmarkService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetBookmark handler invoked")

		var params struct {
			ID uint `json:"id"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		bookmark, err := service.GetBookmarkByID(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to get bookmark: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get bookmark: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"bookmark": bookmark})
	}
}

// createUpdateBookmarkHandler creates a handler for RPCUpdateBookmark
func createUpdateBookmarkHandler(service *notes.BookmarkService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCUpdateBookmark handler invoked")

		var params notes.BookmarkModel

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		err := service.UpdateBookmark(ctx, &params)
		if err != nil {
			logger.Warn("Failed to update bookmark: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update bookmark: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"success": true})
	}
}

// createDeleteBookmarkHandler creates a handler for RPCDeleteBookmark
func createDeleteBookmarkHandler(service *notes.BookmarkService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCDeleteBookmark handler invoked")

		var params struct {
			ID uint `json:"id"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		err := service.DeleteBookmark(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to delete bookmark: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete bookmark: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"success": true})
	}
}

// createListBookmarksHandler creates a handler for RPCListBookmarks
func createListBookmarksHandler(service *notes.BookmarkService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCListBookmarks handler invoked")

		var params struct {
			State  string `json:"state"`
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
		}

		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
				logger.Warn("Failed to parse request: %v", errResp.Error)
				return errResp, nil
			}
		}

		if params.Limit <= 0 || params.Limit > 1000 {
			params.Limit = 100
		}
		if params.Offset < 0 {
			params.Offset = 0
		}

		bookmarks, total, err := service.ListBookmarks(ctx, params.State, params.Offset, params.Limit)
		if err != nil {
			logger.Warn("Failed to list bookmarks: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list bookmarks: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"bookmarks": bookmarks,
			"total":     total,
			"offset":    params.Offset,
			"limit":     params.Limit,
		})
	}
}

// createAddNoteHandler creates a handler for RPCAddNote
func createAddNoteHandler(service *notes.NoteService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCAddNote handler invoked")

		var params struct {
			BookmarkID uint    `json:"bookmark_id"`
			Content    string  `json:"content"`
			Author     *string `json:"author"`
			IsPrivate  bool    `json:"is_private"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.BookmarkID == 0 {
			logger.Warn("bookmark_id is required")
			return subprocess.NewErrorResponse(msg, "bookmark_id is required"), nil
		}
		if params.Content == "" {
			logger.Warn("content is required")
			return subprocess.NewErrorResponse(msg, "content is required"), nil
		}

		note, err := service.AddNote(ctx, params.BookmarkID, params.Content, params.Author, params.IsPrivate)
		if err != nil {
			logger.Warn("Failed to add note: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to add note: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"note": note})
	}
}

// createGetNoteHandler creates a handler for RPCGetNote
func createGetNoteHandler(service *notes.NoteService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetNote handler invoked")

		var params struct {
			ID uint `json:"id"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		note, err := service.GetNoteByID(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to get note: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get note: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"note": note})
	}
}

// createUpdateNoteHandler creates a handler for RPCUpdateNote
func createUpdateNoteHandler(service *notes.NoteService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCUpdateNote handler invoked")

		var params notes.NoteModel

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		err := service.UpdateNote(ctx, &params)
		if err != nil {
			logger.Warn("Failed to update note: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update note: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"success": true})
	}
}

// createDeleteNoteHandler creates a handler for RPCDeleteNote
func createDeleteNoteHandler(service *notes.NoteService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCDeleteNote handler invoked")

		var params struct {
			ID uint `json:"id"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.ID == 0 {
			logger.Warn("id is required")
			return subprocess.NewErrorResponse(msg, "id is required"), nil
		}

		err := service.DeleteNote(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to delete note: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete note: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"success": true})
	}
}

// createGetNotesByBookmarkHandler creates a handler for RPCGetNotesByBookmark
func createGetNotesByBookmarkHandler(service *notes.NoteService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetNotesByBookmark handler invoked")

		var params struct {
			BookmarkID uint `json:"bookmark_id"`
		}

		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}

		if params.BookmarkID == 0 {
			logger.Warn("bookmark_id is required")
			return subprocess.NewErrorResponse(msg, "bookmark_id is required"), nil
		}

		notes, err := service.GetNotesByBookmarkID(ctx, params.BookmarkID)
		if err != nil {
			logger.Warn("Failed to get notes by bookmark: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get notes: %v", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{"notes": notes})
	}
}
