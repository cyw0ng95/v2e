// Package main provides SSG import job RPC handlers for the meta service.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/ssg/job"
)

// createSSGStartImportJobHandler creates a handler for RPCSSGStartImportJob
func createSSGStartImportJobHandler(importer *job.Importer, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGStartImportJob request")
		var req struct {
			RunID string `json:"run_id"`
		}
		// Generate run ID if not provided
		req.RunID = generateRunID()
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse RPCSSGStartImportJob request: %v", errResp.Error)
				// Keep generated run ID on parse error
			}
		}

		if err := importer.StartImport(ctx, req.RunID); err != nil {
			logger.Warn("Failed to start import job: %v", err)
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		logger.Info("SSG import job started: %s", req.RunID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
			"run_id":  req.RunID,
		})
	}
}

// createSSGStopImportJobHandler creates a handler for RPCSSGStopImportJob
func createSSGStopImportJobHandler(importer *job.Importer, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGStopImportJob request")

		if err := importer.StopImport(ctx); err != nil {
			logger.Warn("Failed to stop import job: %v", err)
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		logger.Info("SSG import job stopped")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createSSGPauseImportJobHandler creates a handler for RPCSSGPauseImportJob
func createSSGPauseImportJobHandler(importer *job.Importer, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGPauseImportJob request")

		if err := importer.PauseImport(ctx); err != nil {
			logger.Warn("Failed to pause import job: %v", err)
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		logger.Info("SSG import job paused")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createSSGResumeImportJobHandler creates a handler for RPCSSGResumeImportJob
func createSSGResumeImportJobHandler(importer *job.Importer, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGResumeImportJob request")
		var req struct {
			RunID string `json:"run_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGResumeImportJob request: %v", errResp.Error)
			return errResp, nil
		}

		if err := importer.ResumeImport(ctx, req.RunID); err != nil {
			logger.Warn("Failed to resume import job: %v", err)
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		logger.Info("SSG import job resumed: %s", req.RunID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createSSGGetImportStatusHandler creates a handler for RPCSSGGetImportStatus
func createSSGGetImportStatusHandler(importer *job.Importer, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetImportStatus request")

		status, err := importer.GetStatus(ctx)
		if err != nil {
			logger.Warn("Failed to get import status: %v", err)
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		return subprocess.NewSuccessResponse(msg, status)
	}
}

// RegisterSSGJobHandlers registers all SSG job RPC handlers
func RegisterSSGJobHandlers(sp *subprocess.Subprocess, importer *job.Importer, logger *common.Logger) {
	sp.RegisterHandler("RPCSSGStartImportJob", createSSGStartImportJobHandler(importer, logger))
	logger.Info("RPC handler registered: RPCSSGStartImportJob")

	sp.RegisterHandler("RPCSSGStopImportJob", createSSGStopImportJobHandler(importer, logger))
	logger.Info("RPC handler registered: RPCSSGStopImportJob")

	sp.RegisterHandler("RPCSSGPauseImportJob", createSSGPauseImportJobHandler(importer, logger))
	logger.Info("RPC handler registered: RPCSSGPauseImportJob")

	sp.RegisterHandler("RPCSSGResumeImportJob", createSSGResumeImportJobHandler(importer, logger))
	logger.Info("RPC handler registered: RPCSSGResumeImportJob")

	sp.RegisterHandler("RPCSSGGetImportStatus", createSSGGetImportStatusHandler(importer, logger))
	logger.Info("RPC handler registered: RPCSSGGetImportStatus")
}

// generateRunID generates a unique run ID
func generateRunID() string {
	// Simple timestamp-based ID for now
	// In production, use UUID or similar
	return fmt.Sprintf("ssg-import-%d", time.Now().UnixNano())
}
