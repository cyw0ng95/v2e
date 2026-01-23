package main

import (
	"context"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createImportCAPECsHandler creates a handler for RPCImportCAPECs
func createImportCAPECsHandler(store *capec.LocalCAPECStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCImportCAPECs handler invoked")
		var req struct {
			Path  string `json:"path"`
			XSD   string `json:"xsd"`
			Force bool   `json:"force,omitempty"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("RPCImportCAPECs received path: %s xsd: %s", req.Path, req.XSD)
		if req.Path == "" || req.XSD == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "path and xsd are required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := store.ImportFromXML(req.Path, req.XSD, req.Force); err != nil {
			logger.Error("Failed to import CAPEC from XML: %v (path: %s xsd: %s)", err, req.Path, req.XSD)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Error("CAPEC import file stat error: %v (path: %s)", statErr, req.Path)
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import CAPECs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       []byte(`{"success":true}`),
		}, nil
	}
}

// xmlInnerToPlain strips all XML/HTML tags and returns plain text suitable for
// direct rendering. It also removes xmlns declarations and unescapes entities.
func xmlInnerToPlain(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	// remove xmlns declarations
	xmlnsRe := regexp.MustCompile(`\s+xmlns(:[a-z0-9_]+)?="[^"]*"`)
	s = xmlnsRe.ReplaceAllString(s, "")
	// strip all tags
	tagRe := regexp.MustCompile(`(?s)<[^>]+>`) // dotall to match across lines
	s = tagRe.ReplaceAllString(s, " ")
	// unescape HTML entities
	s = html.UnescapeString(s)
	// collapse whitespace
	wsRe := regexp.MustCompile(`\s+`)
	s = wsRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// createForceImportCAPECsHandler creates a handler for RPCForceImportCAPECs
func createForceImportCAPECsHandler(store *capec.LocalCAPECStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCForceImportCAPECs handler invoked")
		var req struct {
			Path string `json:"path"`
			XSD  string `json:"xsd"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.Path == "" || req.XSD == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "path and xsd are required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := store.ImportFromXML(req.Path, req.XSD, true); err != nil {
			logger.Error("Failed to import CAPEC from XML (force): %v (path: %s xsd: %s)", err, req.Path, req.XSD)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import CAPECs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       []byte(`{"success":true}`),
		}, nil
	}
}

// createGetCAPECByIDHandler creates a handler for RPCGetCAPECByID
func createGetCAPECByIDHandler(store *capec.LocalCAPECStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CAPECID string `json:"capec_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CAPECID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "capec_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("GetCAPECByID request: capec_id=%s", req.CAPECID)
		item, err := store.GetByID(ctx, req.CAPECID)
		if err != nil {
			logger.Error("Failed to get CAPEC: %v (capec_id=%s)", err, req.CAPECID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "CAPEC not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		// Fetch related data (weaknesses, examples, mitigations, references)
		var weaknesses []string
		if rw, err := store.GetRelatedWeaknesses(ctx, item.CAPECID); err == nil {
			for _, w := range rw {
				weaknesses = append(weaknesses, w.CWEID)
			}
		} else {
			logger.Debug("No related weaknesses for CAPEC %d: %v", item.CAPECID, err)
		}

		var examples []string
		if ex, err := store.GetExamples(ctx, item.CAPECID); err == nil {
			for _, e := range ex {
				examples = append(examples, xmlInnerToPlain(e.ExampleText))
			}
		}

		var mitigations []string
		if ms, err := store.GetMitigations(ctx, item.CAPECID); err == nil {
			for _, m := range ms {
				mitigations = append(mitigations, xmlInnerToPlain(m.MitigationText))
			}
		}

		var references []map[string]string
		if refs, err := store.GetReferences(ctx, item.CAPECID); err == nil {
			for _, r := range refs {
				references = append(references, map[string]string{"reference": r.ExternalReference, "url": r.URL})
			}
		}

		// Build a client-friendly payload: use string ID "CAPEC-<n>" and simple keys
		description := xmlInnerToPlain(item.Description)
		payload := map[string]interface{}{
			"id":               fmt.Sprintf("CAPEC-%d", item.CAPECID),
			"name":             item.Name,
			"summary":          xmlInnerToPlain(item.Summary),
			"description":      description,
			"status":           item.Status,
			"likelihood":       item.Likelihood,
			"typical_severity": item.TypicalSeverity,
			"weaknesses":       weaknesses,
			"examples":         examples,
			"mitigations":      mitigations,
			"references":       references,
		}
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal CAPEC: %v (capec_id=%s)", err, req.CAPECID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal CAPEC",
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

// createGetCAPECCatalogMetaHandler creates a handler for RPCGetCAPECCatalogMeta
func createGetCAPECCatalogMetaHandler(store *capec.LocalCAPECStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// No payload expected
		meta, err := store.GetCatalogMeta(ctx)
		if err != nil {
			logger.Debug("No CAPEC catalog metadata: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "no catalog metadata",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		resp := map[string]interface{}{
			"version":     meta.Version,
			"source":      meta.Source,
			"imported_at": meta.ImportedAtUTC,
		}
		data, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal catalog meta: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal meta",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       data,
		}, nil
	}
}

// createListCAPECsHandler creates a handler for RPCListCAPECs
func createListCAPECsHandler(store *capec.LocalCAPECStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListCAPECs handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Error("Failed to parse request: %v", err)
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info("Listing CAPECs with offset=%d, limit=%d", req.Offset, req.Limit)
		items, total, err := store.ListCAPECsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Error("Failed to list CAPECs: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list CAPECs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			// attempt to load related entries; ignore errors to keep listing robust
			var weaknesses []string
			if rw, err := store.GetRelatedWeaknesses(ctx, it.CAPECID); err == nil {
				for _, w := range rw {
					weaknesses = append(weaknesses, w.CWEID)
				}
			}

			var examples []string
			if ex, err := store.GetExamples(ctx, it.CAPECID); err == nil {
				for _, e := range ex {
					examples = append(examples, xmlInnerToPlain(e.ExampleText))
				}
			}

			var mitigations []string
			if ms, err := store.GetMitigations(ctx, it.CAPECID); err == nil {
				for _, m := range ms {
					mitigations = append(mitigations, xmlInnerToPlain(m.MitigationText))
				}
			}

			var references []map[string]string
			if refs, err := store.GetReferences(ctx, it.CAPECID); err == nil {
				for _, r := range refs {
					references = append(references, map[string]string{"reference": r.ExternalReference, "url": r.URL})
				}
			}

			mapped = append(mapped, map[string]interface{}{
				"id":               fmt.Sprintf("CAPEC-%d", it.CAPECID),
				"name":             it.Name,
				"summary":          xmlInnerToPlain(it.Summary),
				"description":      xmlInnerToPlain(it.Description),
				"status":           it.Status,
				"likelihood":       it.Likelihood,
				"typical_severity": it.TypicalSeverity,
				"weaknesses":       weaknesses,
				"examples":         examples,
				"mitigations":      mitigations,
				"references":       references,
			})
		}

		resp := map[string]interface{}{
			"capecs": mapped,
			"offset": req.Offset,
			"limit":  req.Limit,
			"total":  total,
		}
		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal CAPECs: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal CAPECs",
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
