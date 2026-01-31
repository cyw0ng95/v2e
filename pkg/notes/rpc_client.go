package notes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// RPCClient provides RPC-based access to the notes services
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*requestEntry
	mu              sync.Mutex
	logger          interface{} // Logger interface - can be common.Logger or any compatible logger
	correlationSeq  int64
}

// requestEntry holds the response channel for a pending request
type requestEntry struct {
	resp chan *subprocess.Message
}

// signal sends a message to the response channel
func (e *requestEntry) signal(msg *subprocess.Message) {
	select {
	case e.resp <- msg:
	default:
		// Channel is full or closed, ignore
	}
}

// close closes the response channel
func (e *requestEntry) close() {
	close(e.resp)
}

// NewRPCClient creates a new RPC client for the notes services
func NewRPCClient(sp *subprocess.Subprocess) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
	}

	// Register handlers for response and error messages
	sp.RegisterHandler("response", client.handleResponse)
	sp.RegisterHandler("error", client.handleError)

	return client
}

// handleResponse handles response messages from other services
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending request entry and remove it while holding the lock
	c.mu.Lock()
	entry := c.pendingRequests[msg.CorrelationID]
	if entry != nil {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if entry != nil {
		entry.signal(msg)
	}
	// Don't send another response
	return nil, nil
}

// handleError handles error messages from other services (treat them as responses)
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Error messages are also valid responses
	return c.handleResponse(ctx, msg)
}

// InvokeRPC invokes an RPC method on another service through the broker
func (c *RPCClient) InvokeRPC(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	// Generate correlation ID
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("notes-rpc-%d-%d", time.Now().UnixNano(), c.correlationSeq)
	c.mu.Unlock()

	// Create response channel and entry
	resp := make(chan *subprocess.Message, 1)
	entry := &requestEntry{resp: resp}

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()

	// Clean up on exit: remove from map and close entry
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		entry.close()
	}()

	// Create request message
	var payload []byte
	if params != nil {
		data, err := subprocess.MarshalFast(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		payload = data
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            method,
		Payload:       payload,
		Target:        target,
		CorrelationID: correlationID,
		Source:        c.sp.ID,
	}

	// Send request to broker (which will route to target)
	if err := c.sp.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}

	// Wait for response with timeout
	select {
	case response := <-resp:
		return response, nil
	case <-time.After(30 * time.Second): // Default timeout
		return nil, fmt.Errorf("RPC timeout waiting for response from %s", target)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// BookmarkServiceRPCClient implements the BookmarkService interface using RPC
type BookmarkServiceRPCClient struct {
	Client *RPCClient
}

// NewBookmarkServiceRPCClient creates a new RPC client for the bookmark service
func NewBookmarkServiceRPCClient(rpcClient *RPCClient) *BookmarkServiceRPCClient {
	return &BookmarkServiceRPCClient{
		Client: rpcClient,
	}
}

// CreateBookmark creates a new bookmark via RPC
func (s *BookmarkServiceRPCClient) CreateBookmark(ctx context.Context, globalItemID, itemType, itemID, title, description string) (*BookmarkModel, error) {
	params := map[string]interface{}{
		"global_item_id": globalItemID,
		"item_type":      itemType,
		"item_id":        itemID,
		"title":          title,
		"description":    description,
	}

	var result struct {
		Bookmark *BookmarkModel `json:"bookmark"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCCreateBookmark", params)
	if err != nil {
		return nil, fmt.Errorf("failed to create bookmark via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Bookmark, nil
}

// GetBookmarkByID retrieves a bookmark by ID via RPC
func (s *BookmarkServiceRPCClient) GetBookmarkByID(ctx context.Context, id uint) (*BookmarkModel, error) {
	params := map[string]interface{}{
		"id": id,
	}

	var result struct {
		Bookmark *BookmarkModel `json:"bookmark"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetBookmarkByID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark by ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Bookmark, nil
}

// GetBookmarksByGlobalItemID retrieves bookmarks by global item ID via RPC
func (s *BookmarkServiceRPCClient) GetBookmarksByGlobalItemID(ctx context.Context, globalItemID string) ([]*BookmarkModel, error) {
	params := map[string]interface{}{
		"global_item_id": globalItemID,
	}

	var result struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetBookmarksByGlobalItemID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks by global item ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Bookmarks, nil
}

// UpdateBookmark updates a bookmark via RPC
func (s *BookmarkServiceRPCClient) UpdateBookmark(ctx context.Context, bookmark *BookmarkModel) error {
	params := map[string]interface{}{
		"id":          bookmark.ID,
		"title":       bookmark.Title,
		"description": bookmark.Description,
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCUpdateBookmark", params)
	if err != nil {
		return fmt.Errorf("failed to update bookmark via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	return nil
}

// UpdateLearningState updates the learning state of a bookmark via RPC
func (s *BookmarkServiceRPCClient) UpdateLearningState(ctx context.Context, bookmarkID uint, newState LearningState) error {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
		"new_state":   string(newState),
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCUpdateLearningState", params)
	if err != nil {
		return fmt.Errorf("failed to update learning state via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetBookmarksByLearningState retrieves bookmarks by learning state via RPC
func (s *BookmarkServiceRPCClient) GetBookmarksByLearningState(ctx context.Context, state LearningState) ([]*BookmarkModel, error) {
	params := map[string]interface{}{
		"state": string(state),
	}

	var result struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
		Total     int64            `json:"total"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetBookmarksByLearningState", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks by learning state via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Bookmarks, nil
}

// ListBookmarks lists bookmarks with pagination and optional state filtering via RPC
func (s *BookmarkServiceRPCClient) ListBookmarks(ctx context.Context, state string, offset, limit int) ([]*BookmarkModel, int64, error) {
	params := map[string]interface{}{
		"state":  state,
		"offset": offset,
		"limit":  limit,
	}

	var result struct {
		Bookmarks []*BookmarkModel `json:"bookmarks"`
		Total     int64            `json:"total"`
		Offset    int              `json:"offset"`
		Limit     int              `json:"limit"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCListBookmarks", params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookmarks via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, 0, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Bookmarks, result.Total, nil
}

// DeleteBookmark deletes a bookmark via RPC
func (s *BookmarkServiceRPCClient) DeleteBookmark(ctx context.Context, id uint) error {
	params := map[string]interface{}{
		"id": id,
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCDeleteBookmark", params)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// NoteServiceRPCClient implements the NoteService interface using RPC
type NoteServiceRPCClient struct {
	Client *RPCClient
}

// NewNoteServiceRPCClient creates a new RPC client for the note service
func NewNoteServiceRPCClient(rpcClient *RPCClient) *NoteServiceRPCClient {
	return &NoteServiceRPCClient{
		Client: rpcClient,
	}
}

// AddNote adds a note to a bookmark via RPC
func (s *NoteServiceRPCClient) AddNote(ctx context.Context, bookmarkID uint, content string, author *string, isPrivate bool) (*NoteModel, error) {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
		"content":     content,
		"author":      author,
		"is_private":  isPrivate,
	}

	var result struct {
		Note *NoteModel `json:"note"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCAddNote", params)
	if err != nil {
		return nil, fmt.Errorf("failed to add note via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Note, nil
}

// GetNotesByBookmarkID retrieves notes for a bookmark via RPC
func (s *NoteServiceRPCClient) GetNotesByBookmarkID(ctx context.Context, bookmarkID uint) ([]*NoteModel, error) {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
	}

	var result struct {
		Notes []*NoteModel `json:"notes"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetNotesByBookmarkID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by bookmark ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Notes, nil
}

// UpdateNote updates a note via RPC
func (s *NoteServiceRPCClient) UpdateNote(ctx context.Context, note *NoteModel) error {
	params := map[string]interface{}{
		"id":          note.ID,
		"bookmark_id": note.BookmarkID,
		"content":     note.Content,
		"author":      note.Author,
		"is_private":  note.IsPrivate,
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCUpdateNote", params)
	if err != nil {
		return fmt.Errorf("failed to update note via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetNoteByID retrieves a note by ID via RPC
func (s *NoteServiceRPCClient) GetNoteByID(ctx context.Context, id uint) (*NoteModel, error) {
	params := map[string]interface{}{
		"id": id,
	}

	var result struct {
		Note *NoteModel `json:"note"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetNoteByID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get note by ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Note, nil
}

// DeleteNote deletes a note via RPC
func (s *NoteServiceRPCClient) DeleteNote(ctx context.Context, id uint) error {
	params := map[string]interface{}{
		"id": id,
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCDeleteNote", params)
	if err != nil {
		return fmt.Errorf("failed to delete note via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// MemoryCardServiceRPCClient implements the MemoryCardService interface using RPC
type MemoryCardServiceRPCClient struct {
	Client *RPCClient
}

// NewMemoryCardServiceRPCClient creates a new RPC client for the memory card service
func NewMemoryCardServiceRPCClient(rpcClient *RPCClient) *MemoryCardServiceRPCClient {
	return &MemoryCardServiceRPCClient{
		Client: rpcClient,
	}
}

// CreateMemoryCard creates a new memory card via RPC
func (s *MemoryCardServiceRPCClient) CreateMemoryCard(ctx context.Context, bookmarkID uint, front, back string) (*MemoryCardModel, error) {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
		"front":       front,
		"back":        back,
	}

	var result struct {
		Card *MemoryCardModel `json:"card"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCCreateMemoryCard", params)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory card via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Card, nil
}

// GetMemoryCardsByBookmarkID retrieves memory cards for a bookmark via RPC
func (s *MemoryCardServiceRPCClient) GetMemoryCardsByBookmarkID(ctx context.Context, bookmarkID uint) ([]*MemoryCardModel, error) {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
	}

	var result struct {
		Cards []*MemoryCardModel `json:"cards"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetMemoryCardsByBookmarkID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory cards by bookmark ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Cards, nil
}

// GetCardsForReview retrieves cards that are due for review via RPC
func (s *MemoryCardServiceRPCClient) GetCardsForReview(ctx context.Context) ([]*MemoryCardModel, error) {
	params := map[string]interface{}{}

	var result struct {
		Cards []*MemoryCardModel `json:"cards"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetCardsForReview", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards for review via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Cards, nil
}

// GetCardsByLearningState retrieves cards by learning state via RPC
func (s *MemoryCardServiceRPCClient) GetCardsByLearningState(ctx context.Context, state LearningState) ([]*MemoryCardModel, error) {
	params := map[string]interface{}{
		"state": string(state),
	}

	var result struct {
		Cards []*MemoryCardModel `json:"cards"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetCardsByLearningState", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards by learning state via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Cards, nil
}

// UpdateCardAfterReview updates a card after user review via RPC
func (s *MemoryCardServiceRPCClient) UpdateCardAfterReview(ctx context.Context, cardID uint, rating CardRating) error {
	params := map[string]interface{}{
		"card_id": cardID,
		"rating":  string(rating),
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCUpdateCardAfterReview", params)
	if err != nil {
		return fmt.Errorf("failed to update card after review via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// ListMemoryCards retrieves memory cards with filters via RPC
func (s *MemoryCardServiceRPCClient) ListMemoryCards(ctx context.Context, bookmarkID *uint, learningState *string, author *string, isPrivate *bool, offset, limit int) ([]*MemoryCardModel, int64, error) {
	params := map[string]interface{}{
		"offset": offset,
		"limit":  limit,
	}

	if bookmarkID != nil {
		params["bookmark_id"] = float64(*bookmarkID)
	}
	if learningState != nil {
		params["learning_state"] = *learningState
	}
	if author != nil {
		params["author"] = *author
	}
	if isPrivate != nil {
		params["is_private"] = *isPrivate
	}

	var result struct {
		MemoryCards []*MemoryCardModel `json:"memory_cards"`
		Offset      int                `json:"offset"`
		Limit       int                `json:"limit"`
		Total       int64              `json:"total"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCListMemoryCards", params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memory cards via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, 0, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.MemoryCards, result.Total, nil
}

// CrossReferenceServiceRPCClient implements the CrossReferenceService interface using RPC
type CrossReferenceServiceRPCClient struct {
	Client *RPCClient
}

// NewCrossReferenceServiceRPCClient creates a new RPC client for the cross-reference service
func NewCrossReferenceServiceRPCClient(rpcClient *RPCClient) *CrossReferenceServiceRPCClient {
	return &CrossReferenceServiceRPCClient{
		Client: rpcClient,
	}
}

// CreateCrossReference creates a new cross-reference via RPC
func (s *CrossReferenceServiceRPCClient) CreateCrossReference(ctx context.Context, sourceItemID, targetItemID, sourceType, targetType, relationshipType string, strength float32, description *string) (*CrossReferenceModel, error) {
	params := map[string]interface{}{
		"source_item_id":    sourceItemID,
		"target_item_id":    targetItemID,
		"source_type":       sourceType,
		"target_type":       targetType,
		"relationship_type": relationshipType,
		"strength":          strength,
		"description":       description,
	}

	var result struct {
		CrossReference *CrossReferenceModel `json:"cross_reference"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCCreateCrossReference", params)
	if err != nil {
		return nil, fmt.Errorf("failed to create cross-reference via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.CrossReference, nil
}

// GetCrossReferencesBySource retrieves cross-references by source via RPC
func (s *CrossReferenceServiceRPCClient) GetCrossReferencesBySource(ctx context.Context, sourceItemID string) ([]*CrossReferenceModel, error) {
	params := map[string]interface{}{
		"source_item_id": sourceItemID,
	}

	var result struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetCrossReferencesBySource", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by source via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.CrossReferences, nil
}

// GetCrossReferencesByTarget retrieves cross-references by target via RPC
func (s *CrossReferenceServiceRPCClient) GetCrossReferencesByTarget(ctx context.Context, targetItemID string) ([]*CrossReferenceModel, error) {
	params := map[string]interface{}{
		"target_item_id": targetItemID,
	}

	var result struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetCrossReferencesByTarget", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by target via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.CrossReferences, nil
}

// GetCrossReferencesByType retrieves cross-references by type via RPC
func (s *CrossReferenceServiceRPCClient) GetCrossReferencesByType(ctx context.Context, relationshipType RelationshipType) ([]*CrossReferenceModel, error) {
	params := map[string]interface{}{
		"relationship_type": string(relationshipType),
	}

	var result struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetCrossReferencesByType", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by type via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.CrossReferences, nil
}

// GetBidirectionalCrossReferences retrieves cross-references in both directions between two items via RPC
func (s *CrossReferenceServiceRPCClient) GetBidirectionalCrossReferences(ctx context.Context, itemID1, itemID2 string) ([]*CrossReferenceModel, error) {
	params := map[string]interface{}{
		"item_id_1": itemID1,
		"item_id_2": itemID2,
	}

	var result struct {
		CrossReferences []*CrossReferenceModel `json:"cross_references"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetBidirectionalCrossReferences", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bidirectional cross-references via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.CrossReferences, nil
}

// HistoryServiceRPCClient implements the HistoryService interface using RPC
type HistoryServiceRPCClient struct {
	Client *RPCClient
}

// NewHistoryServiceRPCClient creates a new RPC client for the history service
func NewHistoryServiceRPCClient(rpcClient *RPCClient) *HistoryServiceRPCClient {
	return &HistoryServiceRPCClient{
		Client: rpcClient,
	}
}

// GetHistoryByBookmarkID retrieves history for a bookmark via RPC
func (s *HistoryServiceRPCClient) GetHistoryByBookmarkID(ctx context.Context, bookmarkID uint) ([]*BookmarkHistoryModel, error) {
	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
	}

	var result struct {
		History []*BookmarkHistoryModel `json:"history"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCGetHistoryByBookmarkID", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get history by bookmark ID via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.History, nil
}

// RevertBookmarkState reverts a bookmark to a previous state via RPC
func (s *HistoryServiceRPCClient) RevertBookmarkState(ctx context.Context, bookmarkID uint, timestamp interface{}) error {
	var timestampStr string
	switch t := timestamp.(type) {
	case string:
		timestampStr = t
	case json.RawMessage:
		timestampStr = string(t)
	default:
		return fmt.Errorf("invalid timestamp type for RPC call")
	}

	params := map[string]interface{}{
		"bookmark_id": bookmarkID,
		"timestamp":   timestampStr,
	}

	var result struct {
		Success bool `json:"success"`
	}

	response, err := s.Client.InvokeRPC(ctx, "local", "RPCRevertBookmarkState", params)
	if err != nil {
		return fmt.Errorf("failed to revert bookmark state via RPC: %w", err)
	}

	if response.Type == subprocess.MessageTypeError {
		return fmt.Errorf("remote error: %s", response.Error)
	}

	if err := subprocess.UnmarshalPayload(response, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetRPCClients returns RPC clients for all notes services
func (c *RPCClient) GetRPCClients() (
	*BookmarkServiceRPCClient,
	*NoteServiceRPCClient,
	*MemoryCardServiceRPCClient,
	*CrossReferenceServiceRPCClient,
	*HistoryServiceRPCClient,
) {
	return NewBookmarkServiceRPCClient(c),
		NewNoteServiceRPCClient(c),
		NewMemoryCardServiceRPCClient(c),
		NewCrossReferenceServiceRPCClient(c),
		NewHistoryServiceRPCClient(c)
}
