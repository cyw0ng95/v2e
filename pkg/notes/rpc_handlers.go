package notes

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// RPCHandlers provides RPC handlers for the notes services
type RPCHandlers struct {
	container *ServiceContainer
	sp        *subprocess.Subprocess
	logger    *common.Logger
}

// NewRPCHandlers creates a new RPC handlers instance
func NewRPCHandlers(container *ServiceContainer, sp *subprocess.Subprocess, logger *common.Logger) *RPCHandlers {
	handlers := &RPCHandlers{
		container: container,
		sp:        sp,
		logger:    logger,
	}

	// Register RPC methods
	sp.RegisterHandler("RPCCreateBookmark", handlers.handleRPCCreateBookmark)
	sp.RegisterHandler("RPCGetBookmarkByID", handlers.handleRPCGetBookmarkByID)
	sp.RegisterHandler("RPCGetBookmarksByGlobalItemID", handlers.handleRPCGetBookmarksByGlobalItemID)
	sp.RegisterHandler("RPCUpdateBookmark", handlers.handleRPCUpdateBookmark)
	sp.RegisterHandler("RPCUpdateLearningState", handlers.handleRPCUpdateLearningState)
	sp.RegisterHandler("RPCGetBookmarksByLearningState", handlers.handleRPCGetBookmarksByLearningState)
	sp.RegisterHandler("RPCListBookmarks", handlers.handleRPCListBookmarks)
	sp.RegisterHandler("RPCDeleteBookmark", handlers.handleRPCDeleteBookmark)
	sp.RegisterHandler("RPCAddNote", handlers.handleRPCAddNote)
	sp.RegisterHandler("RPCGetNoteByID", handlers.handleRPCGetNoteByID)
	sp.RegisterHandler("RPCGetNotesByBookmarkID", handlers.handleRPCGetNotesByBookmarkID)
	sp.RegisterHandler("RPCUpdateNote", handlers.handleRPCUpdateNote)
	sp.RegisterHandler("RPCDeleteNote", handlers.handleRPCDeleteNote)
	sp.RegisterHandler("RPCCreateMemoryCard", handlers.handleRPCCreateMemoryCard)
	sp.RegisterHandler("RPCGetMemoryCardsByBookmarkID", handlers.handleRPCGetMemoryCardsByBookmarkID)
	sp.RegisterHandler("RPCGetCardsForReview", handlers.handleRPCGetCardsForReview)
	sp.RegisterHandler("RPCGetCardsByLearningState", handlers.handleRPCGetCardsByLearningState)
	sp.RegisterHandler("RPCUpdateCardAfterReview", handlers.handleRPCUpdateCardAfterReview)
	sp.RegisterHandler("RPCCreateCrossReference", handlers.handleRPCCreateCrossReference)
	sp.RegisterHandler("RPCGetCrossReferencesBySource", handlers.handleRPCGetCrossReferencesBySource)
	sp.RegisterHandler("RPCGetCrossReferencesByTarget", handlers.handleRPCGetCrossReferencesByTarget)
	sp.RegisterHandler("RPCGetCrossReferencesByType", handlers.handleRPCGetCrossReferencesByType)
	sp.RegisterHandler("RPCGetBidirectionalCrossReferences", handlers.handleRPCGetBidirectionalCrossReferences)
	sp.RegisterHandler("RPCGetHistoryByBookmarkID", handlers.handleRPCGetHistoryByBookmarkID)
	sp.RegisterHandler("RPCRevertBookmarkState", handlers.handleRPCRevertBookmarkState)
	sp.RegisterHandler("RPCListMemoryCards", handlers.handleRPCListMemoryCards)
	sp.RegisterHandler("RPCUpdateBookmarkStats", handlers.handleRPCUpdateBookmarkStats)
	sp.RegisterHandler("RPCGetBookmarkStats", handlers.handleRPCGetBookmarkStats)

	return handlers
}

// handleRPCCreateBookmark handles RPC request to create a bookmark
func (h *RPCHandlers) handleRPCCreateBookmark(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	globalItemID, ok := params["global_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid global_item_id"), nil
	}
	itemType, ok := params["item_type"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid item_type"), nil
	}
	itemID, ok := params["item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid item_id"), nil
	}
	title, ok := params["title"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid title"), nil
	}
	description, ok := params["description"].(string)
	if !ok {
		description = ""
	}

	bookmark, err := h.container.BookmarkService.CreateBookmark(ctx, globalItemID, itemType, itemID, title, description)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to create bookmark: %v", err)), nil
	}

	result := struct {
		Bookmark *BookmarkModel `json:"bookmark"`
	}{
		Bookmark: bookmark,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetBookmarkByID handles RPC request to get a bookmark by ID
func (h *RPCHandlers) handleRPCGetBookmarkByID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	bookmark, err := h.container.BookmarkService.GetBookmarkByID(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get bookmark: %v", err)), nil
	}

	result := struct {
		Bookmark *BookmarkModel `json:"bookmark"`
	}{
		Bookmark: bookmark,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetBookmarksByGlobalItemID handles RPC request to get bookmarks by global item ID
func (h *RPCHandlers) handleRPCGetBookmarksByGlobalItemID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	globalItemID, ok := params["global_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid global_item_id"), nil
	}

	bookmarks, err := h.container.BookmarkService.GetBookmarksByGlobalItemID(ctx, globalItemID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get bookmarks: %v", err)), nil
	}

	result := struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
	}{
		Bookmarks: bookmarks,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCUpdateBookmark handles RPC request to update a bookmark
func (h *RPCHandlers) handleRPCUpdateBookmark(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	title, ok := params["title"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid title"), nil
	}

	description, ok := params["description"].(string)
	if !ok {
		description = ""
	}

	// Get the existing bookmark first
	existingBookmark, err := h.container.BookmarkService.GetBookmarkByID(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Bookmark not found: %v", err)), nil
	}

	// Update the fields
	existingBookmark.Title = title
	existingBookmark.Description = description

	err = h.container.BookmarkService.UpdateBookmark(ctx, existingBookmark)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to update bookmark: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCUpdateLearningState handles RPC request to update learning state
func (h *RPCHandlers) handleRPCUpdateLearningState(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	newStateStr, ok := params["new_state"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid new_state"), nil
	}
	newState := LearningState(newStateStr)

	err := h.container.BookmarkService.UpdateLearningState(ctx, bookmarkID, newState)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to update learning state: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetBookmarksByLearningState handles RPC request to get bookmarks by learning state
func (h *RPCHandlers) handleRPCGetBookmarksByLearningState(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	stateStr, ok := params["state"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid state"), nil
	}
	state := LearningState(stateStr)

	bookmarks, err := h.container.BookmarkService.GetBookmarksByLearningState(ctx, state)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get bookmarks by learning state: %v", err)), nil
	}

	result := struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
		Total     int64            `json:"total"`
	}{
		Bookmarks: bookmarks,
		Total:     int64(len(bookmarks)),
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCListBookmarks handles RPC request to list bookmarks
func (h *RPCHandlers) handleRPCListBookmarks(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	state, ok := params["state"].(string)
	if !ok {
		state = ""
	}

	offsetFloat, ok := params["offset"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid offset"), nil
	}
	offset := int(offsetFloat)

	limitFloat, ok := params["limit"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid limit"), nil
	}
	limit := int(limitFloat)

	bookmarks, total, err := h.container.BookmarkService.ListBookmarks(ctx, state, offset, limit)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to list bookmarks: %v", err)), nil
	}

	result := struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
		Total     int64            `json:"total"`
		Offset    int              `json:"offset"`
		Limit     int              `json:"limit"`
	}{
		Bookmarks: bookmarks,
		Total:     total,
		Offset:    offset,
		Limit:     limit,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCDeleteBookmark handles RPC request to delete a bookmark
func (h *RPCHandlers) handleRPCDeleteBookmark(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	err := h.container.BookmarkService.DeleteBookmark(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to delete bookmark: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCAddNote handles RPC request to add a note
func (h *RPCHandlers) handleRPCAddNote(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	content, ok := params["content"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid content"), nil
	}

	var author *string
	if authorParam, exists := params["author"]; exists && authorParam != nil {
		if authorStr, ok := authorParam.(string); ok {
			author = &authorStr
		} else {
			return h.createErrorResponse(msg, "Invalid author parameter"), nil
		}
	}

	isPrivate, ok := params["is_private"].(bool)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid is_private"), nil
	}

	note, err := h.container.NoteService.AddNote(ctx, bookmarkID, content, author, isPrivate)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to add note: %v", err)), nil
	}

	result := struct {
		Note *NoteModel `json:"note"`
	}{
		Note: note,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetNoteByID handles RPC request to get a note by ID
func (h *RPCHandlers) handleRPCGetNoteByID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	note, err := h.container.NoteService.GetNoteByID(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get note: %v", err)), nil
	}

	result := struct {
		Note *NoteModel `json:"note"`
	}{
		Note: note,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetNotesByBookmarkID handles RPC request to get notes by bookmark ID
func (h *RPCHandlers) handleRPCGetNotesByBookmarkID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	notes, err := h.container.NoteService.GetNotesByBookmarkID(ctx, bookmarkID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get notes: %v", err)), nil
	}

	result := struct {
		Notes []*NoteModel `json:"notes"`
	}{
		Notes: notes,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCUpdateNote handles RPC request to update a note
func (h *RPCHandlers) handleRPCUpdateNote(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	content, ok := params["content"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid content"), nil
	}

	var author *string
	if authorParam, exists := params["author"]; exists && authorParam != nil {
		if authorStr, ok := authorParam.(string); ok {
			author = &authorStr
		} else {
			return h.createErrorResponse(msg, "Invalid author parameter"), nil
		}
	}

	isPrivate, ok := params["is_private"].(bool)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid is_private"), nil
	}

	// Get the existing note first
	existingNote, err := h.container.NoteService.GetNoteByID(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Note not found: %v", err)), nil
	}

	// Update the fields
	existingNote.BookmarkID = bookmarkID
	existingNote.Content = content
	existingNote.Author = author
	existingNote.IsPrivate = isPrivate

	err = h.container.NoteService.UpdateNote(ctx, existingNote)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to update note: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCDeleteNote handles RPC request to delete a note
func (h *RPCHandlers) handleRPCDeleteNote(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	idFloat, ok := params["id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid id"), nil
	}
	id := uint(idFloat)

	err := h.container.NoteService.DeleteNote(ctx, id)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to delete note: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCCreateMemoryCard handles RPC request to create a memory card
func (h *RPCHandlers) handleRPCCreateMemoryCard(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	front, ok := params["front"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid front"), nil
	}

	back, ok := params["back"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid back"), nil
	}

	card, err := h.container.MemoryCardService.CreateMemoryCard(ctx, bookmarkID, front, back)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to create memory card: %v", err)), nil
	}

	result := struct {
		Card *MemoryCardModel `json:"card"`
	}{
		Card: card,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetMemoryCardsByBookmarkID handles RPC request to get memory cards by bookmark ID
func (h *RPCHandlers) handleRPCGetMemoryCardsByBookmarkID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	cards, err := h.container.MemoryCardService.GetMemoryCardsByBookmarkID(ctx, bookmarkID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get memory cards: %v", err)), nil
	}

	result := struct {
		Cards []*MemoryCardModel `json:"cards"`
	}{
		Cards: cards,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetCardsForReview handles RPC request to get cards for review
func (h *RPCHandlers) handleRPCGetCardsForReview(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	cards, err := h.container.MemoryCardService.GetCardsForReview(ctx)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get cards for review: %v", err)), nil
	}

	result := struct {
		Cards []*MemoryCardModel `json:"cards"`
	}{
		Cards: cards,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetCardsByLearningState handles RPC request to get cards by learning state
func (h *RPCHandlers) handleRPCGetCardsByLearningState(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	stateStr, ok := params["state"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid state"), nil
	}
	state := LearningState(stateStr)

	cards, err := h.container.MemoryCardService.GetCardsByLearningState(ctx, state)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get cards by learning state: %v", err)), nil
	}

	result := struct {
		Cards []*MemoryCardModel `json:"cards"`
	}{
		Cards: cards,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCUpdateCardAfterReview handles RPC request to update a card after review
func (h *RPCHandlers) handleRPCUpdateCardAfterReview(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	cardIDFloat, ok := params["card_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid card_id"), nil
	}
	cardID := uint(cardIDFloat)

	ratingStr, ok := params["rating"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid rating"), nil
	}
	rating := CardRating(ratingStr)

	err := h.container.MemoryCardService.UpdateCardAfterReview(ctx, cardID, rating)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to update card after review: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCCreateCrossReference handles RPC request to create a cross-reference
func (h *RPCHandlers) handleRPCCreateCrossReference(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	sourceItemID, ok := params["source_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid source_item_id"), nil
	}

	targetItemID, ok := params["target_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid target_item_id"), nil
	}

	sourceType, ok := params["source_type"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid source_type"), nil
	}

	targetType, ok := params["target_type"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid target_type"), nil
	}

	relationshipType, ok := params["relationship_type"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid relationship_type"), nil
	}

	strengthFloat, ok := params["strength"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid strength"), nil
	}
	strength := float32(strengthFloat)

	var description *string
	if descParam, exists := params["description"]; exists && descParam != nil {
		if descStr, ok := descParam.(string); ok {
			description = &descStr
		} else {
			return h.createErrorResponse(msg, "Invalid description parameter"), nil
		}
	}

	crossRef, err := h.container.CrossReferenceService.CreateCrossReference(ctx, sourceItemID, targetItemID, sourceType, targetType, relationshipType, strength, description)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to create cross-reference: %v", err)), nil
	}

	result := struct {
		CrossReference *CrossReferenceModel `json:"cross_reference"`
	}{
		CrossReference: crossRef,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetCrossReferencesBySource handles RPC request to get cross-references by source
func (h *RPCHandlers) handleRPCGetCrossReferencesBySource(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	sourceItemID, ok := params["source_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid source_item_id"), nil
	}

	crossRefs, err := h.container.CrossReferenceService.GetCrossReferencesBySource(ctx, sourceItemID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get cross-references by source: %v", err)), nil
	}

	result := struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}{
		CrossReferences: crossRefs,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetCrossReferencesByTarget handles RPC request to get cross-references by target
func (h *RPCHandlers) handleRPCGetCrossReferencesByTarget(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	targetItemID, ok := params["target_item_id"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid target_item_id"), nil
	}

	crossRefs, err := h.container.CrossReferenceService.GetCrossReferencesByTarget(ctx, targetItemID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get cross-references by target: %v", err)), nil
	}

	result := struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}{
		CrossReferences: crossRefs,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetCrossReferencesByType handles RPC request to get cross-references by type
func (h *RPCHandlers) handleRPCGetCrossReferencesByType(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	relationshipTypeStr, ok := params["relationship_type"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid relationship_type"), nil
	}
	relationshipType := RelationshipType(relationshipTypeStr)

	crossRefs, err := h.container.CrossReferenceService.GetCrossReferencesByType(ctx, relationshipType)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get cross-references by type: %v", err)), nil
	}

	result := struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}{
		CrossReferences: crossRefs,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetBidirectionalCrossReferences handles RPC request to get bidirectional cross-references
func (h *RPCHandlers) handleRPCGetBidirectionalCrossReferences(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	itemID1, ok := params["item_id_1"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid item_id_1"), nil
	}

	itemID2, ok := params["item_id_2"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid item_id_2"), nil
	}

	crossRefs, err := h.container.CrossReferenceService.GetBidirectionalCrossReferences(ctx, itemID1, itemID2)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get bidirectional cross-references: %v", err)), nil
	}

	result := struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}{
		CrossReferences: crossRefs,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetHistoryByBookmarkID handles RPC request to get history by bookmark ID
func (h *RPCHandlers) handleRPCGetHistoryByBookmarkID(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	history, err := h.container.HistoryService.GetHistoryByBookmarkID(ctx, bookmarkID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get history: %v", err)), nil
	}

	result := struct {
		History []*BookmarkHistoryModel `json:"history"`
	}{
		History: history,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCRevertBookmarkState handles RPC request to revert bookmark state
func (h *RPCHandlers) handleRPCRevertBookmarkState(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	timestamp, ok := params["timestamp"].(string)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid timestamp"), nil
	}

	err := h.container.HistoryService.RevertBookmarkState(ctx, bookmarkID, timestamp)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to revert bookmark state: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCListMemoryCards handles RPC request to list memory cards with filters
func (h *RPCHandlers) handleRPCListMemoryCards(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	// Extract optional parameters with defaults
	var bookmarkID *uint
	if bookmarkIDParam, exists := params["bookmark_id"]; exists && bookmarkIDParam != nil {
		if bookmarkIDFloat, ok := bookmarkIDParam.(float64); ok {
			id := uint(bookmarkIDFloat)
			bookmarkID = &id
		} else {
			return h.createErrorResponse(msg, "Invalid bookmark_id parameter"), nil
		}
	}

	var learningState *string
	if learningStateParam, exists := params["learning_state"]; exists && learningStateParam != nil {
		if stateStr, ok := learningStateParam.(string); ok {
			learningState = &stateStr
		} else {
			return h.createErrorResponse(msg, "Invalid learning_state parameter"), nil
		}
	}

	var author *string
	if authorParam, exists := params["author"]; exists && authorParam != nil {
		if authorStr, ok := authorParam.(string); ok {
			author = &authorStr
		} else {
			return h.createErrorResponse(msg, "Invalid author parameter"), nil
		}
	}

	var isPrivate *bool
	if isPrivateParam, exists := params["is_private"]; exists && isPrivateParam != nil {
		if privateBool, ok := isPrivateParam.(bool); ok {
			isPrivate = &privateBool
		} else {
			return h.createErrorResponse(msg, "Invalid is_private parameter"), nil
		}
	}

	// Extract pagination parameters with defaults
	offset := 0
	if offsetParam, exists := params["offset"]; exists {
		if offsetFloat, ok := offsetParam.(float64); ok {
			offset = int(offsetFloat)
		} else {
			return h.createErrorResponse(msg, "Invalid offset parameter"), nil
		}
	}

	limit := 50 // Default limit
	if limitParam, exists := params["limit"]; exists {
		if limitFloat, ok := limitParam.(float64); ok {
			limit = int(limitFloat)
		} else {
			return h.createErrorResponse(msg, "Invalid limit parameter"), nil
		}
	}

	// Call the service method
	cards, total, err := h.container.MemoryCardService.ListMemoryCards(ctx, bookmarkID, learningState, author, isPrivate, offset, limit)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to list memory cards: %v", err)), nil
	}

	result := struct {
		MemoryCards []*MemoryCardModel `json:"memory_cards"`
		Offset      int                `json:"offset"`
		Limit       int                `json:"limit"`
		Total       int64              `json:"total"`
	}{
		MemoryCards: cards,
		Offset:      offset,
		Limit:       limit,
		Total:       total,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCUpdateBookmarkStats handles RPC request to update bookmark statistics
func (h *RPCHandlers) handleRPCUpdateBookmarkStats(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	viewCountDeltaFloat, ok := params["view_count_delta"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid view_count_delta"), nil
	}
	viewCountDelta := int(viewCountDeltaFloat)

	studySessionDeltaFloat, ok := params["study_session_delta"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid study_session_delta"), nil
	}
	studySessionDelta := int(studySessionDeltaFloat)

	err := h.container.BookmarkService.UpdateBookmarkStats(ctx, bookmarkID, viewCountDelta, studySessionDelta)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to update bookmark stats: %v", err)), nil
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// handleRPCGetBookmarkStats handles RPC request to get bookmark statistics
func (h *RPCHandlers) handleRPCGetBookmarkStats(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to unmarshal params: %v", err)), nil
	}

	bookmarkIDFloat, ok := params["bookmark_id"].(float64)
	if !ok {
		return h.createErrorResponse(msg, "Missing or invalid bookmark_id"), nil
	}
	bookmarkID := uint(bookmarkIDFloat)

	stats, err := h.container.BookmarkService.GetBookmarkStats(ctx, bookmarkID)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to get bookmark stats: %v", err)), nil
	}

	result := struct {
		Stats map[string]interface{} `json:"stats"`
	}{
		Stats: stats,
	}

	payload, err := subprocess.MarshalFast(result)
	if err != nil {
		return h.createErrorResponse(msg, fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            msg.ID,
		Payload:       payload,
		Target:        msg.Source,
		CorrelationID: msg.CorrelationID,
		Source:        h.sp.ID,
	}, nil
}

// createErrorResponse creates an error response message
func (h *RPCHandlers) createErrorResponse(requestMsg *subprocess.Message, errorMsg string) *subprocess.Message {
	h.logger.Error(LogMsgRPCError, errorMsg)

	return &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		ID:            requestMsg.ID,
		Error:         errorMsg,
		Target:        requestMsg.Source,
		CorrelationID: requestMsg.CorrelationID,
		Source:        h.sp.ID,
	}
}
