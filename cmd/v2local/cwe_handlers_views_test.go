package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestCWEViewHandlers_CreateGetListDelete(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCWEViewHandlers_CreateGetListDelete", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test_cwe_handlers.db")
		store, err := cwe.NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("NewLocalCWEStore error: %v", err)
		}
		// NewLocalCWEStore performs AutoMigrate including view tables; no extra migrate needed here.
		logger := common.NewLogger(os.Stdout, "test", common.DebugLevel)

		// Build Save handler
		saveH := createSaveCWEViewHandler(store, logger)
		v := cwe.CWEView{ID: "V-HT-1", Name: "HandlerTestView"}
		payload, _ := subprocess.MarshalFast(v)
		msg := &subprocess.Message{ID: "1", Payload: payload, Source: "test", CorrelationID: "c1"}
		_, err = saveH(context.Background(), msg)
		if err != nil {
			t.Fatalf("save handler returned error: %v", err)
		}

		// Get
		getH := createGetCWEViewHandler(store, logger)
		getReq := map[string]string{"id": v.ID}
		gp, _ := json.Marshal(getReq)
		gmsg := &subprocess.Message{ID: "2", Payload: gp, Source: "test", CorrelationID: "c2"}
		resp, err := getH(context.Background(), gmsg)
		if err != nil {
			t.Fatalf("get handler error: %v", err)
		}
		var got cwe.CWEView
		if err := subprocess.UnmarshalFast(resp.Payload, &got); err != nil {
			t.Fatalf("failed to unmarshal get response: %v", err)
		}
		if got.ID != v.ID {
			t.Fatalf("get returned wrong id: %s", got.ID)
		}

		// List
		listH := createListCWEViewsHandler(store, logger)
		lreq := map[string]int{"offset": 0, "limit": 10}
		lp, _ := json.Marshal(lreq)
		lmsg := &subprocess.Message{ID: "3", Payload: lp, Source: "test", CorrelationID: "c3"}
		lresp, err := listH(context.Background(), lmsg)
		if err != nil {
			t.Fatalf("list handler error: %v", err)
		}
		var lbody map[string]interface{}
		if err := subprocess.UnmarshalFast(lresp.Payload, &lbody); err != nil {
			t.Fatalf("failed to unmarshal list response: %v", err)
		}
		if lbody["total"].(float64) < 1 {
			t.Fatalf("expected total >= 1, got %v", lbody["total"])
		}

		// Delete
		delH := createDeleteCWEViewHandler(store, logger)
		drq, _ := json.Marshal(map[string]string{"id": v.ID})
		dmsg := &subprocess.Message{ID: "4", Payload: drq, Source: "test", CorrelationID: "c4"}
		_, err = delH(context.Background(), dmsg)
		if err != nil {
			t.Fatalf("delete handler error: %v", err)
		}
	})

}
