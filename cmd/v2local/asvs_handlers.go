package main

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/asvs"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createImportASVSHandler creates a handler for RPCImportASVS
func createImportASVSHandler(store *asvs.LocalASVSStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgImportASVSInvoked)
		logger.Debug("RPCImportASVS handler invoked. msg.ID=%s, correlation_id=%s", msg.ID, msg.CorrelationID)

		var req struct {
			URL string `json:"url"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing ImportASVS request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Debug(LogMsgImportASVSReceivedURL, req.URL)

		if req.URL == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "url is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Info(LogMsgStartingImportASVS, req.URL)
		err := store.ImportFromCSV(ctx, req.URL)
		if err != nil {
			logger.Warn(LogMsgFailedImportASVS, err, req.URL)
			logger.Debug("Processing ImportASVS request failed for URL %s: %v", req.URL, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import ASVS requirements",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Info(LogMsgImportASVSCompleted, req.URL)
		logger.Debug("Processing ImportASVS request completed successfully for URL %s. correlation_id=%s", req.URL, msg.CorrelationID)

		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       []byte(`{"success":true}`),
		}, nil
	}
}

// createListASVSHandler creates a handler for RPCListASVS
func createListASVSHandler(store *asvs.LocalASVSStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgRPCListASVSInvoked, msg.ID)

		var req struct {
			Offset  int    `json:"offset"`
			Limit   int    `json:"limit"`
			Chapter string `json:"chapter"`
			Level   int    `json:"level"`
		}

		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse request: %v", err)
				logger.Debug("Processing ListASVS request failed due to malformed payload: %s", string(msg.Payload))
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
		}

		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}

		logger.Info(LogMsgListASVSParams, req.Offset, req.Limit, req.Chapter, req.Level)

		items, total, err := store.ListASVSPaginated(ctx, req.Offset, req.Limit, req.Chapter, req.Level)
		if err != nil {
			logger.Warn(LogMsgFailedListASVS, err)
			logger.Debug(LogMsgProcessingListASVSFailed, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ASVS requirements",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Debug(LogMsgListASVSCompleted, len(items))

		resp := map[string]interface{}{
			"requirements": items,
			"offset":       req.Offset,
			"limit":        req.Limit,
			"total":        total,
		}

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListASVS, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ASVS requirements",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}

// createGetASVSByIDHandler creates a handler for RPCGetASVSByID
func createGetASVSByIDHandler(store *asvs.LocalASVSStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			RequirementID string `json:"requirement_id"`
		}

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing GetASVSByID request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		if req.RequirementID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "requirement_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Debug(LogMsgGetASVSByIDReq, req.RequirementID)

		item, err := store.GetByID(ctx, req.RequirementID)
		if err != nil {
			logger.Warn(LogMsgFailedGetASVS, err, req.RequirementID)
			logger.Debug(LogMsgProcessingGetASVSFailedErr, req.RequirementID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ASVS requirement not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Debug(LogMsgFoundASVS, item)
		logger.Debug(LogMsgProcessingGetASVSCompleted, req.RequirementID)

		jsonData, err := subprocess.MarshalFast(item)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalASVS, err, req.RequirementID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ASVS requirement",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}
