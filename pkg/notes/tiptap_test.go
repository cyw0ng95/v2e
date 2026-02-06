package notes

import (
	"testing"
)

func TestValidateTipTapJSON_Empty(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty string",
			content: "",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			content: "   ",
			wantErr: true, // Invalid JSON
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTipTapJSON(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTipTapJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTipTapJSON_ValidDocuments(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty doc",
			content: `{"type":"doc","content":[]}`,
			wantErr: false,
		},
		{
			name:    "doc with paragraph",
			content: `{"type":"doc","content":[{"type":"paragraph"}]}`,
			wantErr: false,
		},
		{
			name:    "doc with text",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello"}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with heading",
			content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Title"}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with bold text",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello ","marks":[{"type":"bold"}]},{"type":"text","text":"World"}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with italic",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Italic","marks":[{"type":"italic"}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with code",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"code","marks":[{"type":"code"}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with link",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"link","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with code block",
			content: `{"type":"doc","content":[{"type":"codeBlock","content":[{"type":"text","text":"console.log('hello')"}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with bullet list",
			content: `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with ordered list",
			content: `{"type":"doc","content":[{"type":"orderedList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with blockquote",
			content: `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"Quote"}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with hard break",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Line 1"},{"type":"hardBreak"},{"type":"text","text":"Line 2"}]}]}`,
			wantErr: false,
		},
		{
			name:    "doc with task list",
			content: `{"type":"doc","content":[{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"Task"}]}]}]}]}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTipTapJSON(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTipTapJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTipTapJSON_InvalidDocuments(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			content: `{not json`,
			wantErr: true,
		},
		{
			name:    "not a doc type",
			content: `{"type":"paragraph","content":[]}`,
			wantErr: true,
		},
		{
			name:    "unknown node type",
			content: `{"type":"doc","content":[{"type":"unknown","content":[]}]}`,
			wantErr: true,
		},
		{
			name:    "unknown mark type",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello","marks":[{"type":"unknown"}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "link without href - allowed (empty attrs)",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"link","marks":[{"type":"link"}]}]}]}`,
			wantErr: false, // Empty attrs is allowed for links (will be set by editor)
		},
		{
			name:    "empty text node",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text"}]}]}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTipTapJSON(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTipTapJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsTipTapJSONEmpty(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantEmpty bool
	}{
		{
			name:     "empty string",
			content:  "",
			wantEmpty: true,
		},
		{
			name:     "empty doc",
			content:  `{"type":"doc","content":[]}`,
			wantEmpty: true,
		},
		{
			name:     "doc with empty paragraph",
			content:  `{"type":"doc","content":[{"type":"paragraph"}]}`,
			wantEmpty: false, // Paragraph node exists, so not empty
		},
		{
			name:     "doc with text",
			content:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello"}]}]}`,
			wantEmpty: false,
		},
		{
			name:     "invalid JSON",
			content:  `not json`,
			wantEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTipTapJSONEmpty(tt.content); got != tt.wantEmpty {
				t.Errorf("IsTipTapJSONEmpty() = %v, wantEmpty %v", got, tt.wantEmpty)
			}
		})
	}
}

func TestGetTipTapText(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "empty string",
			content: "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "simple text",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World"}]}]}`,
			want:    "Hello World\n",
			wantErr: false,
		},
		{
			name:    "multiple paragraphs",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"First"}]},{"type":"paragraph","content":[{"type":"text","text":"Second"}]}]}`,
			want:    "First\nSecond\n",
			wantErr: false,
		},
		{
			name:    "bold text",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello ","marks":[{"type":"bold"}]},{"type":"text","text":"World"}]}]}`,
			want:    "Hello World\n",
			wantErr: false,
		},
		{
			name:    "heading",
			content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Title"}]}]}`,
			want:    "Title\n",
			wantErr: false,
		},
		{
			name:    "hard break",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Line 1"},{"type":"hardBreak"},{"type":"text","text":"Line 2"}]}]}`,
			want:    "Line 1\nLine 2\n",
			wantErr: false,
		},
		{
			name:    "bullet list",
			content: `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]}]}]}`,
			want:    "Item 1\n\n\n",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			content: `not json`,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTipTapText(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTipTapText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTipTapText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateEmptyTipTapDocument(t *testing.T) {
	doc := CreateEmptyTipTapDocument()
	if doc == "" {
		t.Fatal("CreateEmptyTipTapDocument() returned empty string")
	}
	err := ValidateTipTapJSON(doc)
	if err != nil {
		t.Errorf("CreateEmptyTipTapDocument() created invalid document: %v", err)
	}
}

func TestCreateTipTapDocumentFromText(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "empty text",
			text: "",
		},
		{
			name: "simple text",
			text: "Hello World",
		},
		{
			name: "multiline text",
			text: "Line 1\nLine 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := CreateTipTapDocumentFromText(tt.text)
			if doc == "" {
				t.Fatal("CreateTipTapDocumentFromText() returned empty string")
			}
			err := ValidateTipTapJSON(doc)
			if err != nil {
				t.Errorf("CreateTipTapDocumentFromText() created invalid document: %v", err)
			}
			// Verify the text is preserved
			extracted, err := GetTipTapText(doc)
			if err != nil {
				t.Errorf("GetTipTapText() error: %v", err)
			}
			if tt.text != "" && extracted != tt.text+"\n" {
				t.Errorf("GetTipTapText() = %q, want %q\\n", extracted, tt.text)
			}
		})
	}
}

// Round-trip test: serialize and deserialize should preserve content
func TestTipTapRoundTrip(t *testing.T) {
	original := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello ","marks":[{"type":"bold"}]},{"type":"text","text":"World"}]}]}`

	// Validate original
	err := ValidateTipTapJSON(original)
	if err != nil {
		t.Fatalf("ValidateTipTapJSON() error = %v", err)
	}

	// Extract text
	text, err := GetTipTapText(original)
	if err != nil {
		t.Fatalf("GetTipTapText() error = %v", err)
	}

	// Check text contains "Hello World"
	if text == "" {
		t.Fatal("GetTipTapText() returned empty text")
	}

	// Trim trailing newlines before recreating (paragraph adds newline)
	trimmedText := text
	for len(trimmedText) > 0 && trimmedText[len(trimmedText)-1] == '\n' {
		trimmedText = trimmedText[:len(trimmedText)-1]
	}

	// Recreate from trimmed text
	recreated := CreateTipTapDocumentFromText(trimmedText)

	// Validate recreated
	err = ValidateTipTapJSON(recreated)
	if err != nil {
		t.Fatalf("ValidateTipTapJSON() on recreated doc error = %v", err)
	}

	// Text should match
	recreatedText, err := GetTipTapText(recreated)
	if err != nil {
		t.Fatalf("GetTipTapText() on recreated doc error = %v", err)
	}

	// Both should extract to same plain text
	if text != recreatedText {
		t.Errorf("Round-trip text mismatch: original = %q, recreated = %q", text, recreatedText)
	}
}
