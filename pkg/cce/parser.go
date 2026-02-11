package cce

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

const (
	CCEIDIndex       = 0 // Column A: CCE ID
	TitleIndex       = 1 // Column B: Title
	DescriptionIndex = 2 // Column C: Description
	OwnerIndex       = 3 // Column D: Owner
	StatusIndex      = 4 // Column E: Status
	TypeIndex        = 5 // Column F: Type
	ReferenceIndex   = 6 // Column G: Reference
)

// Parser handles parsing of CCE Excel files
type Parser struct {
	filePath string
}

// NewParser creates a new CCE parser
func NewParser(filePath string) *Parser {
	return &Parser{
		filePath: filePath,
	}
}

// ParseAll reads the Excel file and returns all CCE entries
func (p *Parser) ParseAll() ([]CCE, error) {
	f, err := excelize.OpenFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get all sheet names
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	// Use the first sheet
	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no data rows found")
	}

	var cceEntries []CCE
	// Skip header row (first row), start from index 1
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 4 {
			continue // Skip rows with insufficient columns
		}

		// Skip empty ID rows
		cceID := row[CCEIDIndex]
		if cceID == "" {
			continue
		}

		entry := CCE{
			ID:          cceID,
			Title:       getCellValue(row, TitleIndex),
			Description: getCellValue(row, DescriptionIndex),
			Owner:       getCellValue(row, OwnerIndex),
			Status:      getCellValue(row, StatusIndex),
			Type:        getCellValue(row, TypeIndex),
			Reference:   getCellValue(row, ReferenceIndex),
			Metadata:    "",
		}

		cceEntries = append(cceEntries, entry)
	}

	return cceEntries, nil
}

// ParseBatch reads the Excel file and returns a batch of CCE entries
func (p *Parser) ParseBatch(offset, batchSize int) ([]CCE, int, error) {
	f, err := excelize.OpenFile(p.filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, 0, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, 0, fmt.Errorf("no data rows found")
	}

	totalRows := len(rows) - 1 // Exclude header
	startRow := offset + 1     // +1 for header
	endRow := startRow + batchSize

	if startRow >= len(rows) {
		return nil, totalRows, nil
	}

	if endRow > len(rows) {
		endRow = len(rows)
	}

	var cceEntries []CCE
	for i := startRow; i < endRow; i++ {
		row := rows[i]
		if len(row) < 4 {
			continue
		}

		cceID := row[CCEIDIndex]
		if cceID == "" {
			continue
		}

		entry := CCE{
			ID:          cceID,
			Title:       getCellValue(row, TitleIndex),
			Description: getCellValue(row, DescriptionIndex),
			Owner:       getCellValue(row, OwnerIndex),
			Status:      getCellValue(row, StatusIndex),
			Type:        getCellValue(row, TypeIndex),
			Reference:   getCellValue(row, ReferenceIndex),
			Metadata:    "",
		}

		cceEntries = append(cceEntries, entry)
	}

	return cceEntries, totalRows, nil
}

// getCellValue safely retrieves a cell value by index
func getCellValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

// ParseRowCount returns the total number of CCE entries (excluding header)
func (p *Parser) ParseRowCount() (int, error) {
	f, err := excelize.OpenFile(p.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	return len(rows) - 1, nil // Exclude header row
}

// ParseByCCEID looks up a specific CCE by ID
func (p *Parser) ParseByCCEID(cceID string) (*CCE, error) {
	f, err := excelize.OpenFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	// Skip header row
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 1 {
			continue
		}

		if row[CCEIDIndex] == cceID {
			entry := &CCE{
				ID:          row[CCEIDIndex],
				Title:       getCellValue(row, TitleIndex),
				Description: getCellValue(row, DescriptionIndex),
				Owner:       getCellValue(row, OwnerIndex),
				Status:      getCellValue(row, StatusIndex),
				Type:        getCellValue(row, TypeIndex),
				Reference:   getCellValue(row, ReferenceIndex),
				Metadata:    "",
			}
			return entry, nil
		}
	}

	return nil, fmt.Errorf("CCE %s not found", cceID)
}

// GetOffsetFromCCEID returns the offset (row index) for a given CCE ID
func (p *Parser) GetOffsetFromCCEID(cceID string) (int, error) {
	f, err := excelize.OpenFile(p.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 1 {
			continue
		}

		if row[CCEIDIndex] == cceID {
			return i - 1, nil // Offset excludes header
		}
	}

	return 0, fmt.Errorf("CCE %s not found", cceID)
}

// ParseNumericOffset attempts to parse a numeric offset from checkpoint URN
func ParseNumericOffset(urnKey string) (int, error) {
	return strconv.Atoi(urnKey)
}
