package main

import (
	"context"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// capecStore captures the subset of CAPEC store behaviors needed by handlers.
type capecStore interface {
	ImportFromXML(xmlPath, xsdPath string, force bool) error
	GetCatalogMeta(ctx context.Context) (*capec.CAPECCatalogMeta, error)
	ListCAPECsPaginated(ctx context.Context, offset, limit int) ([]capec.CAPECItemModel, int64, error)
	GetByID(ctx context.Context, capecID string) (*capec.CAPECItemModel, error)
	GetRelatedWeaknesses(ctx context.Context, capecID int) ([]capec.CAPECRelatedWeaknessModel, error)
	GetExamples(ctx context.Context, capecID int) ([]capec.CAPECExampleModel, error)
	GetMitigations(ctx context.Context, capecID int) ([]capec.CAPECMitigationModel, error)
	GetReferences(ctx context.Context, capecID int) ([]capec.CAPECReferenceModel, error)
}

// createImportCAPECsHandler creates a handler for RPCImportCAPECs
func createImportCAPECsHandler(store capecStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCImportCAPECs handler invoked")
		var req struct {
			Path  string `json:"path"`
			XSD   string `json:"xsd"`
			Force bool   `json:"force,omitempty"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse request: %v", err)
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
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
			logger.Warn("Failed to import CAPEC from XML: %v (path: %s xsd: %s)", err, req.Path, req.XSD)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Warn("CAPEC import file stat error: %v (path: %s)", statErr, req.Path)
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
func createForceImportCAPECsHandler(store capecStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCForceImportCAPECs handler invoked")
		var req struct {
			Path string `json:"path"`
			XSD  string `json:"xsd"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse request: %v", err)
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
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
			logger.Warn("Failed to import CAPEC from XML (force): %v (path: %s xsd: %s)", err, req.Path, req.XSD)
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
func createGetCAPECByIDHandler(store capecStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CAPECID string `json:"capec_id"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse request: %v", err)
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
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
			logger.Warn("Failed to get CAPEC: %v (capec_id=%s)", err, req.CAPECID)
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
		jsonData, err := subprocess.MarshalFast(payload)
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
func createGetCAPECCatalogMetaHandler(store capecStore, logger *common.Logger) subprocess.Handler {
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
		data, err := subprocess.MarshalFast(resp)
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
func createListCAPECsHandler(store capecStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing ListCAPECs request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse ListCAPECs request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
				logger.Debug("Processing ListCAPECs request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         fmt.Sprintf("failed to parse request: %v", err),
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
		logger.Info("Processing ListCAPECs request - Message ID: %s, Correlation ID: %s, Offset: %d, Limit: %d", msg.ID, msg.CorrelationID, req.Offset, req.Limit)
		items, total, err := store.ListCAPECsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn("Failed to list CAPECs from store - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListCAPECs request failed - Message ID: %s, Error details: %v", msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to list CAPECs: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			logger.Debug("Mapping CAPEC item - Message ID: %s, CAPEC ID: %d", msg.ID, it.CAPECID)
			// attempt to load related entries; ignore errors to keep listing robust
			var weaknesses []string
			if rw, err := store.GetRelatedWeaknesses(ctx, it.CAPECID); err == nil {
				for _, w := range rw {
					weaknesses = append(weaknesses, w.CWEID)
				}
			} else {
				logger.Debug("No related weaknesses found for CAPEC %d - Message ID: %s, Error: %v", it.CAPECID, msg.ID, err)
			}

			var examples []string
			if ex, err := store.GetExamples(ctx, it.CAPECID); err == nil {
				for _, e := range ex {
					examples = append(examples, xmlInnerToPlain(e.ExampleText))
				}
			} else {
				logger.Debug("No examples found for CAPEC %d - Message ID: %s, Error: %v", it.CAPECID, msg.ID, err)
			}

			var mitigations []string
			if ms, err := store.GetMitigations(ctx, it.CAPECID); err == nil {
				for _, m := range ms {
					mitigations = append(mitigations, xmlInnerToPlain(m.MitigationText))
				}
			} else {
				logger.Debug("No mitigations found for CAPEC %d - Message ID: %s, Error: %v", it.CAPECID, msg.ID, err)
			}

			var references []map[string]string
			if refs, err := store.GetReferences(ctx, it.CAPECID); err == nil {
				for _, r := range refs {
					references = append(references, map[string]string{"reference": r.ExternalReference, "url": r.URL})
				}
			} else {
				logger.Debug("No references found for CAPEC %d - Message ID: %s, Error: %v", it.CAPECID, msg.ID, err)
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
		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Error("Failed to marshal ListCAPECs response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         fmt.Sprintf("failed to marshal CAPECs: %v", err),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("Successfully processed ListCAPECs request - Message ID: %s, Correlation ID: %s, Returned: %d, Total: %d", msg.ID, msg.CorrelationID, len(items), total)
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}
