package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/ssg"
)

// createDeploySSGPackageHandler creates a handler for RPCDeploySSGPackage
func createDeploySSGPackageHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			PackageData []byte `json:"package_data"`
		}

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if len(req.PackageData) == 0 {
			return subprocess.NewErrorResponse(msg, "package_data is required"), nil
		}

		logger.Info("Deploying SSG package (%d bytes)", len(req.PackageData))

		// Deploy package
		if err := store.DeployPackage(req.PackageData); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to deploy package: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"success": true,
			"message": "SSG package deployed successfully",
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createListSSGProfilesHandler creates a handler for RPCListSSGProfiles
func createListSSGProfilesHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}

		// Set defaults
		req.Offset = 0
		req.Limit = 10
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		logger.Info("Listing SSG profiles (offset=%d, limit=%d)", req.Offset, req.Limit)

		profiles, total, err := store.ListProfiles(req.Offset, req.Limit)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list profiles: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"profiles": profiles,
			"total":    total,
			"offset":   req.Offset,
			"limit":    req.Limit,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createGetSSGProfileHandler creates a handler for RPCGetSSGProfile
func createGetSSGProfileHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ProfileID string `json:"profile_id"`
		}

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.ProfileID == "" {
			return subprocess.NewErrorResponse(msg, "profile_id is required"), nil
		}

		logger.Info("Getting SSG profile: %s", req.ProfileID)

		profile, err := store.GetProfile(req.ProfileID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("profile not found: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"profile": profile,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createListSSGRulesHandler creates a handler for RPCListSSGRules
func createListSSGRulesHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Offset   int               `json:"offset"`
			Limit    int               `json:"limit"`
			Severity string            `json:"severity"`
			Profile  string            `json:"profile"`
		}

		// Set defaults
		req.Offset = 0
		req.Limit = 10
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		logger.Info("Listing SSG rules (offset=%d, limit=%d, severity=%s)", req.Offset, req.Limit, req.Severity)

		// Build filters
		filters := make(map[string]string)
		if req.Severity != "" {
			filters["severity"] = req.Severity
		}
		if req.Profile != "" {
			filters["profile"] = req.Profile
		}

		rules, total, err := store.ListRules(req.Offset, req.Limit, filters)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list rules: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"rules":  rules,
			"total":  total,
			"offset": req.Offset,
			"limit":  req.Limit,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createGetSSGRuleHandler creates a handler for RPCGetSSGRule
func createGetSSGRuleHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			RuleID string `json:"rule_id"`
		}

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.RuleID == "" {
			return subprocess.NewErrorResponse(msg, "rule_id is required"), nil
		}

		logger.Info("Getting SSG rule: %s", req.RuleID)

		rule, err := store.GetRule(req.RuleID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("rule not found: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"rule": rule,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createSearchSSGContentHandler creates a handler for RPCSearchSSGContent
func createSearchSSGContentHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Query  string `json:"query"`
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
		}

		// Set defaults
		req.Offset = 0
		req.Limit = 10
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		if req.Query == "" {
			return subprocess.NewErrorResponse(msg, "query is required"), nil
		}

		logger.Info("Searching SSG content: %s (offset=%d, limit=%d)", req.Query, req.Offset, req.Limit)

		results, total, err := store.SearchContent(req.Query, req.Offset, req.Limit)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("search failed: %v", err)), nil
		}

		respPayload := map[string]interface{}{
			"results": results,
			"total":   total,
			"offset":  req.Offset,
			"limit":   req.Limit,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}

// createGetSSGMetadataHandler creates a handler for RPCGetSSGMetadata
func createGetSSGMetadataHandler(store *ssg.LocalSSGStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("Getting SSG metadata")

		metadata := store.GetMetadata()

		respPayload := map[string]interface{}{
			"metadata": metadata,
		}

		return subprocess.NewSuccessResponse(msg, respPayload)
	}
}
