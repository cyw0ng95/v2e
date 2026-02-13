package asvs

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cyw0ng95/v2e/pkg/common"
)

var (
	// buildASVSDBPath can be overridden at build time via ldflags
	buildASVSDBPath = ""
	// buildASVSCSVURL can be overridden at build time via ldflags
	buildASVSCSVURL = "https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv"

	// globalHTTPClient is a shared HTTP client with connection pooling
	globalHTTPClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
)

// LocalASVSStore manages a local database of ASVS requirements
type LocalASVSStore struct {
	db *gorm.DB
}

// NewLocalASVSStore creates or opens a local ASVS database at dbPath
func NewLocalASVSStore(dbPath string) (*LocalASVSStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	// Set SQLite PRAGMAs for WAL mode and better concurrency
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=-40000")
		// Set busy_timeout to handle lock contention when multiple services access the database
		db.Exec("PRAGMA busy_timeout=30000")
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&ASVSRequirementModel{}); err != nil {
		return nil, err
	}

	return &LocalASVSStore{db: db}, nil
}

// ImportFromCSV imports ASVS requirements from a CSV URL
func (s *LocalASVSStore) ImportFromCSV(ctx context.Context, url string) error {
	common.Info("Importing ASVS data from URL: %s", url)

	// Download the CSV file
	client := globalHTTPClient

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the CSV
	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Find column indices
	var (
		reqIDIdx   = -1
		chapterIdx = -1
		sectionIdx = -1
		descIdx    = -1
		level1Idx  = -1
		level2Idx  = -1
		level3Idx  = -1
		cweIdx     = -1
	)

	for i, col := range header {
		colLower := strings.ToLower(strings.TrimSpace(col))
		switch {
		case strings.Contains(colLower, "id") && strings.Contains(colLower, "requirement"):
			reqIDIdx = i
		case strings.Contains(colLower, "chapter"):
			chapterIdx = i
		case strings.Contains(colLower, "section"):
			sectionIdx = i
		case strings.Contains(colLower, "description"):
			descIdx = i
		case strings.Contains(colLower, "l1") || strings.Contains(colLower, "level 1"):
			level1Idx = i
		case strings.Contains(colLower, "l2") || strings.Contains(colLower, "level 2"):
			level2Idx = i
		case strings.Contains(colLower, "l3") || strings.Contains(colLower, "level 3"):
			level3Idx = i
		case strings.Contains(colLower, "cwe"):
			cweIdx = i
		}
	}

	if reqIDIdx == -1 || descIdx == -1 {
		return fmt.Errorf("required columns not found in CSV")
	}

	// Read and import records
	var records []ASVSRequirementModel
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			common.Warn("Failed to read CSV row: %v", err)
			continue
		}

		if len(row) <= reqIDIdx {
			continue
		}

		reqID := strings.TrimSpace(row[reqIDIdx])
		if reqID == "" {
			continue
		}

		record := ASVSRequirementModel{
			RequirementID: reqID,
		}

		if chapterIdx >= 0 && chapterIdx < len(row) {
			record.Chapter = strings.TrimSpace(row[chapterIdx])
		}
		if sectionIdx >= 0 && sectionIdx < len(row) {
			record.Section = strings.TrimSpace(row[sectionIdx])
		}
		if descIdx >= 0 && descIdx < len(row) {
			record.Description = strings.TrimSpace(row[descIdx])
		}
		if level1Idx >= 0 && level1Idx < len(row) {
			record.Level1 = parseBoolColumn(row[level1Idx])
		}
		if level2Idx >= 0 && level2Idx < len(row) {
			record.Level2 = parseBoolColumn(row[level2Idx])
		}
		if level3Idx >= 0 && level3Idx < len(row) {
			record.Level3 = parseBoolColumn(row[level3Idx])
		}
		if cweIdx >= 0 && cweIdx < len(row) {
			record.CWE = strings.TrimSpace(row[cweIdx])
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return fmt.Errorf("no valid records found in CSV")
	}

	// Batch insert using GORM
	common.Info("Importing %d ASVS requirements", len(records))
	if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(records, 100).Error; err != nil {
		return fmt.Errorf("failed to insert records: %w", err)
	}

	common.Info("Successfully imported %d ASVS requirements", len(records))
	return nil
}

// parseBoolColumn parses a boolean value from a CSV column
func parseBoolColumn(val string) bool {
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "" {
		return false
	}
	// Check for various true values
	if val == "x" || val == "✓" || val == "true" || val == "yes" || val == "1" {
		return true
	}
	// Try parsing as int
	if i, err := strconv.Atoi(val); err == nil && i > 0 {
		return true
	}
	return false
}

// ListASVSPaginated returns paginated ASVS requirements with optional filters
func (s *LocalASVSStore) ListASVSPaginated(ctx context.Context, offset, limit int, chapter string, level int) ([]ASVSRequirement, int64, error) {
	query := s.db.WithContext(ctx).Model(&ASVSRequirementModel{})

	// Apply filters
	if chapter != "" {
		query = query.Where("chapter = ?", chapter)
	}
	if level > 0 {
		switch level {
		case 1:
			query = query.Where("level1 = ?", true)
		case 2:
			query = query.Where("level2 = ?", true)
		case 3:
			query = query.Where("level3 = ?", true)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var models []ASVSRequirementModel
	if err := query.Offset(offset).Limit(limit).Order("requirement_id ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// Convert models to ASVSRequirement
	requirements := make([]ASVSRequirement, len(models))
	for i, m := range models {
		requirements[i] = ASVSRequirement{
			RequirementID: m.RequirementID,
			Chapter:       m.Chapter,
			Section:       m.Section,
			Description:   m.Description,
			Level1:        m.Level1,
			Level2:        m.Level2,
			Level3:        m.Level3,
			CWE:           m.CWE,
		}
	}

	return requirements, total, nil
}

// GetByID retrieves an ASVS requirement by its ID
func (s *LocalASVSStore) GetByID(ctx context.Context, requirementID string) (*ASVSRequirement, error) {
	var model ASVSRequirementModel
	if err := s.db.WithContext(ctx).Where("requirement_id = ?", requirementID).First(&model).Error; err != nil {
		return nil, err
	}

	return &ASVSRequirement{
		RequirementID: model.RequirementID,
		Chapter:       model.Chapter,
		Section:       model.Section,
		Description:   model.Description,
		Level1:        model.Level1,
		Level2:        model.Level2,
		Level3:        model.Level3,
		CWE:           model.CWE,
	}, nil
}

// GetByCWE retrieves ASVS requirements by CWE
func (s *LocalASVSStore) GetByCWE(ctx context.Context, cwe string) ([]ASVSRequirement, error) {
	var models []ASVSRequirementModel
	if err := s.db.WithContext(ctx).Where("cwe LIKE ?", "%"+cwe+"%").Find(&models).Error; err != nil {
		return nil, err
	}

	requirements := make([]ASVSRequirement, len(models))
	for i, m := range models {
		requirements[i] = ASVSRequirement{
			RequirementID: m.RequirementID,
			Chapter:       m.Chapter,
			Section:       m.Section,
			Description:   m.Description,
			Level1:        m.Level1,
			Level2:        m.Level2,
			Level3:        m.Level3,
			CWE:           m.CWE,
		}
	}

	return requirements, nil
}

// Count returns the total number of ASVS requirements
func (s *LocalASVSStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&ASVSRequirementModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
