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
		var req struct {
			CVE cve.CVEItem `json:"cve"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to parse request: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CVE.ID == "" {
			logger.Error("cve.id is required")
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cve.id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := db.SaveCVE(&req.CVE); err != nil {
			logger.Error("Failed to save CVE: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to save CVE: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Saved CVE %s to local database", req.CVE.ID)
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

// createIsCVEStoredByIDHandler creates a handler for RPCIsCVEStoredByID
func createIsCVEStoredByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
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
		logger.Debug("CVE %s stored status: %v", req.CVEID, stored)
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
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
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
		cveItem, err := db.GetCVE(req.CVEID)
		if err != nil {
			logger.Error("Failed to get CVE from database: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("CVE not found: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Retrieved CVE %s from local database", req.CVEID)
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		jsonData, err := sonic.Marshal(cveItem)
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

// createDeleteCVEByIDHandler creates a handler for RPCDeleteCVEByID
func createDeleteCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
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
		if err := db.DeleteCVE(req.CVEID); err != nil {
			logger.Error("Failed to delete CVE from database: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to delete CVE: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Deleted CVE %s from local database", req.CVEID)
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
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		req.Offset = 0
		req.Limit = 10
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}
		logger.Debug("Listing CVEs with offset=%d, limit=%d", req.Offset, req.Limit)
		cves, err := db.ListCVEs(req.Offset, req.Limit)
		if err != nil {
			logger.Error("Failed to list CVEs from database: %v", err)
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
			logger.Error("Failed to get CVE count from database: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to get CVE count: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Listed %d CVEs (total: %d, offset: %d, limit: %d)", len(cves), total, req.Offset, req.Limit)
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

// createCountCVEsHandler creates a handler for RPCCountCVEs
func createCountCVEsHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Counting CVEs in database")
		count, err := db.Count()
		if err != nil {
			logger.Error("Failed to count CVEs in database: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to count CVEs: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("CVE count: %d", count)
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
