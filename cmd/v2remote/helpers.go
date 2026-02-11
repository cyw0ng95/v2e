package main

import (
	"errors"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	StartIndex     int `json:"start_index"`
	ResultsPerPage int `json:"results_per_page"`
}

// parsePaginationRequest parses pagination request with default values
func parsePaginationRequest(msg *subprocess.Message, defaultStartIndex, defaultResultsPerPage int) (PaginationRequest, error) {
	var req PaginationRequest
	req.StartIndex = defaultStartIndex
	req.ResultsPerPage = defaultResultsPerPage

	if msg.Payload != nil {
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return req, fmt.Errorf("failed to parse request: %v", err)
		}
	}

	// Clamp pagination values to valid ranges
	if req.StartIndex < 0 {
		req.StartIndex = 0
	}
	if req.ResultsPerPage < 1 {
		req.ResultsPerPage = defaultResultsPerPage
	}

	return req, nil
}

// validatePagination validates pagination parameters
func validatePagination(req PaginationRequest) error {
	validator := subprocess.NewValidator()
	validator.ValidateIntPositive(req.StartIndex, "start_index")
	validator.ValidateIntRange(req.ResultsPerPage, 1, 2000, "results_per_page")
	if validator.HasErrors() {
		return validator.Error()
	}
	return nil
}

// checkRateLimitError checks if the error is a rate limit error and returns an appropriate response
func checkRateLimitError(msg *subprocess.Message, err error) (*subprocess.Message, error) {
	if err == remote.ErrRateLimited || errors.Is(err, remote.ErrRateLimited) {
		return subprocess.NewErrorResponse(msg, ErrMsgNVDRateLimited), nil
	}
	return nil, err
}
