// SSG handlers for local service
package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	ssgparser "github.com/cyw0ng95/v2e/pkg/ssg/parser"
	"github.com/cyw0ng95/v2e/pkg/ssg/local"
)

// createSSGGetGuideHandler creates a handler for RPCSSGGetGuide
func createSSGGetGuideHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetGuide request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetGuide request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			logger.Warn("id is required for RPCSSGGetGuide")
			return errResp, nil
		}
		guide, err := store.GetGuide(req.ID)
		if err != nil {
			logger.Warn("Failed to get guide: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get guide: %v", err)), nil
		}
		logger.Info("Got guide %s", req.ID)
		return subprocess.NewSuccessResponse(msg, guide)
	}
}

// createSSGListGuidesHandler creates a handler for RPCSSGListGuides
func createSSGListGuidesHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGListGuides request")
		var req struct {
			Product   string `json:"product"`
			ProfileID string `json:"profile_id"`
		}
		// Default to empty filters
		req.Product = ""
		req.ProfileID = ""
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse RPCSSGListGuides request: %v", errResp.Error)
				return errResp, nil
			}
		}
		logger.Debug("Listing SSG guides with filters: product=%s profile_id=%s", req.Product, req.ProfileID)
		guides, err := store.ListGuides(req.Product, req.ProfileID)
		if err != nil {
			logger.Error("Failed to list SSG guides: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list guides: %v", err)), nil
		}
		logger.Info("Listed %d SSG guides (product=%s profile_id=%s)", len(guides), req.Product, req.ProfileID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"guides": guides,
			"count":  len(guides),
		})
	}
}

// createSSGGetTreeHandler creates a handler for RPCSSGGetTree
func createSSGGetTreeHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetTree request")
		var req struct {
			GuideID string `json:"guide_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetTree request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.GuideID, "guide_id"); errResp != nil {
			logger.Warn("guide_id is required for RPCSSGGetTree")
			return errResp, nil
		}
		tree, err := store.GetTree(req.GuideID)
		if err != nil {
			logger.Warn("Failed to get tree: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get tree: %v", err)), nil
		}
		logger.Info("Got tree for guide %s", req.GuideID)
		return subprocess.NewSuccessResponse(msg, tree)
	}
}

// createSSGGetTreeNodeHandler creates a handler for RPCSSGGetTreeNode
// Returns the tree structure as TreeNode pointers for frontend consumption
func createSSGGetTreeNodeHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetTreeNode request")
		var req struct {
			GuideID string `json:"guide_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetTreeNode request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.GuideID, "guide_id"); errResp != nil {
			logger.Warn("guide_id is required for RPCSSGGetTreeNode")
			return errResp, nil
		}
		nodes, err := store.BuildTreeNodes(req.GuideID)
		if err != nil {
			logger.Warn("Failed to build tree nodes: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to build tree nodes: %v", err)), nil
		}
		logger.Info("Built %d tree nodes for guide %s", len(nodes), req.GuideID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"nodes": nodes,
			"count": len(nodes),
		})
	}
}

// createSSGGetGroupHandler creates a handler for RPCSSGGetGroup
func createSSGGetGroupHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetGroup request")
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetGroup request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			logger.Warn("id is required for RPCSSGGetGroup")
			return errResp, nil
		}
		group, err := store.GetGroup(req.ID)
		if err != nil {
			logger.Warn("Failed to get group: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get group: %v", err)), nil
		}
		logger.Info("Got group %s", req.ID)
		return subprocess.NewSuccessResponse(msg, group)
	}
}

// createSSGGetChildGroupsHandler creates a handler for RPCSSGGetChildGroups
func createSSGGetChildGroupsHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetChildGroups request")
		var req struct {
			ParentID string `json:"parent_id"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse RPCSSGGetChildGroups request: %v", errResp.Error)
				return errResp, nil
			}
		}
		groups, err := store.GetChildGroups(req.ParentID)
		if err != nil {
			logger.Warn("Failed to get child groups: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get child groups: %v", err)), nil
		}
		logger.Info("Got %d child groups for parent %s", len(groups), req.ParentID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"groups": groups,
			"count":  len(groups),
		})
	}
}

// createSSGGetRuleHandler creates a handler for RPCSSGGetRule
func createSSGGetRuleHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetRule request")
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetRule request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			logger.Warn("id is required for RPCSSGGetRule")
			return errResp, nil
		}
		rule, err := store.GetRule(req.ID)
		if err != nil {
			logger.Warn("Failed to get rule: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get rule: %v", err)), nil
		}
		logger.Info("Got rule %s", req.ID)
		return subprocess.NewSuccessResponse(msg, rule)
	}
}

// createSSGListRulesHandler creates a handler for RPCSSGListRules
func createSSGListRulesHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGListRules request")
		var req struct {
			GroupID  string `json:"group_id"`
			Severity string `json:"severity"`
			Offset   int    `json:"offset"`
			Limit    int    `json:"limit"`
		}
		// Set defaults
		req.Offset = 0
		req.Limit = 100
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse RPCSSGListRules request: %v", errResp.Error)
				return errResp, nil
			}
		}
		rules, total, err := store.ListRules(req.GroupID, req.Severity, req.Offset, req.Limit)
		if err != nil {
			logger.Warn("Failed to list rules: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list rules: %v", err)), nil
		}
		logger.Info("Listed %d rules (total: %d)", len(rules), total)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"rules": rules,
			"total": total,
		})
	}
}

// createSSGGetChildRulesHandler creates a handler for RPCSSGGetChildRules
func createSSGGetChildRulesHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGGetChildRules request")
		var req struct {
			GroupID string `json:"group_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGGetChildRules request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.GroupID, "group_id"); errResp != nil {
			logger.Warn("group_id is required for RPCSSGGetChildRules")
			return errResp, nil
		}
		rules, err := store.GetChildRules(req.GroupID)
		if err != nil {
			logger.Warn("Failed to get child rules: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get child rules: %v", err)), nil
		}
		logger.Info("Got %d child rules for group %s", len(rules), req.GroupID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"rules": rules,
			"count": len(rules),
		})
	}
}

// createSSGImportGuideHandler creates a handler for RPCSSGImportGuide
// Parses an HTML guide file and imports it into the database
func createSSGImportGuideHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGImportGuide request")
		var req struct {
			Path string `json:"path"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGImportGuide request: %v", errResp.Error)
			return errResp, nil
		}
		if errMsg := subprocess.RequireField(msg, req.Path, "path"); errMsg != nil {
			logger.Warn("path is required for RPCSSGImportGuide")
			return errMsg, nil
		}

		logger.Info("Starting SSG guide import from path: %s", req.Path)

		// Parse the guide file
		guide, groups, rules, err := ssgparser.ParseGuideFile(req.Path)
		if err != nil {
			logger.Error("Failed to parse guide file %s: %v", req.Path, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to parse guide: %v", err)), nil
		}
		logger.Info("Parsed guide %s: product=%s profile=%s groups=%d rules=%d", guide.ID, guide.Product, guide.ProfileID, len(groups), len(rules))

		// Save guide to database
		if err := store.SaveGuide(guide); err != nil {
			logger.Error("Failed to save guide %s to database: %v", guide.ID, err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to save guide: %v", err)), nil
		}
		logger.Debug("Saved guide %s to database", guide.ID)

		// Save all groups
		for i := range groups {
			if err := store.SaveGroup(&groups[i]); err != nil {
				logger.Error("Failed to save group %s for guide %s: %v", groups[i].ID, guide.ID, err)
				return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to save group: %v", err)), nil
			}
		}
		logger.Debug("Saved %d groups for guide %s", len(groups), guide.ID)

		// Save all rules with references
		for i := range rules {
			if err := store.SaveRule(&rules[i]); err != nil {
				logger.Error("Failed to save rule %s for guide %s: %v", rules[i].ID, guide.ID, err)
				return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to save rule: %v", err)), nil
			}
		}
		logger.Debug("Saved %d rules for guide %s", len(rules), guide.ID)

		logger.Info("Successfully imported guide %s: product=%s profile=%s groups=%d rules=%d", guide.ID, guide.Product, guide.ProfileID, len(groups), len(rules))
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":    true,
			"guide_id":   guide.ID,
			"group_count": len(groups),
			"rule_count":  len(rules),
		})
	}
}

// createSSGDeleteGuideHandler creates a handler for RPCSSGDeleteGuide
func createSSGDeleteGuideHandler(store *local.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing RPCSSGDeleteGuide request")
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse RPCSSGDeleteGuide request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			logger.Warn("id is required for RPCSSGDeleteGuide")
			return errResp, nil
		}
		if err := store.DeleteGuide(req.ID); err != nil {
			logger.Warn("Failed to delete guide: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete guide: %v", err)), nil
		}
		logger.Info("Deleted guide %s", req.ID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
			"id":      req.ID,
		})
	}
}

// RegisterSSGHandlers registers all SSG local RPC handlers
func RegisterSSGHandlers(sp *subprocess.Subprocess, store *local.Store, logger *common.Logger) {
	sp.RegisterHandler("RPCSSGImportGuide", createSSGImportGuideHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGImportGuide")

	sp.RegisterHandler("RPCSSGGetGuide", createSSGGetGuideHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetGuide")

	sp.RegisterHandler("RPCSSGListGuides", createSSGListGuidesHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGListGuides")

	sp.RegisterHandler("RPCSSGGetTree", createSSGGetTreeHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetTree")

	sp.RegisterHandler("RPCSSGGetTreeNode", createSSGGetTreeNodeHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetTreeNode")

	sp.RegisterHandler("RPCSSGGetGroup", createSSGGetGroupHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetGroup")

	sp.RegisterHandler("RPCSSGGetChildGroups", createSSGGetChildGroupsHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetChildGroups")

	sp.RegisterHandler("RPCSSGGetRule", createSSGGetRuleHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetRule")

	sp.RegisterHandler("RPCSSGListRules", createSSGListRulesHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGListRules")

	sp.RegisterHandler("RPCSSGGetChildRules", createSSGGetChildRulesHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGGetChildRules")

	sp.RegisterHandler("RPCSSGDeleteGuide", createSSGDeleteGuideHandler(store, logger))
	logger.Info("RPC handler registered: RPCSSGDeleteGuide")
}
