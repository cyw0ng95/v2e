package ssg

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// LocalSSGStore manages local storage of SSG data
type LocalSSGStore struct {
	docPath    string
	metadata   *SSGMetadata
	benchmarks map[string]*SSGBenchmark // key: benchmark ID
	logger     *common.Logger
}

// NewLocalSSGStore creates a new local SSG store
func NewLocalSSGStore(docPath string, logger *common.Logger) (*LocalSSGStore, error) {
	logger.Info(LogMsgSSGDocPathConfigured, docPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create SSG doc path: %w", err)
	}

	store := &LocalSSGStore{
		docPath:    docPath,
		benchmarks: make(map[string]*SSGBenchmark),
		logger:     logger,
	}

	// Load existing benchmarks if any
	if err := store.loadBenchmarks(); err != nil {
		logger.Warn("Failed to load existing benchmarks: %v", err)
	}

	return store, nil
}

// DeployPackage extracts and deploys an SSG package
func (s *LocalSSGStore) DeployPackage(tarGzData []byte) error {
	// Extract package
	if err := ExtractSSGPackage(tarGzData, s.docPath, s.logger); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Reload benchmarks
	if err := s.loadBenchmarks(); err != nil {
		return fmt.Errorf("failed to load benchmarks after deployment: %w", err)
	}

	s.logger.Info(LogMsgSSGPackageDeployed, s.docPath)
	return nil
}

// loadBenchmarks loads all SSG benchmarks from the doc path
func (s *LocalSSGStore) loadBenchmarks() error {
	// Find all SSG XML files
	files, err := FindSSGXMLFiles(s.docPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		s.logger.Info("No SSG XML files found in %s", s.docPath)
		return nil
	}

	// Parse each file
	for _, file := range files {
		benchmark, err := ParseSSGBenchmark(file, s.logger)
		if err != nil {
			s.logger.Warn("Failed to parse %s: %v", file, err)
			continue
		}

		s.benchmarks[benchmark.ID] = benchmark
	}

	// Update metadata
	s.updateMetadata()

	return nil
}

// updateMetadata updates the store metadata based on loaded benchmarks
func (s *LocalSSGStore) updateMetadata() {
	totalProfiles := 0
	totalRules := 0
	var version string
	var benchmarkID string
	var title string
	var description string

	for id, benchmark := range s.benchmarks {
		if benchmarkID == "" {
			benchmarkID = id
			title = benchmark.Title
			description = benchmark.Description
			version = benchmark.Version
		}

		totalProfiles += len(benchmark.Profiles)
		totalRules += countRules(benchmark)
	}

	s.metadata = &SSGMetadata{
		Version:      version,
		BenchmarkID:  benchmarkID,
		Title:        title,
		Description:  description,
		ProfileCount: totalProfiles,
		RuleCount:    totalRules,
		LastUpdated:  "",
	}
}

// GetMetadata returns metadata about the SSG store
func (s *LocalSSGStore) GetMetadata() *SSGMetadata {
	if s.metadata == nil {
		s.updateMetadata()
	}
	return s.metadata
}

// ListProfiles lists all available SSG profiles
func (s *LocalSSGStore) ListProfiles(offset, limit int) ([]SSGProfileSummary, int, error) {
	var profiles []SSGProfileSummary

	// Collect all profiles from all benchmarks
	for _, benchmark := range s.benchmarks {
		for _, profile := range benchmark.Profiles {
			profiles = append(profiles, SSGProfileSummary{
				ID:          profile.ID,
				Title:       profile.Title,
				Description: profile.Description,
				RuleCount:   len(profile.Selects),
			})
		}
	}

	total := len(profiles)

	// Apply pagination
	if offset >= total {
		return []SSGProfileSummary{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return profiles[offset:end], total, nil
}

// GetProfile retrieves a specific profile by ID
func (s *LocalSSGStore) GetProfile(profileID string) (*SSGProfile, error) {
	for _, benchmark := range s.benchmarks {
		for _, profile := range benchmark.Profiles {
			if profile.ID == profileID {
				return &profile, nil
			}
		}
	}

	return nil, fmt.Errorf(LogMsgSSGProfileNotFound, profileID)
}

// ListRules lists all available SSG rules
func (s *LocalSSGStore) ListRules(offset, limit int, filters map[string]string) ([]SSGRuleSummary, int, error) {
	var rules []SSGRuleSummary

	// Collect all rules from all benchmarks
	for _, benchmark := range s.benchmarks {
		// Add top-level rules
		for _, rule := range benchmark.Rules {
			if matchesFilters(rule, filters) {
				rules = append(rules, SSGRuleSummary{
					ID:          rule.ID,
					Title:       rule.Title,
					Severity:    rule.Severity,
					Description: rule.Description,
				})
			}
		}

		// Add rules from groups
		for _, group := range benchmark.Groups {
			collectGroupRules(&group, &rules, filters)
		}
	}

	total := len(rules)

	// Apply pagination
	if offset >= total {
		return []SSGRuleSummary{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return rules[offset:end], total, nil
}

// collectGroupRules recursively collects rules from a group
func collectGroupRules(group *SSGGroup, rules *[]SSGRuleSummary, filters map[string]string) {
	for _, rule := range group.Rules {
		if matchesFilters(rule, filters) {
			*rules = append(*rules, SSGRuleSummary{
				ID:          rule.ID,
				Title:       rule.Title,
				Severity:    rule.Severity,
				Description: rule.Description,
			})
		}
	}

	for _, subGroup := range group.Groups {
		collectGroupRules(&subGroup, rules, filters)
	}
}

// matchesFilters checks if a rule matches the given filters
func matchesFilters(rule SSGRule, filters map[string]string) bool {
	if severity, ok := filters["severity"]; ok && severity != "" {
		if rule.Severity != severity {
			return false
		}
	}

	if profile, ok := filters["profile"]; ok && profile != "" {
		// This would require checking if the rule is selected by the profile
		// For simplicity, we skip this filter for now
		_ = profile
	}

	return true
}

// GetRule retrieves a specific rule by ID
func (s *LocalSSGStore) GetRule(ruleID string) (*SSGRule, error) {
	for _, benchmark := range s.benchmarks {
		// Check top-level rules
		for _, rule := range benchmark.Rules {
			if rule.ID == ruleID {
				return &rule, nil
			}
		}

		// Check rules in groups
		for _, group := range benchmark.Groups {
			if rule := findRuleInGroup(&group, ruleID); rule != nil {
				return rule, nil
			}
		}
	}

	return nil, fmt.Errorf(LogMsgSSGRuleNotFound, ruleID)
}

// findRuleInGroup recursively finds a rule in a group
func findRuleInGroup(group *SSGGroup, ruleID string) *SSGRule {
	for _, rule := range group.Rules {
		if rule.ID == ruleID {
			return &rule
		}
	}

	for _, subGroup := range group.Groups {
		if rule := findRuleInGroup(&subGroup, ruleID); rule != nil {
			return rule
		}
	}

	return nil
}

// SearchContent searches across SSG content
func (s *LocalSSGStore) SearchContent(query string, offset, limit int) ([]interface{}, int, error) {
	var results []interface{}
	query = strings.ToLower(query)

	// Search profiles
	for _, benchmark := range s.benchmarks {
		for _, profile := range benchmark.Profiles {
			if strings.Contains(strings.ToLower(profile.Title), query) ||
				strings.Contains(strings.ToLower(profile.Description), query) {
				results = append(results, SSGProfileSummary{
					ID:          profile.ID,
					Title:       profile.Title,
					Description: profile.Description,
					RuleCount:   len(profile.Selects),
				})
			}
		}

		// Search rules
		for _, rule := range benchmark.Rules {
			if strings.Contains(strings.ToLower(rule.Title), query) ||
				strings.Contains(strings.ToLower(rule.Description), query) {
				results = append(results, SSGRuleSummary{
					ID:          rule.ID,
					Title:       rule.Title,
					Severity:    rule.Severity,
					Description: rule.Description,
				})
			}
		}

		// Search rules in groups
		for _, group := range benchmark.Groups {
			searchGroupRules(&group, query, &results)
		}
	}

	total := len(results)

	// Apply pagination
	if offset >= total {
		return []interface{}{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return results[offset:end], total, nil
}

// searchGroupRules recursively searches rules in a group
func searchGroupRules(group *SSGGroup, query string, results *[]interface{}) {
	for _, rule := range group.Rules {
		if strings.Contains(strings.ToLower(rule.Title), query) ||
			strings.Contains(strings.ToLower(rule.Description), query) {
			*results = append(*results, SSGRuleSummary{
				ID:          rule.ID,
				Title:       rule.Title,
				Severity:    rule.Severity,
				Description: rule.Description,
			})
		}
	}

	for _, subGroup := range group.Groups {
		searchGroupRules(&subGroup, query, results)
	}
}

// ExportMetadataJSON exports metadata as JSON
func (s *LocalSSGStore) ExportMetadataJSON() ([]byte, error) {
	metadata := s.GetMetadata()
	return json.MarshalIndent(metadata, "", "  ")
}
