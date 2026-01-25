package main

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// RegisterCWEViewHandlers registers the view handlers onto a subprocess instance.
func RegisterCWEViewHandlers(sp *subprocess.Subprocess, store *cwe.LocalCWEStore, logger *common.Logger) {
	sp.RegisterHandler("RPCSaveCWEView", createSaveCWEViewHandler(store, logger))
	sp.RegisterHandler("RPCGetCWEViewByID", createGetCWEViewHandler(store, logger))
	sp.RegisterHandler("RPCListCWEViews", createListCWEViewsHandler(store, logger))
	sp.RegisterHandler("RPCDeleteCWEView", createDeleteCWEViewHandler(store, logger))
}

// createSaveCWEViewHandler handles RPCSaveCWEView
func createSaveCWEViewHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var view cwe.CWEView
		if err := subprocess.UnmarshalPayload(msg, &view); err != nil {
			logger.Warn("Failed to parse SaveCWEView request: %v", err)
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "invalid request", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		if view.ID == "" {
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "ID is required", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		if err := store.SaveView(ctx, &view); err != nil {
			logger.Warn("SaveView error: %v", err)
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "failed to save view", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: []byte(`{"success":true}`)}, nil
	}
}

// createGetCWEViewHandler handles RPCGetCWEViewByID
func createGetCWEViewHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "invalid request", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "id required", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		v, err := store.GetViewByID(ctx, req.ID)
		if err != nil {
			logger.Warn("GetViewByID error: %v", err)
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "view not found", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		b, _ := subprocess.MarshalFast(v)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: b}, nil
	}
}

// createListCWEViewsHandler handles RPCListCWEViews
func createListCWEViewsHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "invalid request", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		items, total, err := store.ListViewsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn("ListViews error: %v", err)
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "failed to list views", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		resp := map[string]interface{}{"views": items, "offset": req.Offset, "limit": req.Limit, "total": total}
		b, _ := subprocess.MarshalFast(resp)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: b}, nil
	}
}

// createDeleteCWEViewHandler handles RPCDeleteCWEView
func createDeleteCWEViewHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "invalid request", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "id required", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		if err := store.DeleteView(ctx, req.ID); err != nil {
			logger.Error("DeleteView error: %v", err)
			return &subprocess.Message{Type: subprocess.MessageTypeError, ID: msg.ID, Error: "failed to delete view", CorrelationID: msg.CorrelationID, Target: msg.Source}, nil
		}
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: []byte(`{"success":true}`)}, nil
	}
}
