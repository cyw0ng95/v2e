package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// handlerResult represents the result of a handler operation with optional data.
type handlerResult struct {
	Success bool
	Data    interface{}
	Message string
}

// newSuccessResult creates a success result with data.
func newSuccessResult(data interface{}) *handlerResult {
	return &handlerResult{
		Success: true,
		Data:    data,
	}
}

// newErrorResult creates an error result with a message.
func newErrorResult(format string, args ...interface{}) *handlerResult {
	return &handlerResult{
		Success: false,
		Message: fmt.Sprintf(format, args...),
	}
}

// parseRequestWithDefaults parses a request with default values for optional fields.
// If payload is nil or empty, the default values are used.
func parseRequestWithDefaults(msg *subprocess.Message, req interface{}, defaults func()) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		if defaults != nil {
			defaults()
		}
		return nil
	}
	if errResp := subprocess.ParseRequest(msg, req); errResp != nil {
		return fmt.Errorf("parse error: %s", errResp.Error)
	}
	return nil
}

// sendSuccessResponse sends a successful response with the given data.
func sendSuccessResponse(msg *subprocess.Message, data interface{}) (*subprocess.Message, error) {
	result := map[string]interface{}{
		"success": true,
	}
	if data != nil {
		switch v := data.(type) {
		case map[string]interface{}:
			for key, val := range v {
				result[key] = val
			}
		default:
			result["data"] = data
		}
	}
	resp, err := subprocess.NewSuccessResponse(msg, result)
	if err != nil {
		return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return resp, nil
}

// sendErrorResponse sends an error response with the given message.
func sendErrorResponse(msg *subprocess.Message, format string, args ...interface{}) (*subprocess.Message, error) {
	message := fmt.Sprintf(format, args...)
	return subprocess.NewErrorResponse(msg, message), nil
}

// executeWithValidation executes a handler function with common error handling and logging.
// It wraps the common pattern of: parse request -> validate -> execute operation -> send response.
func executeWithValidation(
	ctx context.Context,
	msg *subprocess.Message,
	logger *common.Logger,
	operationName string,
	parse func() error,
	validate func() error,
	execute func() (interface{}, error),
) (*subprocess.Message, error) {
	logger.Debug("Processing %s request - Message ID: %s, Correlation ID: %s", operationName, msg.ID, msg.CorrelationID)

	// Parse request
	if err := parse(); err != nil {
		logger.Warn("Failed to parse %s request: %v", operationName, err)
		logger.Debug("Processing %s request failed due to malformed payload: %s", operationName, string(msg.Payload))
		return subprocess.NewErrorResponse(msg, err.Error()), nil
	}

	// Validate
	if err := validate(); err != nil {
		logger.Warn("Validation failed for %s: %v", operationName, err)
		return subprocess.NewErrorResponse(msg, err.Error()), nil
	}

	// Execute operation
	result, err := execute()
	if err != nil {
		logger.Warn("Failed to execute %s: %v", operationName, err)
		return sendErrorResponse(msg, "failed to %s: %v", operationName, err)
	}

	logger.Info("Successfully completed %s - Message ID: %s", operationName, msg.ID)
	return sendSuccessResponse(msg, result)
}

// validateIDField validates a required ID field and returns a formatted error if invalid.
func validateIDField(id string, fieldName string) error {
	if id == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// validateRequiredField validates a required field with a custom error message.
func validateRequiredField(value interface{}, fieldName string) error {
	if value == nil || value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}
