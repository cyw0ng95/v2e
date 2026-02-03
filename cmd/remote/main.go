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
	"crypto/sha512"
	"encoding/hex"
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
	"github.com/cyw0ng95/v2e/pkg/ssg"
)

func main() {
	// Use common startup utility to standardize initialization
	configStruct := subprocess.StandardStartupConfig{
		DefaultProcessID: "remote",
		LogPrefix:        "[REMOTE] ",
	}
	sp, logger := subprocess.StandardStartup(configStruct)

	// Get API key from environment (optional)
	apiKey := os.Getenv("NVD_API_KEY")
	if apiKey != "" {
		logger.Info(LogMsgAPIKeyDetected)
	} else {
		logger.Info(LogMsgAPIKeyNotSet)
	}

	// Create CVE fetcher
	fetcher := remote.NewFetcher(apiKey)
	logger.Info(LogMsgFetcherCreated, apiKey != "")

	// Register RPC handlers
	logger.Info("Registering RPC handlers...")
	sp.RegisterHandler("RPCGetCVEByID", createGetCVEByIDHandler(fetcher))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCVEByID")
	sp.RegisterHandler("RPCGetCVECnt", createGetCVECntHandler(fetcher))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCVECnt")
	sp.RegisterHandler("RPCFetchCVEs", createFetchCVEsHandler(fetcher))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCFetchCVEs")
	sp.RegisterHandler("RPCFetchViews", createFetchViewsHandler())
	logger.Info(LogMsgRPCHandlerRegistered, "RPCFetchViews")
	sp.RegisterHandler("RPCFetchSSGPackage", createFetchSSGPackageHandler(logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCFetchSSGPackage")

	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	// Run with default lifecycle management
	logger.Info("Starting subprocess with default lifecycle management")
	subprocess.RunWithDefaults(sp, logger)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
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
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedDownloadArchive, err)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgUnexpectedHTTPStatus, resp.Status)), nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedReadBody, err)), nil
		}

		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedOpenZip, err)), nil
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

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createFetchSSGPackageHandler creates a handler for RPCFetchSSGPackage which downloads
// the SSG package from GitHub and verifies its SHA512 checksum.
func createFetchSSGPackageHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Version string `json:"version"`
		}

		// Set default version
		req.Version = ssg.DefaultSSGVersion
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		logger.Info("Fetching SSG package version %s", req.Version)

		// Build URLs
		packageURL := fmt.Sprintf(ssg.SSGReleaseURLTemplate, req.Version, req.Version)
		sha512URL := fmt.Sprintf(ssg.SSGSHA512URLTemplate, req.Version, req.Version)

		// Download SHA512 checksum file first
		logger.Info("Downloading SHA512 checksum from %s", sha512URL)
		sha512Resp, err := http.Get(sha512URL)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to download SHA512: %v", err)), nil
		}
		defer sha512Resp.Body.Close()

		if sha512Resp.StatusCode != http.StatusOK {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("SHA512 download failed with status: %s", sha512Resp.Status)), nil
		}

		sha512Data, err := io.ReadAll(sha512Resp.Body)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to read SHA512: %v", err)), nil
		}

		// Parse SHA512 (format: "checksum  filename")
		expectedChecksum := strings.Fields(string(sha512Data))[0]
		logger.Info("Expected SHA512 checksum: %s", expectedChecksum)

		// Download package
		logger.Info("Downloading SSG package from %s", packageURL)
		packageResp, err := http.Get(packageURL)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to download package: %v", err)), nil
		}
		defer packageResp.Body.Close()

		if packageResp.StatusCode != http.StatusOK {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("package download failed with status: %s", packageResp.Status)), nil
		}

		packageData, err := io.ReadAll(packageResp.Body)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to read package: %v", err)), nil
		}

		// Calculate SHA512 checksum
		hash := sha512.New()
		hash.Write(packageData)
		actualChecksum := hex.EncodeToString(hash.Sum(nil))

		logger.Info("Calculated SHA512 checksum: %s", actualChecksum)

		// Verify checksum
		verified := actualChecksum == expectedChecksum
		if !verified {
			logger.Warn("SHA512 checksum mismatch! Expected: %s, Got: %s", expectedChecksum, actualChecksum)
			return subprocess.NewErrorResponse(msg, "SHA512 checksum verification failed"), nil
		}

		logger.Info("SHA512 checksum verified successfully")

		respPayload := map[string]interface{}{
			"package_data": packageData,
			"sha512":       expectedChecksum,
			"verified":     verified,
			"version":      req.Version,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createGetCVEByIDHandler creates a handler for RPCGetCVEByID
func createGetCVEByIDHandler(fetcher *remote.Fetcher) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if errMsg := subprocess.ParseRequest(msg, &req); errMsg != nil {
			return errMsg, nil
		}

		// Validate required field
		if errMsg := subprocess.RequireField(msg, req.CVEID, "cve_id"); errMsg != nil {
			return errMsg, nil
		}

		// Fetch CVE from NVD
		response, err := fetcher.FetchCVEByID(req.CVEID)
		if err != nil {
			// Check if this is a rate limit error
			if err == remote.ErrRateLimited {
				return subprocess.NewErrorResponse(msg, ErrMsgNVDRateLimited), nil
			}
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedFetchCVE, err)), nil
		}

		return subprocess.NewSuccessResponse(msg, response)
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
				return subprocess.NewErrorResponse(msg, ErrMsgNVDRateLimited), nil
			}
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedFetchCount, err)), nil
		}

		// Create response with count
		result := map[string]interface{}{
			"total_results": response.TotalResults,
		}

		return subprocess.NewSuccessResponse(msg, result)
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
				return subprocess.NewErrorResponse(msg, ErrMsgNVDRateLimited), nil
			}
			return subprocess.NewErrorResponse(msg, fmt.Sprintf(ErrMsgFailedFetchCVEs, err)), nil
		}

		return subprocess.NewSuccessResponse(msg, response)
	}
}
