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
		if errResp := subprocess.ParseRequest(msg, &view); errResp != nil {
			logger.Warn("Failed to parse SaveCWEView request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, view.ID, "id"); errResp != nil {
			return errResp, nil
		}
		if err := store.SaveView(ctx, &view); err != nil {
			logger.Warn("SaveView error: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to save view"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}

// createGetCWEViewHandler handles RPCGetCWEViewByID
func createGetCWEViewHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		v, err := store.GetViewByID(ctx, req.ID)
		if err != nil {
			logger.Warn("GetViewByID error: %v", err)
			return subprocess.NewErrorResponse(msg, "view not found"), nil
		}
		return subprocess.NewSuccessResponse(msg, v)
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
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse request: %v", errResp.Error)
				return errResp, nil
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
			return subprocess.NewErrorResponse(msg, "failed to list views"), nil
		}
		resp := map[string]interface{}{"views": items, "offset": req.Offset, "limit": req.Limit, "total": total}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// createDeleteCWEViewHandler handles RPCDeleteCWEView
func createDeleteCWEViewHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		if err := store.DeleteView(ctx, req.ID); err != nil {
			logger.Error("DeleteView error: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to delete view"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}
