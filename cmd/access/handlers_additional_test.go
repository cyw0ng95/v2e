package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	client     *RPCClient
	respType   subprocess.MessageType
	payload    interface{}
	errMessage string
	buf        bytes.Buffer
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

		resp := &subprocess.Message{Type: w.respType, ID: msg.ID, CorrelationID: msg.CorrelationID}
		if w.respType == subprocess.MessageTypeError {
			resp.Error = w.errMessage
		} else if w.payload != nil {
			if data, err := subprocess.MarshalFast(w.payload); err == nil {
				resp.Payload = data
			}
		}

		w.client.handleResponse(context.Background(), resp)
	}
	return len(p), nil
}

func newRPCClientWithResponse(respType subprocess.MessageType, payload interface{}, errMessage string) *RPCClient {
	sp := subprocess.New("test-client")
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
		rpcTimeout:      time.Second,
	}
	sp.SetOutput(&responseWriter{client: client, respType: respType, payload: payload, errMessage: errMessage})
	return client
}

func TestRegisterHandlers_RPCSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newRPCClientWithResponse(subprocess.MessageTypeResponse, map[string]string{"ok": "yes"}, "")
	router := gin.New()
	registerHandlers(router.Group("/restful"), client, 1)

	body := bytes.NewBufferString(`{"method":"RPCPing","params":{"foo":"bar"}}`)
	req := httptest.NewRequest(http.MethodPost, "/restful/rpc", body)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp struct {
		Retcode int                    `json:"retcode"`
		Message string                 `json:"message"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Retcode != 0 || resp.Message != "success" {
		t.Fatalf("Unexpected RPC response metadata: %+v", resp)
	}
	if resp.Payload["ok"] != "yes" {
		t.Fatalf("Payload mismatch: %+v", resp.Payload)
	}
}

func TestRegisterHandlers_RPCErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newRPCClientWithResponse(subprocess.MessageTypeError, nil, "boom")
	router := gin.New()
	registerHandlers(router.Group("/restful"), client, 1)

	body := bytes.NewBufferString(`{"method":"RPCFail","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/restful/rpc", body)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp struct {
		Retcode int         `json:"retcode"`
		Message string      `json:"message"`
		Payload interface{} `json:"payload"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Retcode != 500 {
		t.Fatalf("Expected retcode 500, got %d", resp.Retcode)
	}
	if resp.Message != "boom" {
		t.Fatalf("Unexpected error message: %s", resp.Message)
	}
	if resp.Payload != nil {
		t.Fatalf("Expected nil payload on error, got %v", resp.Payload)
	}
}
