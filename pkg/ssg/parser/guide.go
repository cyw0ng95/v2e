// Package parser provides HTML parsing for SSG guide files.
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

// ParseGuideFile parses an SSG HTML guide file and extracts guide, groups, and rules.
func ParseGuideFile(path string) (*ssg.SSGGuide, []ssg.SSGGroup, []ssg.SSGRule, error) {
	// Read HTML content
	htmlContent, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract metadata
	guideID, product, shortID := extractIDFromPath(path)
	profileID, title := extractMetadata(doc)

	if title == "" {
		title = guideID // Fallback to ID if no title found
	}

	// Create guide
	guide := &ssg.SSGGuide{
		ID:          guideID,
		Product:     product,
		ProfileID:   profileID,
		ShortID:     shortID,
		Title:       title,
		HTMLContent: string(htmlContent),
	}

	// Parse tree structure to get all nodes
	nodes, err := parseHTMLTree(doc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse tree: %w", err)
	}

	// Extract groups and rules from nodes
	var groups []ssg.SSGGroup
	var rules []ssg.SSGRule

	for _, node := range nodes {
		if node.Type == "group" {
			group := parseGroupFromNode(node, guideID, doc)
			groups = append(groups, *group)
		} else if node.Type == "rule" {
			rule := parseRuleFromNode(node, guideID, doc)
			rules = append(rules, *rule)
		}
	}

	// Update group counts
	updateGroupCounts(&groups, rules)

	return guide, groups, rules, nil
}

// extractIDFromPath extracts guide ID, product, and short ID from file path.
// Example: "guides/ssg-al2023-guide-cis.html" → ("ssg-al2023-guide-cis", "al2023", "cis")
func extractIDFromPath(path string) (id, product, shortID string) {
	filename := filepath.Base(path)
	// Remove .html extension
	id = strings.TrimSuffix(filename, ".html")

	// Extract product and short ID from name
	// Format: ssg-{product}-guide-{short_id}.html
	re := regexp.MustCompile(`^ssg-([^-]+)-guide-([^.]+)`)
	matches := re.FindStringSubmatch(id)
	if len(matches) == 3 {
		product = matches[1]
		shortID = matches[2]
	}

	return id, product, shortID
}

// extractMetadata extracts title and profile ID from HTML document.
func extractMetadata(doc *goquery.Document) (profileID, title string) {
	// Try to find profile ID in HTML
	// Look for "Profile ID" table entry
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		th := s.Find("th").First()
		if th.Text() == "Profile ID" {
			td := s.Find("td").First()
			profileID = strings.TrimSpace(td.Text())
		}
		if th.Text() == "Profile Title" {
			td := s.Find("td").First()
			title = strings.TrimSpace(td.Text())
		}
	})

	// Also try to find title in h2
	if title == "" {
		doc.Find("h2").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "Guide to the Secure Configuration") {
				title = text
			}
		})
	}

	// Fallback to title tag
	if title == "" {
		if t := doc.Find("title").Text(); t != "" {
			title = strings.TrimSpace(t)
			// Clean up title (remove " | OpenSCAP Security Guide")
			if idx := strings.Index(title, " | OpenSCAP"); idx > 0 {
				title = title[:idx]
			}
		}
	}

	// Extract profile ID from text if in specific format
	re := regexp.MustCompile(`xccdf_org\.ssgproject\.content_profile_([a-z0-9_]+)`)
	if matches := re.FindStringSubmatch(profileID); len(matches) > 1 {
		profileID = matches[0] // Keep full profile ID
	}

	return profileID, title
}

// parseHTMLTree extracts tree structure from data-tt-id/data-tt-parent-id attributes.
func parseHTMLTree(doc *goquery.Document) ([]*ParseTreeNode, error) {
	var nodes []*ParseTreeNode

	// Find all elements with data-tt-id
	doc.Find("[data-tt-id]").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("data-tt-id")
		parentID, _ := s.Attr("data-tt-parent-id")

		// Skip "children-" prefixed IDs (they're just tree expanders)
		if strings.HasPrefix(id, "children-") {
			return
		}

		// Determine type and level
		nodeType := "unknown"
		if strings.Contains(id, "content_group_") {
			nodeType = "group"
		} else if strings.Contains(id, "content_rule_") {
			nodeType = "rule"
		} else if strings.Contains(id, "content_benchmark_") {
			// Skip benchmark root node
			return
		}

		// Calculate level from parent depth
		level := 0
		if parentID != "" {
			// Count depth by looking at parent hierarchy
			level = calculateLevel(doc, id, parentID)
		}

		node := &ParseTreeNode{
			ID:       id,
			ParentID: normalizeParentID(parentID),
			Level:    level,
			Type:     nodeType,
		}
		nodes = append(nodes, node)
	})

	return nodes, nil
}

// normalizeParentID removes "children-" prefix from parent IDs.
func normalizeParentID(parentID string) string {
	if strings.HasPrefix(parentID, "children-") {
		return strings.TrimPrefix(parentID, "children-")
	}
	return parentID
}

// calculateLevel calculates the tree level for a node.
func calculateLevel(doc *goquery.Document, id, parentID string) int {
	level := 0
	currentParent := normalizeParentID(parentID)

	// Walk up the tree
	for currentParent != "" && !strings.Contains(currentParent, "benchmark") {
		level++
		// Find parent element
		var foundParent bool
		doc.Find("[data-tt-id]").Each(func(i int, s *goquery.Selection) {
			if pid, _ := s.Attr("data-tt-id"); pid == currentParent || pid == "children-"+currentParent {
				if ppid, _ := s.Attr("data-tt-parent-id"); ppid != "" {
					currentParent = normalizeParentID(ppid)
					foundParent = true
				}
			}
		})
		if !foundParent {
			break
		}
	}

	return level
}

// ParseTreeNode represents a node during HTML parsing.
type ParseTreeNode struct {
	ID       string
	ParentID string
	Level    int
	Type     string // "group" or "rule"
}

// parseGroupFromNode creates an SSGGroup from a parse tree node.
func parseGroupFromNode(node *ParseTreeNode, guideID string, doc *goquery.Document) *ssg.SSGGroup {
	// Find the element for this group
	var title, description string
	var groupCount, ruleCount int

	// First try to find title from anchor link that references this group
	// Format: <a href="#xccdf_org.ssgproject.content_group_system">System Settings</a>
	doc.Find("a[href*='#"+node.ID+"']").Each(func(i int, s *goquery.Selection) {
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}
	})

	// If no title found, try finding from the data-tt-id element
	doc.Find("[data-tt-id]").Each(func(i int, s *goquery.Selection) {
		if id, _ := s.Attr("data-tt-id"); id == node.ID || id == "children-"+node.ID {
			// Try to find title in the element text
			if title == "" {
				// Look for text after "Group" label
				text := strings.TrimSpace(s.Text())
				// Remove "Group contains..." text
				if idx := strings.Index(text, "Group contains"); idx > 0 {
					text = strings.TrimSpace(text[:idx])
				}
				// Remove "Group" label
				text = strings.TrimSpace(strings.TrimPrefix(text, "Group"))
				if text != "" && len(text) < 200 {
					title = text
				}
			}

			// Look for description in nearby elements
			s.Find(".description, .profile-description").Each(func(i int, desc *goquery.Selection) {
				if description == "" {
					description = strings.TrimSpace(desc.Text())
				}
			})

			// Count children from "Group contains X groups and Y rules" text
			text := s.Text()
			re := regexp.MustCompile(`Group contains (\d+) groups? and (\d+) rules?`)
			if matches := re.FindStringSubmatch(text); len(matches) == 3 {
				fmt.Sscanf(matches[1], "%d", &groupCount)
				fmt.Sscanf(matches[2], "%d", &ruleCount)
			}
		}
	})

	// Fallback: extract title from ID
	if title == "" {
		shortID := extractShortID(node.ID, "group")
		title = strings.ReplaceAll(shortID, "_", " ")
		title = strings.Title(title)
	}

	return &ssg.SSGGroup{
		ID:          node.ID,
		GuideID:     guideID,
		ParentID:    node.ParentID,
		Title:       cleanTitle(title),
		Description: description,
		Level:       node.Level,
		GroupCount:  groupCount,
		RuleCount:   ruleCount,
	}
}

// parseRuleFromNode creates an SSGRule from a parse tree node.
func parseRuleFromNode(node *ParseTreeNode, guideID string, doc *goquery.Document) *ssg.SSGRule {
	// Find the element for this rule
	var title, description, rationale, severity string

	doc.Find("[data-tt-id]").Each(func(i int, s *goquery.Selection) {
		if id, _ := s.Attr("data-tt-id"); id == node.ID {
			// Try to find title - look for "Rule" label or nearby text
			s.Find(".label-default").Each(func(i int, label *goquery.Selection) {
				if label.Text() == "Rule" {
					// Title is usually the text after "Rule" label
					title = strings.TrimSpace(label.Parent().Text())
					title = strings.TrimPrefix(title, "Rule")
					title = strings.TrimSpace(title)
				}
			})

			// If no title found, try extracting from ID
			if title == "" {
				title = extractShortID(node.ID, "rule")
				title = strings.ReplaceAll(title, "_", " ")
				title = strings.Title(title)
			}

			// Look for description
			s.Find(".description").Each(func(i int, desc *goquery.Selection) {
				if description == "" {
					description = strings.TrimSpace(desc.Text())
				}
			})

			// Look for rationale
			s.Find(".rationale").Each(func(i int, rat *goquery.Selection) {
				if rationale == "" {
					rationale = strings.TrimSpace(rat.Text())
				}
			})

			// Look for severity
			s.Find(".severity").Each(func(i int, sev *goquery.Selection) {
				if severity == "" {
					severity = strings.ToLower(strings.TrimSpace(sev.Text()))
				}
			})

			// Also try to find severity in class attributes
			if severity == "" {
				if s.HasClass("severity-high") || s.Find(".severity-high").Length() > 0 {
					severity = "high"
				} else if s.HasClass("severity-medium") || s.Find(".severity-medium").Length() > 0 {
					severity = "medium"
				} else if s.HasClass("severity-low") || s.Find(".severity-low").Length() > 0 {
					severity = "low"
				}
			}
		}
	})

	// Default severity if not found
	if severity == "" {
		severity = "medium"
	}

	// Extract short ID from full ID
	shortID := extractShortID(node.ID, "rule")

	// Find parent group from parent ID
	// If parentID is a "children-" prefixed ID, extract the actual group ID
	groupID := node.ParentID
	if strings.HasPrefix(groupID, "children-") {
		groupID = strings.TrimPrefix(groupID, "children-")
	}

	return &ssg.SSGRule{
		ID:          node.ID,
		GuideID:     guideID,
		GroupID:     groupID,
		ShortID:     shortID,
		Title:       cleanTitle(title),
		Description: description,
		Rationale:   rationale,
		Severity:    severity,
		Level:       node.Level,
		References:  []ssg.SSGReference{}, // TODO: Parse references
	}
}

// extractShortID extracts a short ID from a full XCCDF ID.
// e.g., "xccdf_org.ssgproject.content_group_system" → "system"
// e.g., "xccdf_org.ssgproject.content_rule_package_aide_installed" → "package_aide_installed"
func extractShortID(fullID, elementType string) string {
	prefix := "xccdf_org.ssgproject.content_" + elementType + "_"
	if strings.HasPrefix(fullID, prefix) {
		return strings.TrimPrefix(fullID, prefix)
	}
	return fullID
}

// cleanTitle cleans up the title text.
func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	// Remove extra whitespace
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	return title
}

// updateGroupCounts updates group and rule counts for each group.
func updateGroupCounts(groups *[]ssg.SSGGroup, rules []ssg.SSGRule) {
	groupMap := make(map[string]*ssg.SSGGroup)
	for i := range *groups {
		groupMap[(*groups)[i].ID] = &(*groups)[i]
	}

	// Count children for each group
	for _, rule := range rules {
		if parent, ok := groupMap[rule.GroupID]; ok {
			parent.RuleCount++
		}
	}

	for _, group := range *groups {
		if group.ParentID != "" {
			if parent, ok := groupMap[group.ParentID]; ok {
				parent.GroupCount++
			}
		}
	}
}
