package kit

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

// The wire form of a request carries the document as a JSON string
// holding the authored text verbatim. An embedded JSON value would be
// decoded by the adapter's own JSON parser, which collapses duplicate
// member names before the implementation ever sees them; the string
// survives that decoding untouched.
func TestRequestWireFormCarriesTheDocumentAsAString(t *testing.T) {
	authored := `{"version": "1.1", "timezone": "Asia/Tokyo", "timezone": "UTC"}`
	req := Request{Action: "eval", Document: json.RawMessage(authored)}

	wire, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var sent struct {
		Action   string `json:"action"`
		Document string `json:"document"`
	}
	if err := json.Unmarshal(wire, &sent); err != nil {
		t.Fatalf("the wire form does not parse: %v", err)
	}
	if sent.Action != "eval" {
		t.Errorf("action: %q", sent.Action)
	}
	if sent.Document != authored {
		t.Errorf("the document must ride verbatim as a string:\nwant %s\ngot  %s", authored, sent.Document)
	}
}

// A case file still authors the document as the JSON value it is; the
// loaded case keeps the authored bytes, duplicate member names included,
// so the stringifying at send time loses nothing.
func TestLoadCasesKeepsTheAuthoredDocumentBytes(t *testing.T) {
	fsys := fstest.MapFS{
		"document/invalid-duplicate-member.json": &fstest.MapFile{Data: []byte(`{
			"description": "Duplicate member names are rejected",
			"spec": "Document model - objects reject duplicate member names",
			"request": {
				"action": "eval",
				"document": {"version": "1.1", "timezone": "Asia/Tokyo", "timezone": "UTC"},
				"query": {"type": "point", "at": "2026-07-27T10:00:00+09:00"}
			},
			"response": {"invalid": true}
		}`)},
	}

	cases, err := LoadCases(fsys)
	if err != nil {
		t.Fatalf("LoadCases: %v", err)
	}
	if got := string(cases[0].Request.Document); got != `{"version": "1.1", "timezone": "Asia/Tokyo", "timezone": "UTC"}` {
		t.Errorf("the loaded document must keep the authored bytes, got %s", got)
	}
}
