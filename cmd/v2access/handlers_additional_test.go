package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/gin-gonic/gin"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockSubprocessForRPC helps simulate broker replies by calling back into RPCClient
// responseWriter captures writes from Subprocess.SendMessage and invokes the RPCClient
type responseWriter struct {
	client      *RPCClient
	respType    subprocess.MessageType
	payload     interface{}
	errMessage  string
	buf         bytes.Buffer
	lastRequest *subprocess.Message
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		data := w.buf.Bytes()
		idx := bytes.IndexByte(data, '\n')
		if idx == -1 {
			break
		}
		line := data[:idx]
		w.buf.Next(idx + 1)
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var msg subprocess.Message
		if err := json.Unmarshal(line, &msg); err != nil {
			return len(p), err
		}
		w.lastRequest = &msg
		resp := &subprocess.Message{Type: w.respType, ID: msg.ID, CorrelationID: msg.CorrelationID}
		if w.respType == subprocess.MessageTypeError {
			resp.Error = w.errMessage
		} else if w.payload != nil {
			switch v := w.payload.(type) {
			case []byte:
				resp.Payload = v
			case json.RawMessage:
				resp.Payload = []byte(v)
			default:
				if data, err := subprocess.MarshalFast(w.payload); err == nil {
					resp.Payload = data
				}
			}
		}
		w.client.handleResponse(context.Background(), resp)
	}
	return len(p), nil
}

func newRPCClientWithResponse(respType subprocess.MessageType, payload interface{}, errMessage string) (*RPCClient, *responseWriter) {
	sp := subprocess.New("test-client")
	logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)
	client := NewRPCClientWithSubprocess(sp, logger, time.Second)
	rw := &responseWriter{client: client, respType: respType, payload: payload, errMessage: errMessage}
	sp.SetOutput(rw)
	return client, rw
}

// errorWriter returns an error on Write to simulate send failures
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) { return 0, errors.New("write error") }

// Test missing method produces 400
func TestRPCHandler_MissingMethod(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRPCHandler_MissingMethod", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		sp := subprocess.New("test-client")
		logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)
		rpcClient := NewRPCClientWithSubprocess(sp, logger, 1*time.Second)
		r := gin.Default()
		rg := r.Group("/restful")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

}

// Test RPC forward when SendMessage fails (simulate transport error)
func TestRPCHandler_RPCSendError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRPCHandler_RPCSendError", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")

		// Use a real Subprocess with an output writer that returns an error to simulate send failures
		sp := subprocess.New("err-client")
		logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)
		rpcClient := NewRPCClientWithSubprocess(sp, logger, 1*time.Second)
		sp.SetOutput(&errorWriter{})
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc", strings.NewReader(`{"method":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if int(resp["retcode"].(float64)) == 0 {
			t.Fatalf("expected non-zero retcode when send fails")
		}
	})

}

// Test RPC returns MessageTypeError -> handler maps to retcode 500 and message
func TestRPCHandler_MessageTypeError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_MessageTypeError", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")

		rpcClient, _ := newRPCClientWithResponse(subprocess.MessageTypeError, nil, "boom")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc", strings.NewReader(`{"method":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if int(resp["retcode"].(float64)) == 0 {
			t.Fatalf("expected non-zero retcode for error message")
		}
		if resp["message"].(string) != "boom" {
			t.Fatalf("expected message 'boom', got %v", resp["message"])
		}
	})

}

// Test invalid payload from RPC produces parse error handling
func TestRPCHandler_InvalidPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_InvalidPayload", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")

		// Create a client whose response writer will set raw []byte payload directly
		rpcClient, _ := newRPCClientWithResponse(subprocess.MessageTypeResponse, []byte("not-json"), "")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc", strings.NewReader(`{"method":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if int(resp["retcode"].(float64)) == 0 {
			t.Fatalf("expected non-zero retcode for parse error")
		}
	})

}

// Test default target is broker when empty target provided
func TestRPCHandler_DefaultTargetBroker(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_DefaultTargetBroker", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")
		rpcClient, rw := newRPCClientWithResponse(subprocess.MessageTypeResponse, map[string]bool{"ok": true}, "")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc", strings.NewReader(`{"method":"x","target":""}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if rw.lastRequest == nil {
			t.Fatalf("no request captured by response writer")
		}
		if rw.lastRequest.Target != "broker" {
			t.Fatalf("expected default target 'broker', got '%s'", rw.lastRequest.Target)
		}
	})

}

// Test path-based RPC endpoint: /restful/rpc/cve/list
func TestPathRPC_CVEList(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPathRPC_CVEList", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")
		rpcClient, _ := newRPCClientWithResponse(subprocess.MessageTypeResponse, map[string]interface{}{"cves": []string{"CVE-2021-44228"}, "total": 1}, "")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc/cve/list", strings.NewReader(`{"offset":0,"limit":10}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if int(resp["retcode"].(float64)) != 0 {
			t.Fatalf("expected retcode 0, got %v", resp["retcode"])
		}
	})

}

// Test path-based RPC endpoint: /restful/rpc/session/status
func TestPathRPC_SessionStatus(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPathRPC_SessionStatus", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()
		rg := r.Group("/restful")
		rpcClient, _ := newRPCClientWithResponse(subprocess.MessageTypeResponse, map[string]bool{"hasSession": false}, "")
		registerHandlers(rg, rpcClient)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/restful/rpc/session/status", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if int(resp["retcode"].(float64)) != 0 {
			t.Fatalf("expected retcode 0, got %v", resp["retcode"])
		}
	})

}
