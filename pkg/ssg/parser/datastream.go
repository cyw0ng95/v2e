// Package parser provides parsers for SCAP Security Guide data formats.
package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/ssg"
)

// XML namespace constants for SCAP data streams
const (
	XMLNS_DS    = "http://scap.nist.gov/schema/scap/source/1.2"
	XMLNS_XCCDF = "http://checklists.nist.gov/xccdf/1.2"
	XMLNS_HTML  = "http://www.w3.org/1999/xhtml"
	XMLNS_DC    = "http://purl.org/dc/elements/1.1/"
)

// productPattern extracts product name from filenames like "ssg-al2023-ds.xml"
var productPattern = regexp.MustCompile(`^ssg-([^-]+)-ds\.xml$`)

// htmlTagRegex is used to strip HTML tags from XMLString content.
// This is compiled once at package initialization to avoid repeated compilations
// during parsing, which significantly improves performance.
var htmlTagRegex = regexp.MustCompile(`<[^>]+>`)

// ============================================================================
// XML Structure Definitions
// ============================================================================

// DataStreamCollection represents the root element of a SCAP data stream XML file.
type DataStreamCollection struct {
	XMLName    xml.Name     `xml:"data-stream-collection"`
	ID         string       `xml:"id,attr"`
	DataStream []DataStream `xml:"data-stream"`
	Components []Component  `xml:"component"`
}

// DataStream represents a single data stream within the collection.
type DataStream struct {
	ID          string     `xml:"id,attr"`
	ScapVersion string     `xml:"scap-version,attr"`
	Timestamp   string     `xml:"timestamp,attr"`
	Checklists  Checklists `xml:"checklists"`
}

// Checklists contains references to checklist components (XCCDF files).
type Checklists struct {
	ComponentRef []ComponentRef `xml:"component-ref"`
}

// ComponentRef is a reference to a component by ID.
type ComponentRef struct {
	ID   string `xml:"id,attr"`
	Href string `xml:"href,attr"`
}

// Component contains the actual embedded content (XCCDF, OVAL, etc.).
type Component struct {
	ID        string     `xml:"id,attr"`
	Timestamp string     `xml:"timestamp,attr"`
	Benchmark *Benchmark `xml:"Benchmark"`
}

// Benchmark represents an XCCDF 1.2 Benchmark element.
type Benchmark struct {
	ID          string    `xml:"id,attr"`
	Title       string    `xml:"title"`
	Description XMLString `xml:"description"`
	Version     string    `xml:"version"`
	Status      Status    `xml:"status"`
	Profiles    []Profile `xml:"Profile"`
	Groups      []Group   `xml:"Group"`
}

// Status represents the status element with date attribute.
type Status struct {
	Value string `xml:",chardata"`
	Date  string `xml:"date,attr"`
}

// Profile represents an XCCDF Profile element.
type Profile struct {
	ID          string    `xml:"id,attr"`
	Title       XMLString `xml:"title"`
	Description XMLString `xml:"description"`
	Version     string    `xml:"version"`
	Selects     []Select  `xml:"select"`
}

// Select represents a rule selection in a profile.
type Select struct {
	IDRef    string `xml:"idref,attr"`
	Selected string `xml:"selected,attr"` // "true" or "false"
}

// Group represents an XCCDF Group element (hierarchical organization).
type Group struct {
	ID          string    `xml:"id,attr"`
	Title       string    `xml:"title"`
	Description XMLString `xml:"description"`
	Groups      []Group   `xml:"Group"` // Nested groups
	Rules       []Rule    `xml:"Rule"`  // Rules in this group
}

// Rule represents an XCCDF Rule element.
type Rule struct {
	ID          string       `xml:"id,attr"`
	Selected    string       `xml:"selected,attr"` // "true" or "false"
	Severity    string       `xml:"severity,attr"`
	Weight      string       `xml:"weight,attr"`
	Title       string       `xml:"title"`
	Description XMLString    `xml:"description"`
	Rationale   XMLString    `xml:"rationale"`
	Version     string       `xml:"version"`
	References  []Reference  `xml:"reference"`
	Identifiers []Identifier `xml:"ident"`
}

// Reference represents a reference/citation in a rule.
type Reference struct {
	Href  string `xml:"href,attr"`
	Value string `xml:",chardata"`
}

// Identifier represents an external identifier (CCE, CVE, etc.).
type Identifier struct {
	System string `xml:"system,attr"`
	Value  string `xml:",chardata"`
}

// XMLString handles mixed content (text + HTML elements) in XML.
type XMLString struct {
	Content string `xml:",innerxml"`
}

// String returns the plain text content, stripping HTML tags.
func (x XMLString) String() string {
	// Remove HTML tags but preserve text content
	content := x.Content
	// Replace <html:br/> with newline
	content = strings.ReplaceAll(content, "<html:br/>", "\n")
	content = strings.ReplaceAll(content, "<html:br />", "\n")
	// Strip all remaining HTML tags using pre-compiled regex
	content = htmlTagRegex.ReplaceAllString(content, "")
	// Clean up excessive whitespace
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

// ============================================================================
// Parser Functions
// ============================================================================

// ParseDataStreamFile parses a SCAP data stream XML file and extracts SSG models.
// Returns data stream metadata, benchmark, profiles, groups, and rules.
// Uses a streaming approach to parse only the necessary Benchmark component,
// reducing memory overhead for large XML files.
func ParseDataStreamFile(r io.Reader, filename string) (*ssg.SSGDataStream, *ssg.SSGBenchmark, []ssg.SSGDSProfile, []ssg.SSGDSGroup, []ssg.SSGDSRule, error) {
	// Extract product from filename
	product, err := extractProduct(filename)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	decoder := xml.NewDecoder(r)

	// Variables to hold parsed data
	var dataStreamID, scapVersion, timestamp string
	var benchmark *Benchmark

	// Stream through the XML using token-based parsing
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to read XML token: %w", err)
		}

		// Look for start elements
		if se, ok := token.(xml.StartElement); ok {
			switch se.Name.Local {
			case "data-stream":
				// Extract data stream attributes
				if dataStreamID == "" {
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "id":
							dataStreamID = attr.Value
						case "scap-version":
							scapVersion = attr.Value
						case "timestamp":
							timestamp = attr.Value
						}
					}
				}
			case "Benchmark":
				// Found the Benchmark element - decode it directly
				var b Benchmark
				if err := decoder.DecodeElement(&b, &se); err != nil {
					return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode Benchmark: %w", err)
				}
				benchmark = &b
				// We have what we need, can break early
				goto done
			}
		}
	}

done:
	// Validate that we have required data
	if dataStreamID == "" {
		return nil, nil, nil, nil, nil, fmt.Errorf("no data stream found in collection")
	}

	if benchmark == nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("no XCCDF benchmark found in data stream")
	}

	// Create SSGDataStream model
	dataStream := &ssg.SSGDataStream{
		ID:          dataStreamID,
		Product:     product,
		ScapVersion: scapVersion,
		Timestamp:   timestamp,
	}

	// Create SSGBenchmark model
	ssgBenchmark := &ssg.SSGBenchmark{
		ID:           benchmark.ID,
		DataStreamID: dataStream.ID,
		Title:        benchmark.Title,
		Description:  benchmark.Description.String(),
		Version:      benchmark.Version,
		Status:       benchmark.Status.Value,
		StatusDate:   benchmark.Status.Date,
	}

	// Parse profiles
	profiles := make([]ssg.SSGDSProfile, 0, len(benchmark.Profiles))
	for _, p := range benchmark.Profiles {
		profile := ssg.SSGDSProfile{
			ID:          p.ID,
			BenchmarkID: ssgBenchmark.ID,
			Title:       p.Title.String(),
			Description: p.Description.String(),
			Version:     p.Version,
			RuleCount:   len(p.Selects),
		}

		// Parse selected rules
		selectedRules := make([]ssg.SSGDSProfileRule, 0, len(p.Selects))
		for _, sel := range p.Selects {
			selectedRules = append(selectedRules, ssg.SSGDSProfileRule{
				ProfileID: profile.ID,
				RuleID:    sel.IDRef,
				Selected:  sel.Selected == "true",
			})
		}
		profile.SelectedRules = selectedRules

		profiles = append(profiles, profile)
	}

	// Parse groups and rules recursively
	groups := make([]ssg.SSGDSGroup, 0)
	rules := make([]ssg.SSGDSRule, 0)
	parseGroups(benchmark.Groups, ssgBenchmark.ID, "", 0, &groups, &rules)

	// Update counts
	ssgBenchmark.ProfileCount = len(profiles)
	ssgBenchmark.GroupCount = len(groups)
	ssgBenchmark.RuleCount = len(rules)

	return dataStream, ssgBenchmark, profiles, groups, rules, nil
}

// parseGroups recursively parses XCCDF groups and rules.
func parseGroups(xmlGroups []Group, benchmarkID, parentID string, level int, groups *[]ssg.SSGDSGroup, rules *[]ssg.SSGDSRule) {
	for _, g := range xmlGroups {
		// Create group model
		group := ssg.SSGDSGroup{
			ID:          g.ID,
			BenchmarkID: benchmarkID,
			ParentID:    parentID,
			Title:       g.Title,
			Description: g.Description.String(),
			Level:       level,
		}
		*groups = append(*groups, group)

		// Parse rules in this group
		for _, r := range g.Rules {
			rule := ssg.SSGDSRule{
				ID:          r.ID,
				BenchmarkID: benchmarkID,
				GroupID:     g.ID,
				Title:       r.Title,
				Description: r.Description.String(),
				Rationale:   r.Rationale.String(),
				Severity:    r.Severity,
				Selected:    r.Selected == "true",
				Weight:      r.Weight,
				Version:     r.Version,
			}

			// Parse references
			references := make([]ssg.SSGDSRuleReference, 0, len(r.References))
			for _, ref := range r.References {
				references = append(references, ssg.SSGDSRuleReference{
					RuleID: rule.ID,
					Href:   ref.Href,
					RefID:  ref.Value,
				})
			}
			rule.References = references

			// Parse identifiers (CCE, CVE, etc.)
			identifiers := make([]ssg.SSGDSRuleIdentifier, 0, len(r.Identifiers))
			for _, ident := range r.Identifiers {
				identifiers = append(identifiers, ssg.SSGDSRuleIdentifier{
					RuleID:     rule.ID,
					System:     ident.System,
					Identifier: ident.Value,
				})
			}
			rule.Identifiers = identifiers

			*rules = append(*rules, rule)
		}

		// Recursively parse nested groups
		parseGroups(g.Groups, benchmarkID, g.ID, level+1, groups, rules)
	}
}

// extractProduct extracts the product name from a data stream filename.
func extractProduct(filename string) (string, error) {
	matches := productPattern.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid data stream filename format: %s (expected ssg-<product>-ds.xml)", filename)
	}
	return matches[1], nil
}
