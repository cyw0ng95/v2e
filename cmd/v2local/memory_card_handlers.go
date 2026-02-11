package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Handler for RPCCreateMemoryCard
func createMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			BookmarkID uint            `json:"bookmark_id"`
			Front      string          `json:"front_content"`
			Back       string          `json:"back_content"`
			MajorClass string          `json:"major_class"`
			MinorClass string          `json:"minor_class"`
			Status     string          `json:"status"`
			Content    json.RawMessage `json:"content"`
			CardType   string          `json:"card_type"`
			Author     string          `json:"author"`
			IsPrivate  bool            `json:"is_private"`
			Metadata   map[string]any  `json:"metadata"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		card, err := service.CreateMemoryCardFull(ctx, params.BookmarkID, params.Front, params.Back, params.MajorClass, params.MinorClass, params.Status, string(params.Content), params.CardType, params.Author, params.IsPrivate, params.Metadata)
		if err != nil {
			logger.Warn("Failed to create memory card: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create memory card: %v", err)), nil
		}
		resp := map[string]any{"success": true, "memory_card": card}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// Handler for RPCGetMemoryCard
func getMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
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
		card, err := service.GetMemoryCardByID(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to get memory card: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get memory card: %v", err)), nil
		}
		resp := map[string]any{"memory_card": card}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// Handler for RPCUpdateMemoryCard
func updateMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]any
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		// Defensive: validate status if provided
		if raw, ok := params["status"]; ok {
			if sstr, ok := raw.(string); ok {
				if _, err := notes.ParseCardStatus(sstr); err != nil {
					logger.Warn("Invalid status value: %v", err)
					return subprocess.NewErrorResponse(msg, fmt.Sprintf("invalid status: %v", err)), nil
				}
			} else {
				logger.Warn("status must be a string")
				return subprocess.NewErrorResponse(msg, "status must be a string"), nil
			}
		}

		card, err := service.UpdateMemoryCardFields(ctx, params)
		if err != nil {
			logger.Warn("Failed to update memory card: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update memory card: %v", err)), nil
		}
		resp := map[string]any{"success": true, "memory_card": card}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// Handler for RPCDeleteMemoryCard
func deleteMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
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
		err := service.DeleteMemoryCard(ctx, params.ID)
		if err != nil {
			logger.Warn("Failed to delete memory card: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete memory card: %v", err)), nil
		}
		resp := map[string]any{"success": true}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// Handler for RPCListMemoryCards
func listMemoryCardsHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			BookmarkID    *uint   `json:"bookmark_id"`
			MajorClass    *string `json:"major_class"`
			MinorClass    *string `json:"minor_class"`
			Status        *string `json:"status"`
			LearningState *string `json:"learning_state"`
			Author        *string `json:"author"`
			IsPrivate     *bool   `json:"is_private"`
			Offset        int     `json:"offset"`
			Limit         int     `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
				logger.Warn("Failed to parse request: %v", errResp.Error)
				return errResp, nil
			}
		}
		// Set defaults
		if params.Limit <= 0 || params.Limit > 1000 {
			params.Limit = 100
		}
		if params.Offset < 0 {
			params.Offset = 0
		}
		cards, total, err := service.ListMemoryCardsFull(ctx, params.BookmarkID, params.MajorClass, params.MinorClass, params.Status, params.LearningState, params.Author, params.IsPrivate, params.Offset, params.Limit)
		if err != nil {
			logger.Warn("Failed to list memory cards: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list memory cards: %v", err)), nil
		}
		resp := map[string]any{"memory_cards": cards, "total": total, "offset": params.Offset, "limit": params.Limit}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// Handler for RPCRateMemoryCard
func rateMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			CardID uint             `json:"card_id"`
			Rating notes.CardRating `json:"rating"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if params.CardID == 0 {
			logger.Warn("card_id is required")
			return subprocess.NewErrorResponse(msg, "card_id is required"), nil
		}
		if params.Rating == "" {
			logger.Warn("rating is required")
			return subprocess.NewErrorResponse(msg, "rating is required"), nil
		}

		// Validate rating value
		validRatings := map[notes.CardRating]bool{
			notes.CardRatingAgain: true,
			notes.CardRatingHard:  true,
			notes.CardRatingGood:  true,
			notes.CardRatingEasy:  true,
		}
		if !validRatings[params.Rating] {
			logger.Warn("invalid rating: %s", params.Rating)
			return subprocess.NewErrorResponse(msg, "invalid rating: must be 'again', 'hard', 'good', or 'easy'"), nil
		}

		card, err := service.RateMemoryCard(ctx, params.CardID, params.Rating)
		if err != nil {
			logger.Warn("Failed to rate memory card: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to rate memory card: %v", err)), nil
		}
		resp := map[string]any{"success": true, "memory_card": card}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}
