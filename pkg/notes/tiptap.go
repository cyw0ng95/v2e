package notes

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TipTap errors
var (
	ErrInvalidTipTapJSON  = errors.New("invalid TipTap JSON")
	ErrMissingDocument    = errors.New("TipTap document node is required")
	ErrInvalidNodeType    = errors.New("invalid TipTap node type")
	ErrMissingContent     = errors.New("TipTap content array is required")
	ErrInvalidTextContent = errors.New("invalid text content")
)

// TipTapDocument represents a TipTap document structure
// See: https://tiptap.dev/docs/editor/api/document
type TipTapDocument struct {
	Type     string       `json:"type"`
	Content  []TipTapNode `json:"content,omitempty"`
	Attrs    interface{}  `json:"attrs,omitempty"`
	Markdown string       `json:"markdown,omitempty"`
	Text     string       `json:"text,omitempty"`
}

// TipTapNode represents a generic TipTap node
type TipTapNode struct {
	Type    string       `json:"type"`
	Content []TipTapNode `json:"content,omitempty"`
	Attrs   interface{}  `json:"attrs,omitempty"`
	Text    string       `json:"text,omitempty"`
	Marks   []TipTapMark `json:"marks,omitempty"`
}

// TipTapMark represents text formatting marks (bold, italic, etc.)
type TipTapMark struct {
	Type  string            `json:"type"`
	Attrs map[string]string `json:"attrs,omitempty"`
}

// ValidTipTapNodeTypes defines the allowed TipTap node types
var ValidTipTapNodeTypes = map[string]bool{
	// Document structure
	"doc":         true,
	"paragraph":   true,
	"heading":     true,
	"codeBlock":   true,
	"blockquote":  true,
	"listItem":    true,
	"bulletList":  true,
	"orderedList": true,
	"text":        true,
	"hardBreak":   true,

	// Formatting
	"bold":   true,
	"italic": true,
	"strike": true,
	"code":   true,
	"link":   true,

	// Task lists
	"taskList": true,
	"taskItem": true,

	// Common extensions
	"image":          true,
	"horizontalRule": true,
}

// ValidTipTapMarkTypes defines the allowed TipTap mark types
var ValidTipTapMarkTypes = map[string]bool{
	"bold":   true,
	"italic": true,
	"strike": true,
	"code":   true,
	"link":   true,
}

// ValidateTipTapJSON validates that a JSON string is a valid TipTap document
func ValidateTipTapJSON(content string) error {
	if content == "" {
		return nil // Empty content is valid (no content yet)
	}

	var doc TipTapDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return fmt.Errorf("%w: JSON parse error: %v", ErrInvalidTipTapJSON, err)
	}

	return ValidateTipTapDocument(&doc)
}

// ValidateTipTapDocument validates a TipTap document structure
func ValidateTipTapDocument(doc *TipTapDocument) error {
	if doc == nil {
		return ErrMissingDocument
	}

	// Root node must be "doc"
	if doc.Type != "doc" {
		return fmt.Errorf("%w: root type must be 'doc', got '%s'", ErrInvalidNodeType, doc.Type)
	}

	// Validate content nodes
	for i, node := range doc.Content {
		if err := validateTipTapNode(&node); err != nil {
			return fmt.Errorf("content[%d]: %w", i, err)
		}
	}

	return nil
}

// validateTipTapNode recursively validates a TipTap node
func validateTipTapNode(node *TipTapNode) error {
	if node == nil {
		return ErrMissingContent
	}

	// Check node type is valid
	if !ValidTipTapNodeTypes[node.Type] {
		return fmt.Errorf("%w: unknown node type '%s'", ErrInvalidNodeType, node.Type)
	}

	// Text nodes must have text content
	if node.Type == "text" {
		if node.Text == "" && len(node.Content) == 0 {
			return ErrInvalidTextContent
		}
	}

	// Validate marks if present
	for _, mark := range node.Marks {
		if !ValidTipTapMarkTypes[mark.Type] {
			return fmt.Errorf("%w: unknown mark type '%s'", ErrInvalidNodeType, mark.Type)
		}
		// Link marks must have href
		if mark.Type == "link" && mark.Attrs != nil {
			if _, ok := mark.Attrs["href"]; !ok {
				return fmt.Errorf("%w: link mark must have href attribute", ErrInvalidNodeType)
			}
		}
	}

	// Recursively validate child content
	for i, child := range node.Content {
		if err := validateTipTapNode(&child); err != nil {
			return fmt.Errorf("child[%d]: %w", i, err)
		}
	}

	return nil
}

// IsTipTapJSONEmpty checks if TipTap content is effectively empty
func IsTipTapJSONEmpty(content string) bool {
	if content == "" {
		return true
	}

	var doc TipTapDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return true
	}

	return len(doc.Content) == 0
}

// GetTipTapText extracts plain text from TipTap JSON content
func GetTipTapText(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	var doc TipTapDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return "", err
	}

	// Extract text from document content nodes
	result := ""
	for _, node := range doc.Content {
		result += extractTextFromNode(&node)
	}
	return result, nil
}

// extractTextFromNode recursively extracts text from a TipTap node
func extractTextFromNode(node *TipTapNode) string {
	if node == nil {
		return ""
	}

	// Text node returns its text
	if node.Type == "text" {
		return node.Text
	}

	// Hard break becomes newline
	if node.Type == "hardBreak" {
		return "\n"
	}

	// Recursively extract from children
	result := ""
	for _, child := range node.Content {
		result += extractTextFromNode(&child)
	}

	// Add newlines after block-level elements
	switch node.Type {
	case "paragraph", "heading", "codeBlock", "blockquote", "listItem", "bulletList", "orderedList", "taskList", "taskItem":
		if result != "" {
			result += "\n"
		}
	}

	return result
}

// CreateEmptyTipTapDocument creates an empty TipTap document
func CreateEmptyTipTapDocument() string {
	doc := TipTapDocument{
		Type: "doc",
		Content: []TipTapNode{
			{
				Type:    "paragraph",
				Content: []TipTapNode{},
			},
		},
	}
	data, _ := json.Marshal(doc)
	return string(data)
}

// CreateTipTapDocumentFromText creates a TipTap document from plain text
func CreateTipTapDocumentFromText(text string) string {
	if text == "" {
		return CreateEmptyTipTapDocument()
	}

	doc := TipTapDocument{
		Type: "doc",
		Content: []TipTapNode{
			{
				Type: "paragraph",
				Content: []TipTapNode{
					{
						Type: "text",
						Text: text,
					},
				},
			},
		},
	}
	data, _ := json.Marshal(doc)
	return string(data)
}
