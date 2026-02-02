package subprocess

import (
	"strings"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name              string
		msg               *Message
		errMsg            string
		wantType          proc.MessageType
		wantID            string
		wantError         string
		wantCorrelationID string
		wantTarget        string
		wantSource        string
	}{
		{
			name: "basic error response",
			msg: &Message{
				ID:            "req-123",
				CorrelationID: "corr-abc",
				Source:        "broker",
				Target:        "local",
			},
			errMsg:            "something went wrong",
			wantType:          proc.MessageTypeError,
			wantID:            "req-123",
			wantError:         "something went wrong",
			wantCorrelationID: "corr-abc",
			wantTarget:        "broker",
			wantSource:        "",
		},
		{
			name: "error response with empty correlation ID",
			msg: &Message{
				ID:     "req-456",
				Source: "access",
				Target: "remote",
			},
			errMsg:            "connection failed",
			wantType:          proc.MessageTypeError,
			wantID:            "req-456",
			wantError:         "connection failed",
			wantCorrelationID: "",
			wantTarget:        "access",
			wantSource:        "",
		},
		{
			name: "error response with special characters",
			msg: &Message{
				ID:            "req-789",
				CorrelationID: "corr-xyz",
				Source:        "meta",
				Target:        "local",
			},
			errMsg:            "error: invalid JSON at position 42",
			wantType:          proc.MessageTypeError,
			wantID:            "req-789",
			wantError:         "error: invalid JSON at position 42",
			wantCorrelationID: "corr-xyz",
			wantTarget:        "meta",
			wantSource:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponse(tt.msg, tt.errMsg)

			if got.Type != tt.wantType {
				t.Errorf("NewErrorResponse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.ID != tt.wantID {
				t.Errorf("NewErrorResponse() ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.Error != tt.wantError {
				t.Errorf("NewErrorResponse() Error = %v, want %v", got.Error, tt.wantError)
			}
			if got.CorrelationID != tt.wantCorrelationID {
				t.Errorf("NewErrorResponse() CorrelationID = %v, want %v", got.CorrelationID, tt.wantCorrelationID)
			}
			if got.Target != tt.wantTarget {
				t.Errorf("NewErrorResponse() Target = %v, want %v", got.Target, tt.wantTarget)
			}
			if got.Source != tt.wantSource {
				t.Errorf("NewErrorResponse() Source = %v, want %v", got.Source, tt.wantSource)
			}
		})
	}
}

func TestNewErrorResponseWithPrefix(t *testing.T) {
	tests := []struct {
		name              string
		msg               *Message
		prefix            string
		errMsg            string
		wantError         string
	}{
		{
			name: "error with service prefix",
			msg: &Message{
				ID:            "req-123",
				CorrelationID: "corr-abc",
				Source:        "broker",
				Target:        "local",
			},
			prefix:            "local",
			errMsg:            "database connection failed",
			wantError:         "[local] database connection failed",
		},
		{
			name: "error with multi-word prefix",
			msg: &Message{
				ID:            "req-456",
				CorrelationID: "corr-def",
				Source:        "access",
				Target:        "remote",
			},
			prefix:            "remote-fetcher",
			errMsg:            "timeout after 30s",
			wantError:         "[remote-fetcher] timeout after 30s",
		},
		{
			name: "error with empty prefix",
			msg: &Message{
				ID:            "req-789",
				CorrelationID: "corr-ghi",
				Source:        "meta",
				Target:        "local",
			},
			prefix:            "",
			errMsg:            "job not found",
			wantError:         "[] job not found",
		},
		{
			name: "error with prefix containing special chars",
			msg: &Message{
				ID:            "req-101",
				CorrelationID: "corr-jkl",
				Source:        "sysmon",
				Target:        "broker",
			},
			prefix:            "sys.mon",
			errMsg:            "metrics collection failed",
			wantError:         "[sys.mon] metrics collection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponseWithPrefix(tt.msg, tt.prefix, tt.errMsg)

			if got.Type != proc.MessageTypeError {
				t.Errorf("NewErrorResponseWithPrefix() Type = %v, want %v", got.Type, proc.MessageTypeError)
			}
			if got.ID != tt.msg.ID {
				t.Errorf("NewErrorResponseWithPrefix() ID = %v, want %v", got.ID, tt.msg.ID)
			}
			if got.Error != tt.wantError {
				t.Errorf("NewErrorResponseWithPrefix() Error = %v, want %v", got.Error, tt.wantError)
			}
			if got.CorrelationID != tt.msg.CorrelationID {
				t.Errorf("NewErrorResponseWithPrefix() CorrelationID = %v, want %v", got.CorrelationID, tt.msg.CorrelationID)
			}
			if got.Target != tt.msg.Source {
				t.Errorf("NewErrorResponseWithPrefix() Target = %v, want %v", got.Target, tt.msg.Source)
			}
		})
	}
}

func TestNewSuccessResponse(t *testing.T) {
	tests := []struct {
		name              string
		msg               *Message
		result            interface{}
		wantType          proc.MessageType
		wantID            string
		wantCorrelationID string
		wantTarget        string
		wantSource        string
		wantPayload       bool
		expectError       bool
	}{
		{
			name: "success response with string result",
			msg: &Message{
				ID:            "req-123",
				CorrelationID: "corr-abc",
				Source:        "broker",
				Target:        "local",
			},
			result:            "operation completed",
			wantType:          proc.MessageTypeResponse,
			wantID:            "req-123",
			wantCorrelationID: "corr-abc",
			wantTarget:        "broker",
			wantSource:        "local",
			wantPayload:       true,
			expectError:       false,
		},
		{
			name: "success response with map result",
			msg: &Message{
				ID:            "req-456",
				CorrelationID: "corr-def",
				Source:        "access",
				Target:        "remote",
			},
			result: map[string]interface{}{
				"status": "ok",
				"count":  42,
			},
			wantType:          proc.MessageTypeResponse,
			wantID:            "req-456",
			wantCorrelationID: "corr-def",
			wantTarget:        "access",
			wantSource:        "remote",
			wantPayload:       true,
			expectError:       false,
		},
		{
			name: "success response with nil result",
			msg: &Message{
				ID:            "req-789",
				CorrelationID: "corr-ghi",
				Source:        "meta",
				Target:        "local",
			},
			result:            nil,
			wantType:          proc.MessageTypeResponse,
			wantID:            "req-789",
			wantCorrelationID: "corr-ghi",
			wantTarget:        "meta",
			wantSource:        "local",
			wantPayload:       false,
			expectError:       false,
		},
		{
			name: "success response with struct result",
			msg: &Message{
				ID:            "req-101",
				CorrelationID: "corr-jkl",
				Source:        "sysmon",
				Target:        "broker",
			},
			result: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			wantType:          proc.MessageTypeResponse,
			wantID:            "req-101",
			wantCorrelationID: "corr-jkl",
			wantTarget:        "sysmon",
			wantSource:        "broker",
			wantPayload:       true,
			expectError:       false,
		},
		{
			name: "success response with unmarshalable result",
			msg: &Message{
				ID:            "req-999",
				CorrelationID: "corr-999",
				Source:        "broker",
				Target:        "local",
			},
			result:            make(chan int), // channels cannot be marshaled
			wantType:          proc.MessageTypeResponse,
			wantID:            "req-999",
			wantCorrelationID: "corr-999",
			wantTarget:        "broker",
			wantSource:        "local",
			wantPayload:       false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSuccessResponse(tt.msg, tt.result)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewSuccessResponse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewSuccessResponse() unexpected error = %v", err)
			}

			if got.Type != tt.wantType {
				t.Errorf("NewSuccessResponse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.ID != tt.wantID {
				t.Errorf("NewSuccessResponse() ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.CorrelationID != tt.wantCorrelationID {
				t.Errorf("NewSuccessResponse() CorrelationID = %v, want %v", got.CorrelationID, tt.wantCorrelationID)
			}
			if got.Target != tt.wantTarget {
				t.Errorf("NewSuccessResponse() Target = %v, want %v", got.Target, tt.wantTarget)
			}
			if got.Source != tt.wantSource {
				t.Errorf("NewSuccessResponse() Source = %v, want %v", got.Source, tt.wantSource)
			}
			if tt.wantPayload && len(got.Payload) == 0 {
				t.Errorf("NewSuccessResponse() expected payload but got none")
			}
			if !tt.wantPayload && len(got.Payload) != 0 {
				t.Errorf("NewSuccessResponse() expected no payload but got %v", string(got.Payload))
			}
		})
	}
}

func TestIsErrorResponse(t *testing.T) {
	tests := []struct {
		name          string
		msg           *Message
		wantIsError   bool
		wantErrMsg    string
	}{
		{
			name: "error response message",
			msg: &Message{
				Type:  proc.MessageTypeError,
				ID:    "err-123",
				Error: "something failed",
			},
			wantIsError: true,
			wantErrMsg:  "something failed",
		},
		{
			name: "error response with empty error message",
			msg: &Message{
				Type:  proc.MessageTypeError,
				ID:    "err-456",
				Error: "",
			},
			wantIsError: true,
			wantErrMsg:  "",
		},
		{
			name: "success response message",
			msg: &Message{
				Type:    proc.MessageTypeResponse,
				ID:      "resp-789",
				Payload: []byte(`{"status":"ok"}`),
			},
			wantIsError: false,
			wantErrMsg:  "",
		},
		{
			name: "request message",
			msg: &Message{
				Type:    proc.MessageTypeRequest,
				ID:      "req-101",
				Payload: []byte(`{"action":"test"}`),
			},
			wantIsError: false,
			wantErrMsg:  "",
		},
		{
			name: "event message",
			msg: &Message{
				Type:    proc.MessageTypeEvent,
				ID:      "event-202",
				Payload: []byte(`{"event":"started"}`),
			},
			wantIsError: false,
			wantErrMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsError, gotErrMsg := IsErrorResponse(tt.msg)

			if gotIsError != tt.wantIsError {
				t.Errorf("IsErrorResponse() isError = %v, want %v", gotIsError, tt.wantIsError)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Errorf("IsErrorResponse() errMsg = %v, want %v", gotErrMsg, tt.wantErrMsg)
			}
		})
	}
}

func TestIsErrorResponseNilMessage(t *testing.T) {
	// Test with nil message
	isError, errMsg := IsErrorResponse(nil)
	if isError {
		t.Errorf("IsErrorResponse(nil) should return false")
	}
	if errMsg != "" {
		t.Errorf("IsErrorResponse(nil) errMsg = %v, want empty", errMsg)
	}
}

func TestParseRequest(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name          string
		msg           *Message
		req           interface{}
		wantErrResp   bool
		wantErrString string
		wantReqValue  interface{}
	}{
		{
			name: "successful parsing with valid JSON",
			msg: &Message{
				ID:            "req-123",
				CorrelationID: "corr-abc",
				Source:        "broker",
				Target:        "local",
				Payload:       []byte(`{"name":"test","value":42}`),
			},
			req:         &TestRequest{},
			wantErrResp: false,
			wantReqValue: &TestRequest{
				Name:  "test",
				Value: 42,
			},
		},
		{
			name: "error handling with invalid JSON",
			msg: &Message{
				ID:            "req-456",
				CorrelationID: "corr-def",
				Source:        "access",
				Target:        "remote",
				Payload:       []byte(`{"name":"test","value":invalid}`),
			},
			req:           &TestRequest{},
			wantErrResp:   true,
			wantErrString: "failed to parse request:",
		},
		{
			name: "error handling with malformed payload",
			msg: &Message{
				ID:            "req-789",
				CorrelationID: "corr-ghi",
				Source:        "meta",
				Target:        "local",
				Payload:       []byte(`not json at all`),
			},
			req:           &TestRequest{},
			wantErrResp:   true,
			wantErrString: "failed to parse request:",
		},
		{
			name: "nil payload handling",
			msg: &Message{
				ID:            "req-101",
				CorrelationID: "corr-jkl",
				Source:        "sysmon",
				Target:        "broker",
				Payload:       nil,
			},
			req:           &TestRequest{},
			wantErrResp:   true,
			wantErrString: "failed to parse request:",
		},
		{
			name: "successful parsing with empty object",
			msg: &Message{
				ID:            "req-202",
				CorrelationID: "corr-mno",
				Source:        "broker",
				Target:        "local",
				Payload:       []byte(`{}`),
			},
			req:         &TestRequest{},
			wantErrResp: false,
			wantReqValue: &TestRequest{
				Name:  "",
				Value: 0,
			},
		},
		{
			name: "successful parsing with complex nested JSON",
			msg: &Message{
				ID:            "req-303",
				CorrelationID: "corr-pqr",
				Source:        "access",
				Target:        "remote",
				Payload:       []byte(`{"name":"complex","value":999}`),
			},
			req:         &TestRequest{},
			wantErrResp: false,
			wantReqValue: &TestRequest{
				Name:  "complex",
				Value: 999,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := ParseRequest(tt.msg, tt.req)

			if tt.wantErrResp {
				if errResp == nil {
					t.Errorf("ParseRequest() expected error response but got nil")
					return
				}
				if errResp.Type != MessageTypeError {
					t.Errorf("ParseRequest() error response Type = %v, want %v", errResp.Type, MessageTypeError)
				}
				if errResp.ID != tt.msg.ID {
					t.Errorf("ParseRequest() error response ID = %v, want %v", errResp.ID, tt.msg.ID)
				}
				if errResp.CorrelationID != tt.msg.CorrelationID {
					t.Errorf("ParseRequest() error response CorrelationID = %v, want %v", errResp.CorrelationID, tt.msg.CorrelationID)
				}
				if errResp.Target != tt.msg.Source {
					t.Errorf("ParseRequest() error response Target = %v, want %v", errResp.Target, tt.msg.Source)
				}
				// Check that error string contains expected prefix
				if tt.wantErrString != "" && !strings.Contains(errResp.Error, tt.wantErrString) {
					t.Errorf("ParseRequest() error = %v, want to contain %v", errResp.Error, tt.wantErrString)
				}
			} else {
				if errResp != nil {
					t.Errorf("ParseRequest() unexpected error response = %v", errResp.Error)
					return
				}
				// Verify the request struct was populated correctly
				if tt.wantReqValue != nil {
					gotReq, ok := tt.req.(*TestRequest)
					if !ok {
						t.Errorf("ParseRequest() req type assertion failed")
						return
					}
					wantReq := tt.wantReqValue.(*TestRequest)
					if gotReq.Name != wantReq.Name {
						t.Errorf("ParseRequest() req.Name = %v, want %v", gotReq.Name, wantReq.Name)
					}
					if gotReq.Value != wantReq.Value {
						t.Errorf("ParseRequest() req.Value = %v, want %v", gotReq.Value, wantReq.Value)
					}
				}
			}
		})
	}
}

func TestRequireField(t *testing.T) {
	tests := []struct {
		name          string
		msg           *Message
		value         string
		fieldName     string
		wantErrResp   bool
		wantErrString string
	}{
		{
			name: "field with value passes validation",
			msg: &Message{
				ID:            "req-123",
				CorrelationID: "corr-abc",
				Source:        "broker",
				Target:        "local",
			},
			value:       "test-value",
			fieldName:   "CVE_ID",
			wantErrResp: false,
		},
		{
			name: "empty field value returns error",
			msg: &Message{
				ID:            "req-456",
				CorrelationID: "corr-def",
				Source:        "access",
				Target:        "remote",
			},
			value:         "",
			fieldName:     "CVE_ID",
			wantErrResp:   true,
			wantErrString: "CVE_ID is required",
		},
		{
			name: "multi-word field name in error",
			msg: &Message{
				ID:            "req-789",
				CorrelationID: "corr-ghi",
				Source:        "meta",
				Target:        "local",
			},
			value:         "",
			fieldName:     "Job ID",
			wantErrResp:   true,
			wantErrString: "Job ID is required",
		},
		{
			name: "field with only whitespace passes validation",
			msg: &Message{
				ID:            "req-101",
				CorrelationID: "corr-jkl",
				Source:        "sysmon",
				Target:        "broker",
			},
			value:       "   ",
			fieldName:   "CVE_ID",
			wantErrResp: false, // Whitespace is not empty, so it passes
		},
		{
			name: "field with special characters",
			msg: &Message{
				ID:            "req-202",
				CorrelationID: "corr-mno",
				Source:        "broker",
				Target:        "local",
			},
			value:       "CVE-2024-1234",
			fieldName:   "cve_id",
			wantErrResp: false,
		},
		{
			name: "field name with underscores",
			msg: &Message{
				ID:            "req-303",
				CorrelationID: "corr-pqr",
				Source:        "access",
				Target:        "remote",
			},
			value:         "",
			fieldName:     "job_id",
			wantErrResp:   true,
			wantErrString: "job_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := RequireField(tt.msg, tt.value, tt.fieldName)

			if tt.wantErrResp {
				if errResp == nil {
					t.Errorf("RequireField() expected error response but got nil")
					return
				}
				if errResp.Type != MessageTypeError {
					t.Errorf("RequireField() error response Type = %v, want %v", errResp.Type, MessageTypeError)
				}
				if errResp.Error != tt.wantErrString {
					t.Errorf("RequireField() error = %v, want %v", errResp.Error, tt.wantErrString)
				}
				// Verify the error response has correct metadata
				if errResp.ID != tt.msg.ID {
					t.Errorf("RequireField() error ID = %v, want %v", errResp.ID, tt.msg.ID)
				}
				if errResp.Target != tt.msg.Source {
					t.Errorf("RequireField() error Target = %v, want %v", errResp.Target, tt.msg.Source)
				}
			} else {
				if errResp != nil {
					t.Errorf("RequireField() unexpected error response = %v", errResp.Error)
				}
			}
		})
	}
}

