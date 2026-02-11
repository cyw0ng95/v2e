package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createSaveCVEByIDHandler creates a handler for RPCSaveCVEByID
func createSaveCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgProcessingSaveCVE, msg.ID, msg.CorrelationID)
		var req struct {
			CVE cve.CVEItem `json:"cve"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn(LogMsgFailedParseSaveCVEReq, msg.ID, msg.CorrelationID, errResp.Error)
			logger.Debug(LogMsgProcessingSaveCVEFailed, msg.ID, string(msg.Payload))
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CVE.ID, "cve.id"); errResp != nil {
			logger.Warn(LogMsgCVEIDRequired, msg.ID, msg.CorrelationID)
			logger.Debug(LogMsgProcessingSaveCVEFailedID, msg.ID)
			return errResp, nil
		}
		if err := db.SaveCVE(&req.CVE); err != nil {
			logger.Warn(LogMsgFailedSaveCVE, msg.ID, msg.CorrelationID, req.CVE.ID, err)
			logger.Debug(LogMsgProcessingSaveCVEFailedErr, req.CVE.ID, msg.ID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to save CVE: %v", err)), nil
		}
		logger.Info(LogMsgSuccessSaveCVE, msg.ID, msg.CorrelationID, req.CVE.ID)
		logger.Debug(LogMsgProcessingSaveCVECompleted, msg.ID, req.CVE.ID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVE.ID,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalSaveCVEResp, msg.ID, msg.CorrelationID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createIsCVEStoredByIDHandler creates a handler for RPCIsCVEStoredByID
func createIsCVEStoredByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn(LogMsgFailedParseReq, errResp.Error)
			logger.Debug(LogMsgProcessingIsCVEFailed, string(msg.Payload))
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CVEID, "cve_id"); errResp != nil {
			logger.Warn(LogMsgCVEIDRequiredSimple)
			return errResp, nil
		}
		// Validate CVE ID format for security
		validator := subprocess.NewValidator()
		validator.ValidateCVEID(req.CVEID, "cve_id")
		if validator.HasErrors() {
			return subprocess.NewErrorResponse(msg, validator.Error()), nil
		}
		_, err := db.GetCVE(req.CVEID)
		stored := err == nil
		logger.Debug(LogMsgProcessingIsCVECompleted, req.CVEID, stored)
		result := map[string]interface{}{
			"cve_id": req.CVEID,
			"stored": stored,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalResult, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createGetCVEByIDHandler creates a handler for RPCGetCVEByID
func createGetCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgProcessingGetCVE, msg.ID, msg.CorrelationID)
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn(LogMsgFailedParseGetCVEReq, msg.ID, msg.CorrelationID, errResp.Error)
			logger.Debug(LogMsgProcessingGetCVEFailed, msg.ID, string(msg.Payload))
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CVEID, "cve_id"); errResp != nil {
			logger.Warn(LogMsgCVEIDRequiredGet, msg.ID, msg.CorrelationID)
			logger.Debug(LogMsgProcessingGetCVEFailedID, msg.ID)
			return errResp, nil
		}
		// Validate CVE ID format for security
		validator := subprocess.NewValidator()
		validator.ValidateCVEID(req.CVEID, "cve_id")
		if validator.HasErrors() {
			return subprocess.NewErrorResponse(msg, validator.Error()), nil
		}
		cveItem, err := db.GetCVE(req.CVEID)
		if err != nil {
			logger.Warn(LogMsgFailedGetCVE, msg.ID, msg.CorrelationID, req.CVEID, err)
			logger.Debug(LogMsgProcessingGetCVEFailedErr, req.CVEID, msg.ID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(LogMsgCVEIDNotFound, err)), nil
		}
		logger.Info(LogMsgSuccessGetCVE, msg.ID, msg.CorrelationID, req.CVEID)
		logger.Debug(LogMsgProcessingGetCVECompleted, msg.ID, req.CVEID)
		resp, err := subprocess.NewSuccessResponse(msg, cveItem)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalGetCVEResp, msg.ID, msg.CorrelationID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createDeleteCVEByIDHandler creates a handler for RPCDeleteCVEByID
func createDeleteCVEByIDHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			logger.Debug("Processing DeleteCVEByID request failed due to malformed payload: %s", string(msg.Payload))
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CVEID, "cve_id"); errResp != nil {
			logger.Warn("cve_id is required")
			logger.Debug("Processing DeleteCVEByID request failed: CVE ID missing in payload")
			return errResp, nil
		}
		// Validate CVE ID format for security
		validator := subprocess.NewValidator()
		validator.ValidateCVEID(req.CVEID, "cve_id")
		if validator.HasErrors() {
			return subprocess.NewErrorResponse(msg, validator.Error()), nil
		}
		if err := db.DeleteCVE(req.CVEID); err != nil {
			logger.Warn("Failed to delete CVE from database: %v", err)
			logger.Debug("Processing DeleteCVEByID request failed for CVE ID %s: %v", req.CVEID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete CVE: %v", err)), nil
		}
		logger.Info("Deleted CVE %s from local database", req.CVEID)
		logger.Debug("Processing DeleteCVEByID request completed successfully for CVE ID %s", req.CVEID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVEID,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn("Failed to marshal result: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
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
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse ListCVEs request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, errResp.Error)
				logger.Debug("Processing ListCVEs request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
				return errResp, nil
			}
		}
		// Validate pagination parameters for security
		validator := subprocess.NewValidator()
		validator.ValidateIntPositive(req.Offset, "offset")
		validator.ValidateIntRange(req.Limit, 1, 1000, "limit")
		if validator.HasErrors() {
			return subprocess.NewErrorResponse(msg, validator.Error()), nil
		}
		logger.Info("Processing ListCVEs request - Message ID: %s, Correlation ID: %s, Offset: %d, Limit: %d", msg.ID, msg.CorrelationID, req.Offset, req.Limit)
		cves, err := db.ListCVEs(req.Offset, req.Limit)
		if err != nil {
			logger.Warn("Failed to list CVEs from database - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListCVEs request failed - Message ID: %s, Error details: %v", msg.ID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list CVEs: %v", err)), nil
		}
		total, err := db.Count()
		if err != nil {
			logger.Warn("Failed to get CVE count from database - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListCVEs request failed to get count - Message ID: %s, Error details: %v", msg.ID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get CVE count: %v", err)), nil
		}
		logger.Info("Successfully listed CVEs - Message ID: %s, Correlation ID: %s, Returned: %d, Total: %d, Offset: %d, Limit: %d", msg.ID, msg.CorrelationID, len(cves), total, req.Offset, req.Limit)
		logger.Debug("Processing ListCVEs request completed successfully - Message ID: %s, Returned %d CVEs, Total %d", msg.ID, len(cves), total)
		result := map[string]interface{}{
			"cves":  cves,
			"total": total,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn("Failed to marshal ListCVEs response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
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
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to count CVEs: %v", err)), nil
		}
		logger.Info("CVE count: %d", count)
		logger.Debug("Processing CountCVEs request completed successfully: count %d", count)
		result := map[string]interface{}{
			"count": count,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Error("Failed to marshal result: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createCreateCVEHandler creates a handler for RPCCreateCVE
// Accepts a CVEItem directly (not wrapped) and saves it to the database
func createCreateCVEHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCCreateCVE request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req cve.CVEItem
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCCreateCVE request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "cve.id"); errResp != nil {
			logger.Warn("cve.id is required for RPCCreateCVE")
			return errResp, nil
		}
		if err := db.SaveCVE(&req); err != nil {
			logger.Warn("Failed to create CVE in database: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create CVE: %v", err)), nil
		}
		logger.Info("Created CVE %s in local database", req.ID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.ID,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createUpdateCVEHandler creates a handler for RPCUpdateCVE
// Accepts a CVEItem directly (not wrapped) and updates it in the database
func createUpdateCVEHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCUpdateCVE request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req cve.CVEItem
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCUpdateCVE request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "cve.id"); errResp != nil {
			logger.Warn("cve.id is required for RPCUpdateCVE")
			return errResp, nil
		}
		if err := db.SaveCVE(&req); err != nil {
			logger.Warn("Failed to update CVE in database: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update CVE: %v", err)), nil
		}
		logger.Info("Updated CVE %s in local database", req.ID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.ID,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}

// createDeleteCVEHandler creates a handler for RPCDeleteCVE
// Accepts { cve_id: string } and deletes the CVE from the database
// Note: This is an alias for RPCDeleteCVEByID with the same functionality
func createDeleteCVEHandler(db *local.DB, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCDeleteCVE request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCDeleteCVE request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CVEID, "cve_id"); errResp != nil {
			logger.Warn("cve_id is required for RPCDeleteCVE")
			return errResp, nil
		}
		if err := db.DeleteCVE(req.CVEID); err != nil {
			logger.Warn("Failed to delete CVE from database: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete CVE: %v", err)), nil
		}
		logger.Info("Deleted CVE %s from local database", req.CVEID)
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVEID,
		}
		resp, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return resp, nil
	}
}
