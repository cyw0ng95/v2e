package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// stubCAPECStore provides a lightweight in-memory store for handler tests.
type stubCAPECStore struct {
	importErr     error
	meta          *capec.CAPECCatalogMeta
	listItems     []capec.CAPECItemModel
	listTotal     int64
	listErr       error
	getItem       *capec.CAPECItemModel
	getErr        error
	related       []capec.CAPECRelatedWeaknessModel
	relatedErr    error
	examples      []capec.CAPECExampleModel
	exampleErr    error
	mitigations   []capec.CAPECMitigationModel
	mitigationErr error
	references    []capec.CAPECReferenceModel
	refErr        error
	lastImport    struct {
		path  string
		xsd   string
		force bool
	}
}

func (s *stubCAPECStore) ImportFromXML(xmlPath string, force bool) error {
	s.lastImport = struct {
		path  string
		xsd   string
		force bool
	}{path: xmlPath, force: force}
	return s.importErr
}

func (s *stubCAPECStore) GetCatalogMeta(ctx context.Context) (*capec.CAPECCatalogMeta, error) {
	if s.meta == nil {
		return nil, errors.New("no meta")
	}
	return s.meta, nil
}

func (s *stubCAPECStore) ListCAPECsPaginated(ctx context.Context, offset, limit int) ([]capec.CAPECItemModel, int64, error) {
	if s.listErr != nil {
		return nil, 0, s.listErr
	}
	return s.listItems, s.listTotal, nil
}

func (s *stubCAPECStore) GetByID(ctx context.Context, capecID string) (*capec.CAPECItemModel, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.getItem != nil {
		return s.getItem, nil
	}
	return &capec.CAPECItemModel{CAPECID: 1, Name: "stub", Summary: "<p>sum</p>", Description: "<p>desc</p>"}, nil
}

func (s *stubCAPECStore) GetRelatedWeaknesses(ctx context.Context, capecID int) ([]capec.CAPECRelatedWeaknessModel, error) {
	if s.relatedErr != nil {
		return nil, s.relatedErr
	}
	return s.related, nil
}

func (s *stubCAPECStore) GetExamples(ctx context.Context, capecID int) ([]capec.CAPECExampleModel, error) {
	if s.exampleErr != nil {
		return nil, s.exampleErr
	}
	return s.examples, nil
}

func (s *stubCAPECStore) GetMitigations(ctx context.Context, capecID int) ([]capec.CAPECMitigationModel, error) {
	if s.mitigationErr != nil {
		return nil, s.mitigationErr
	}
	return s.mitigations, nil
}

func (s *stubCAPECStore) GetReferences(ctx context.Context, capecID int) ([]capec.CAPECReferenceModel, error) {
	if s.refErr != nil {
		return nil, s.refErr
	}
	return s.references, nil
}

func TestXmlInnerToPlain_StripsTagsAndUnescapes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestXmlInnerToPlain_StripsTagsAndUnescapes", nil, func(t *testing.T, tx *gorm.DB) {
		in := "<p>Hello &amp; <strong>World</strong></p>\n<em>!</em>"
		out := xmlInnerToPlain(in)
		if out != "Hello & World !" {
			t.Fatalf("unexpected output: %q", out)
		}
	})

}

func TestXmlInnerToPlain_CoversManyCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestXmlInnerToPlain_CoversManyCases", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name  string
			input string
			want  string
		}{
			{name: "empty", input: "", want: ""},
			{name: "spaces only", input: "   \t\n  ", want: ""},
			{name: "plain text", input: "Hello", want: "Hello"},
			{name: "trimmed text", input: "  Hello World  ", want: "Hello World"},
			{name: "single tag", input: "<p>Hello</p>", want: "Hello"},
			{name: "nested tags", input: "<div><b>Bold</b> <i>Italic</i></div>", want: "Bold Italic"},
			{name: "self closing", input: "Start<br/>End", want: "Start End"},
			{name: "attributes", input: "<a href=\"#\">Link</a>", want: "Link"},
			{name: "entities", input: "Fish &amp; Chips", want: "Fish & Chips"},
			{name: "multi paragraphs", input: "<p>first</p>\n<p>second</p>  <div>third</div>", want: "first second third"},
			{name: "uppercase tags", input: "<DIV>UP</DIV>", want: "UP"},
			{name: "mixed case tags", input: "<Div><Span>Case</Span></Div>", want: "Case"},
			{name: "namespaced", input: "<ns:item xmlns:ns=\"x\">Value</ns:item>", want: "Value"},
			{name: "default xmlns", input: "<root xmlns=\"urn:x\">Body</root>", want: "Body"},
			{name: "comment content", input: "<p><!-- note --></p>", want: ""},
			{name: "doctype", input: "<!DOCTYPE html><p>Doc</p>", want: "Doc"},
			{name: "script removed", input: "<script>alert(1)</script>Safe", want: "alert(1) Safe"},
			{name: "style removed", input: "<style>.x{}</style>Text", want: ".x{} Text"},
			{name: "line breaks", input: "Line1\nLine2", want: "Line1 Line2"},
			{name: "tabs collapse", input: "A\t\tB", want: "A B"},
			{name: "angle brackets text", input: "Use < and > as text", want: "Use as text"},
			{name: "multiple xmlns", input: "<r xmlns=\"a\" xmlns:x=\"b\">z</r>", want: "z"},
			{name: "attribute xmlns", input: "<r xmlns:foo=\"bar\">keep</r>", want: "keep"},
			{name: "html escaped quotes", input: "&quot;test&quot;", want: "\"test\""},
			{name: "empty tags", input: "<p></p>", want: ""},
			{name: "whitespace between tags", input: "<p> a </p><p> b </p>", want: "a b"},
			{name: "non breaking space", input: "a&nbsp;b", want: "a\u00a0b"},
			{name: "escaped slash", input: "a &#x2F; b", want: "a / b"},
			{name: "unicode entity", input: "caf&#233;", want: "café"},
			{name: "xml header", input: "<?xml version=\"1.0\"?><root>t</root>", want: "t"},
			{name: "cdata like", input: "<![CDATA[safe]]><p>keep</p>", want: "keep"},
			{name: "multiple breaks", input: "<br><br><p>end</p>", want: "end"},
			{name: "surrounding whitespace", input: " \t<p>x</p> \n", want: "x"},
			{name: "long whitespace collapse", input: "a    b\n\n c", want: "a b c"},
			{name: "tag with newline", input: "<p>line\nwrap</p>", want: "line wrap"},
			{name: "tag with tabs", input: "<p>line\twith\ttabs</p>", want: "line with tabs"},
			{name: "semicolon entity", input: "rock &amp; roll", want: "rock & roll"},
			{name: "angle entity", input: "&lt;safe&gt;", want: "<safe>"},
			{name: "percent encoded", input: "%3Ctag%3E", want: "%3Ctag%3E"},
			{name: "braced attr", input: "<p data-x='1'>txt</p>", want: "txt"},
			{name: "newline tags", input: "<p>top</p>\n<div>bottom</div>", want: "top bottom"},
			{name: "carriage returns", input: "a\r\nb", want: "a b"},
			{name: "angle spaced", input: "< p >weird</ p >", want: "weird"},
			{name: "duplicate spaces", input: "a  b  c", want: "a b c"},
			{name: "emoji text", input: "😀<b>ok</b>", want: "😀 ok"},
			{name: "numeric entity", input: "&#169; 2024", want: "© 2024"},
			{name: "mixed whitespace", input: "\ta \n b", want: "a b"},
		}

		for i := 0; i < 55; i++ {
			cases = append(cases, struct {
				name  string
				input string
				want  string
			}{
				name:  fmt.Sprintf("generated-%02d", i),
				input: fmt.Sprintf("<p xmlns=\"urn:\">value %d</p>", i),
				want:  fmt.Sprintf("value %d", i),
			})
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := xmlInnerToPlain(tc.input)
				if got != tc.want {
					t.Fatalf("unexpected output: want %q got %q", tc.want, got)
				}
				if strings.Contains(got, "xmlns") {
					t.Fatalf("output leaked xmlns: %q", got)
				}
				if strings.Contains(got, "  ") {
					t.Fatalf("output not collapsed: %q", got)
				}
			})
		}
	})

}

func TestCreateImportCAPECsHandler_ValidatesInput(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateImportCAPECsHandler_ValidatesInput", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{}
		handler := createImportCAPECsHandler(store, logger)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCImportCAPECs"}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "path is required" {
			t.Fatalf("expected validation error, got %+v", resp)
		}
	})

}

func TestCreateImportCAPECsHandler_Succeeds(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateImportCAPECsHandler_Succeeds", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{}
		handler := createImportCAPECsHandler(store, logger)

		payload, _ := subprocess.MarshalFast(map[string]any{"path": "file.xml", "force": true})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCImportCAPECs", Payload: payload}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected success response, got %+v", resp)
		}
		if !store.lastImport.force || store.lastImport.path != "file.xml" {
			t.Fatalf("import parameters not recorded correctly: %+v", store.lastImport)
		}
	})

}

func TestCreateImportCAPECsHandler_PropagatesStoreError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateImportCAPECsHandler_PropagatesStoreError", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{importErr: errors.New("boom")}
		handler := createImportCAPECsHandler(store, logger)

		payload, _ := subprocess.MarshalFast(map[string]string{"path": "file.xml"})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCImportCAPECs", Payload: payload}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "failed to import CAPECs" {
			t.Fatalf("expected store failure, got %+v", resp)
		}
	})

}

func TestCreateGetCAPECCatalogMetaHandler_Error(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCAPECCatalogMetaHandler_Error", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{}
		handler := createGetCAPECCatalogMetaHandler(store, logger)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetCAPECCatalogMeta"}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "no catalog metadata" {
			t.Fatalf("expected error response, got %+v", resp)
		}
	})

}

func TestCreateGetCAPECCatalogMetaHandler_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCAPECCatalogMetaHandler_Success", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{meta: &capec.CAPECCatalogMeta{Version: "v1", Source: "src", ImportedAtUTC: 123}}
		handler := createGetCAPECCatalogMetaHandler(store, logger)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetCAPECCatalogMeta"}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}
		var decoded map[string]any
		if err := subprocess.UnmarshalFast(resp.Payload, &decoded); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if decoded["version"] != "v1" || decoded["imported_at"].(float64) != 123 {
			t.Fatalf("unexpected meta payload: %+v", decoded)
		}
	})

}

func TestCreateListCAPECsHandler_NormalizesParams(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateListCAPECsHandler_NormalizesParams", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{
			listItems: []capec.CAPECItemModel{{CAPECID: 1, Name: "n"}},
			listTotal: 1,
		}
		handler := createListCAPECsHandler(store, logger)

		payload, _ := subprocess.MarshalFast(map[string]int{"offset": -5, "limit": 5000})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCListCAPECs", Payload: payload}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}

		var decoded map[string]interface{}
		if err := subprocess.UnmarshalFast(resp.Payload, &decoded); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if decoded["offset"].(float64) != 0 || decoded["limit"].(float64) != 100 {
			t.Fatalf("expected normalized offset/limit, got %v/%v", decoded["offset"], decoded["limit"])
		}
	})

}

func TestCreateListCAPECsHandler_StoreError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateListCAPECsHandler_StoreError", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{listErr: errors.New("boom")}
		handler := createListCAPECsHandler(store, logger)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCListCAPECs"}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error == "" {
			t.Fatalf("expected error response, got %+v", resp)
		}
	})

}

func TestCreateGetCAPECByIDHandler_Validates(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCAPECByIDHandler_Validates", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{}
		handler := createGetCAPECByIDHandler(store, logger)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetCAPECByID"}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "capec_id is required" {
			t.Fatalf("expected validation error, got %+v", resp)
		}
	})

}

func TestCreateGetCAPECByIDHandler_StoreError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCAPECByIDHandler_StoreError", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{getErr: errors.New("nope")}
		handler := createGetCAPECByIDHandler(store, logger)

		payload, _ := subprocess.MarshalFast(map[string]string{"capec_id": "CAPEC-1"})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetCAPECByID", Payload: payload}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "CAPEC not found" {
			t.Fatalf("expected error response, got %+v", resp)
		}
	})

}

func TestCreateGetCAPECByIDHandler_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCAPECByIDHandler_Success", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		store := &stubCAPECStore{
			getItem:     &capec.CAPECItemModel{CAPECID: 2, Name: "Example", Summary: "<p>S</p>", Description: "<p>D</p>", Status: "Draft", Likelihood: "High", TypicalSeverity: "Medium"},
			related:     []capec.CAPECRelatedWeaknessModel{{CWEID: "CWE-79"}},
			examples:    []capec.CAPECExampleModel{{ExampleText: "<p>ex</p>"}},
			mitigations: []capec.CAPECMitigationModel{{MitigationText: "<p>mt</p>"}},
			references:  []capec.CAPECReferenceModel{{ExternalReference: "ref", URL: "http://example.com"}},
		}
		handler := createGetCAPECByIDHandler(store, logger)

		payload, _ := subprocess.MarshalFast(map[string]string{"capec_id": "2"})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetCAPECByID", Payload: payload}
		resp, err := handler(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}
		var decoded map[string]any
		if err := subprocess.UnmarshalFast(resp.Payload, &decoded); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if decoded["id"] != "CAPEC-2" || decoded["name"] != "Example" || decoded["likelihood"] != "High" {
			t.Fatalf("unexpected payload: %+v", decoded)
		}
		weak, _ := decoded["weaknesses"].([]any)
		if len(weak) != 1 || weak[0] != "CWE-79" {
			t.Fatalf("unexpected weaknesses: %+v", decoded["weaknesses"])
		}
		refs, _ := decoded["references"].([]any)
		if len(refs) != 1 {
			t.Fatalf("unexpected references: %+v", decoded["references"])
		}
	})

}

// testWriter adapts testing.T to io.Writer for logger output suppression.
type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Logf("%s", string(p))
	return len(p), nil
}
