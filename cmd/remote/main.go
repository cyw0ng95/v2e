/*
Package main implements the remote RPC service.

Refer to service.md for the RPC API Specification and details about the CVE Remote Service.

Package main provides the implementation of the remote CVE service using RPC.
*/
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "remote"
	}

	// Use a bootstrap logger for initial messages before the full logging system is ready
	bootstrapLogger := common.NewLogger(os.Stderr, "", common.InfoLevel)

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		bootstrapLogger.Error(LogMsgFailedToSetupLogging, err)
		os.Exit(1)
	}

	// Get API key from environment (optional)
	apiKey := os.Getenv("NVD_API_KEY")

	// Create CVE fetcher
	fetcher := remote.NewFetcher(apiKey)

	// Create subprocess instance
	sp := subprocess.New(processID)

	// Register RPC handlers
	sp.RegisterHandler("RPCGetCVEByID", createGetCVEByIDHandler(fetcher))
	sp.RegisterHandler("RPCGetCVECnt", createGetCVECntHandler(fetcher))
	sp.RegisterHandler("RPCFetchCVEs", createFetchCVEsHandler(fetcher))
	sp.RegisterHandler("RPCFetchViews", createFetchViewsHandler())

	logger.Info(LogMsgServiceStarted)

	// Run with default lifecycle management
	subprocess.RunWithDefaults(sp, logger)
}

// createFetchViewsHandler creates a handler for RPCFetchViews which downloads
// the GitHub archive and extracts JSON files under json_repo/V.
func createFetchViewsHandler() subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			StartIndex     int `json:"start_index"`
			ResultsPerPage int `json:"results_per_page"`
		}

		// Set sensible defaults
		req.StartIndex = 0
		req.ResultsPerPage = 100
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		// Download GitHub zip archive
		zipURL := os.Getenv("VIEW_FETCH_URL")
		if zipURL == "" {
			zipURL = "https://github.com/CWE-CAPEC/REST-API-wg/archive/refs/heads/main.zip"
		}
		resp, err := http.Get(zipURL)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedDownloadArchive, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgUnexpectedHTTPStatus, resp.Status),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedReadBody, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedOpenZip, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		var allViews []cwe.CWEView
		for _, f := range zr.File {
			// look for files under json_repo/V and with .json suffix
			// zip entries from GitHub will have a top-level folder like REST-API-wg-main/
			if !strings.Contains(f.Name, "json_repo/"+"V/") {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(f.Name), ".json") {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			var view cwe.CWEView
			// Try to unmarshal with fast unmarshal
			if err := subprocess.UnmarshalFast(data, &view); err != nil {
				// try again as fallback (preserve original behavior)
				_ = subprocess.UnmarshalFast(data, &view)
			}
			// If ID is empty, try to derive filename as ID
			if view.ID == "" {
				view.ID = strings.TrimSuffix(filepath.Base(f.Name), filepath.Ext(f.Name))
			}

			// Skip entries that are just header rows like "view"
			if strings.ToLower(strings.TrimSpace(view.ID)) == "view" {
				continue
			}
			allViews = append(allViews, view)
		}

		// Pagination
		start := req.StartIndex
		if start < 0 {
			start = 0
		}
		pageSize := req.ResultsPerPage
		if pageSize <= 0 {
			pageSize = 100
		}

		if start > len(allViews) {
			start = len(allViews)
		}
		end := start + pageSize
		if end > len(allViews) {
			end = len(allViews)
		}

		respPayload := map[string]interface{}{
			"views": allViews[start:end],
		}

		jsonData, err := subprocess.MarshalFast(respPayload)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedMarshalResp, err),
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

// createGetCVEByIDHandler creates a handler for RPCGetCVEByID
func createGetCVEByIDHandler(fetcher *remote.Fetcher) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedParseReq, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		if req.CVEID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         ErrMsgCVEIDRequired,
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		// Fetch CVE from NVD
		response, err := fetcher.FetchCVEByID(req.CVEID)
		if err != nil {
			// Check if this is a rate limit error
			if err == remote.ErrRateLimited {
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         ErrMsgNVDRateLimited,
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedFetchCVE, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Marshal the response
		jsonData, err := subprocess.MarshalFast(response)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedMarshalResp, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetCVECntHandler creates a handler for RPCGetCVECnt
func createGetCVECntHandler(fetcher *remote.Fetcher) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload (optional parameters)
		var req struct {
			StartIndex     int `json:"start_index"`
			ResultsPerPage int `json:"results_per_page"`
		}

		// Set defaults if not provided
		req.StartIndex = 0
		req.ResultsPerPage = 1 // Minimum to just get the count

		// Try to parse payload, but it's optional
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		// Fetch CVEs to get the total count
		response, err := fetcher.FetchCVEs(req.StartIndex, req.ResultsPerPage)
		if err != nil {
			// Check if this is a rate limit error
			if err == remote.ErrRateLimited {
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         ErrMsgNVDRateLimited,
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedFetchCount, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		// Create response with count
		result := map[string]interface{}{
			"total_results": response.TotalResults,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Marshal the result
		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedMarshalResult, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createFetchCVEsHandler creates a handler for RPCFetchCVEs
func createFetchCVEsHandler(fetcher *remote.Fetcher) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			StartIndex     int `json:"start_index"`
			ResultsPerPage int `json:"results_per_page"`
		}

		// Set defaults
		req.StartIndex = 0
		req.ResultsPerPage = 100

		// Try to parse payload
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		// Fetch CVEs from NVD
		response, err := fetcher.FetchCVEs(req.StartIndex, req.ResultsPerPage)
		if err != nil {
			// Check if this is a rate limit error
			if err == remote.ErrRateLimited {
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         ErrMsgNVDRateLimited,
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedFetchCVEs, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Marshal the response
		jsonData, err := subprocess.MarshalFast(response)
		if err != nil {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf(ErrMsgFailedMarshalResp, err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}
