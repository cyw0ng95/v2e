package main

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createSaveCVEByIDHandler creates a handler for RPCSaveCVEByID
func createSaveCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing SaveCVEByID request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			CVE cve.CVEItem `json:"cve"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse SaveCVEByID request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing SaveCVEByID request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to parse request: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CVE.ID == "" {
			logger.Warn("CVE ID is required in SaveCVEByID request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
			logger.Debug("Processing SaveCVEByID request failed: CVE ID missing in payload - Message ID: %s", msg.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cve.id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := db.SaveCVE(&req.CVE); err != nil {
			logger.Warn("Failed to save CVE to database - Message ID: %s, Correlation ID: %s, CVE ID: %s, Error: %v", msg.ID, msg.CorrelationID, req.CVE.ID, err)
			logger.Debug("Processing SaveCVEByID request failed for CVE ID %s - Message ID: %s, Error details: %v", req.CVE.ID, msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to save CVE: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Successfully saved CVE to local database - Message ID: %s, Correlation ID: %s, CVE ID: %s", msg.ID, msg.CorrelationID, req.CVE.ID)
		logger.Debug("Processing SaveCVEByID request completed successfully - Message ID: %s, CVE ID: %s", msg.ID, req.CVE.ID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVE.ID,
		}
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal SaveCVEByID response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}

// createIsCVEStoredByIDHandler creates a handler for RPCIsCVEStoredByID
func createIsCVEStoredByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing IsCVEStoredByID request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to parse request: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CVEID == "" {
			logger.Error("cve_id is required")
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cve_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		_, err := db.GetCVE(req.CVEID)
		stored := err == nil
		logger.Debug("Processing IsCVEStoredByID request completed successfully for CVE ID %s, stored: %v", req.CVEID, stored)
		result := map[string]interface{}{
			"cve_id": req.CVEID,
			"stored": stored,
		}
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal result: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}

// createGetCVEByIDHandler creates a handler for RPCGetCVEByID
func createGetCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing GetCVEByID request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse GetCVEByID request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing GetCVEByID request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to parse request: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CVEID == "" {
			logger.Warn("CVE ID is required in GetCVEByID request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
			logger.Debug("Processing GetCVEByID request failed: CVE ID missing in payload - Message ID: %s", msg.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cve_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		cveItem, err := db.GetCVE(req.CVEID)
		if err != nil {
			logger.Warn("Failed to get CVE from database - Message ID: %s, Correlation ID: %s, CVE ID: %s, Error: %v", msg.ID, msg.CorrelationID, req.CVEID, err)
			logger.Debug("Processing GetCVEByID request failed for CVE ID %s - Message ID: %s, Error details: %v", req.CVEID, msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("CVE not found: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Retrieved CVE from local database - Message ID: %s, Correlation ID: %s, CVE ID: %s", msg.ID, msg.CorrelationID, req.CVEID)
		logger.Debug("Processing GetCVEByID request completed successfully - Message ID: %s, CVE ID: %s", msg.ID, req.CVEID)
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(cveItem)
		if err != nil {
			logger.Error("Failed to marshal GetCVEByID response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}

// createDeleteCVEByIDHandler creates a handler for RPCDeleteCVEByID
func createDeleteCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing DeleteCVEByID request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to parse request: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CVEID == "" {
			logger.Warn("cve_id is required")
			logger.Debug("Processing DeleteCVEByID request failed: CVE ID missing in payload")
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cve_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := db.DeleteCVE(req.CVEID); err != nil {
			logger.Warn("Failed to delete CVE from database: %v", err)
			logger.Debug("Processing DeleteCVEByID request failed for CVE ID %s: %v", req.CVEID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to delete CVE: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Deleted CVE %s from local database", req.CVEID)
		logger.Debug("Processing DeleteCVEByID request completed successfully for CVE ID %s", req.CVEID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVEID,
		}
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal result: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}

// createListCVEsHandler creates a handler for RPCListCVEs
func createListCVEsHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing ListCVEs request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		req.Offset = 0
		req.Limit = 10
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse ListCVEs request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
				logger.Debug("Processing ListCVEs request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         fmt.Sprintf("failed to parse request: %v", err),
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
		}
		logger.Info("Processing ListCVEs request - Message ID: %s, Correlation ID: %s, Offset: %d, Limit: %d", msg.ID, msg.CorrelationID, req.Offset, req.Limit)
		cves, err := db.ListCVEs(req.Offset, req.Limit)
		if err != nil {
			logger.Warn("Failed to list CVEs from database - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListCVEs request failed - Message ID: %s, Error details: %v", msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to list CVEs: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		total, err := db.Count()
		if err != nil {
			logger.Warn("Failed to get CVE count from database - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListCVEs request failed to get count - Message ID: %s, Error details: %v", msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to get CVE count: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Successfully listed CVEs - Message ID: %s, Correlation ID: %s, Returned: %d, Total: %d, Offset: %d, Limit: %d", msg.ID, msg.CorrelationID, len(cves), total, req.Offset, req.Limit)
		logger.Debug("Processing ListCVEs request completed successfully - Message ID: %s, Returned %d CVEs, Total %d", msg.ID, len(cves), total)
		result := map[string]interface{}{
			"cves":  cves,
			"total": total,
		}
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal ListCVEs response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}

// createCountCVEsHandler creates a handler for RPCCountCVEs
func createCountCVEsHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing CountCVEs request")
		count, err := db.Count()
		if err != nil {
			logger.Warn("Failed to count CVEs in database: %v", err)
			logger.Debug("Processing CountCVEs request failed: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to count CVEs: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("CVE count: %d", count)
		logger.Debug("Processing CountCVEs request completed successfully: count %d", count)
		result := map[string]interface{}{
			"count": count,
		}
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal result: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal result: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData
		return respMsg, nil
	}
}
