// Package parser provides HTML parsing for SSG table files.
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cyw0ng95/v2e/pkg/ssg"
)

// ParseTableFile parses an SSG HTML table file and extracts table and entries.
func ParseTableFile(path string) (*ssg.SSGTable, []ssg.SSGTableEntry, error) {
	// Read HTML content
	htmlContent, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract metadata
	tableID, product, tableType := extractTableIDFromPath(path)
	title := extractTableTitle(doc)

	if title == "" {
		title = tableID // Fallback to ID if no title found
	}

	// Create table
	table := &ssg.SSGTable{
		ID:          tableID,
		Product:     product,
		TableType:   tableType,
		Title:       title,
		Description: "",
		HTMLContent: string(htmlContent),
	}

	// Parse table entries
	entries, err := parseTableEntries(doc, tableID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse table entries: %w", err)
	}

	return table, entries, nil
}

// extractTableIDFromPath extracts table ID, product, and table type from file path.
// Example: "tables/table-al2023-cces.html" → ("table-al2023-cces", "al2023", "cces")
func extractTableIDFromPath(path string) (id, product, tableType string) {
	filename := filepath.Base(path)
	// Remove .html extension
	id = strings.TrimSuffix(filename, ".html")

	// Extract product and table type from name
	// Format: table-{product}-{type}.html
	re := regexp.MustCompile(`^table-([^-]+)-(.+)`)
	matches := re.FindStringSubmatch(id)
	if len(matches) == 3 {
		product = matches[1]
		tableType = matches[2]
	}

	return id, product, tableType
}

// extractTableTitle extracts the title from the HTML document.
func extractTableTitle(doc *goquery.Document) string {
	// Try to find title in h1 or div with large font
	var title string

	// Check for centered div with large font (common pattern in SSG tables)
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		style, _ := s.Attr("style")
		if strings.Contains(style, "text-align: center") && strings.Contains(style, "font-size: x-large") {
			if title == "" {
				title = strings.TrimSpace(s.Text())
			}
		}
	})

	// Fallback to title tag
	if title == "" {
		if t := doc.Find("title").Text(); t != "" {
			title = strings.TrimSpace(t)
		}
	}

	return title
}

// parseTableEntries parses table rows into SSGTableEntry records.
func parseTableEntries(doc *goquery.Document, tableID string) ([]ssg.SSGTableEntry, error) {
	var entries []ssg.SSGTableEntry

	// Find the main table
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		var entry ssg.SSGTableEntry
		entry.TableID = tableID

		// Parse cells in order: Mapping, Rule Title, Description, Rationale
		cells := s.Find("td")
		if cells.Length() < 4 {
			// Skip incomplete rows
			return
		}

		entry.Mapping = strings.TrimSpace(cells.Eq(0).Text())
		entry.RuleTitle = strings.TrimSpace(cells.Eq(1).Text())
		entry.Description = strings.TrimSpace(cells.Eq(2).Text())
		entry.Rationale = strings.TrimSpace(cells.Eq(3).Text())

		// Skip rows with empty mapping (likely header or invalid)
		if entry.Mapping == "" {
			return
		}

		entries = append(entries, entry)
	})

	return entries, nil
}
