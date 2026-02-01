package main

import (
	"context"
	"encoding/json"

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
	       if err := json.Unmarshal(msg.Payload, &params); err != nil {
		       return nil, err
	       }
	       card, err := service.CreateMemoryCardFull(ctx, params.BookmarkID, params.Front, params.Back, params.MajorClass, params.MinorClass, params.Status, string(params.Content), params.CardType, params.Author, params.IsPrivate, params.Metadata)
	       if err != nil {
		       return nil, err
	       }
	       resp := map[string]any{"success": true, "memory_card": card}
	       payload, _ := json.Marshal(resp)
	       return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
       }
}

// Handler for RPCGetMemoryCard
func getMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
       return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	       var params struct {
		       ID uint `json:"id"`
	       }
	       if err := json.Unmarshal(msg.Payload, &params); err != nil {
		       return nil, err
	       }
	       card, err := service.GetMemoryCardByID(ctx, params.ID)
	       if err != nil {
		       return nil, err
	       }
	       resp := map[string]any{"memory_card": card}
	       payload, _ := json.Marshal(resp)
	       return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
       }
}

// Handler for RPCUpdateMemoryCard
func updateMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
       return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	       var params map[string]any
	       if err := json.Unmarshal(msg.Payload, &params); err != nil {
		       return nil, err
	       }
	       card, err := service.UpdateMemoryCardFields(ctx, params)
	       if err != nil {
		       return nil, err
	       }
	       resp := map[string]any{"success": true, "memory_card": card}
	       payload, _ := json.Marshal(resp)
	       return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
       }
}

// Handler for RPCDeleteMemoryCard
func deleteMemoryCardHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
       return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	       var params struct {
		       ID uint `json:"id"`
	       }
	       if err := json.Unmarshal(msg.Payload, &params); err != nil {
		       return nil, err
	       }
	       err := service.DeleteMemoryCard(ctx, params.ID)
	       if err != nil {
		       return nil, err
	       }
	       resp := map[string]any{"success": true}
	       payload, _ := json.Marshal(resp)
	       return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
       }
}

// Handler for RPCListMemoryCards
func listMemoryCardsHandler(service *notes.MemoryCardService, logger *common.Logger) subprocess.Handler {
       return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	       var params struct {
		       BookmarkID *uint   `json:"bookmark_id"`
		       MajorClass *string `json:"major_class"`
		       MinorClass *string `json:"minor_class"`
		       Status     *string `json:"status"`
		       Author     *string `json:"author"`
		       IsPrivate  *bool   `json:"is_private"`
		       Offset     int     `json:"offset"`
		       Limit      int     `json:"limit"`
	       }
	       if err := json.Unmarshal(msg.Payload, &params); err != nil {
		       return nil, err
	       }
	       cards, total, err := service.ListMemoryCardsFull(ctx, params.BookmarkID, params.MajorClass, params.MinorClass, params.Status, params.Author, params.IsPrivate, params.Offset, params.Limit)
	       if err != nil {
		       return nil, err
	       }
	       resp := map[string]any{"memory_cards": cards, "total": total, "offset": params.Offset, "limit": params.Limit}
	       payload, _ := json.Marshal(resp)
	       return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
       }
}
