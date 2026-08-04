package kit

import (
	"testing"
	"testing/fstest"
)

// A case file is one JSON object holding the request, the expected
// response, and the authoring metadata used for FAIL reporting. The path
// under the root names the case.
func TestLoadCasesReadsEveryJSONFileWithItsPathAsName(t *testing.T) {
	fsys := fstest.MapFS{
		"point/timed-exact-instant.json": &fstest.MapFile{Data: []byte(`{
			"description": "A timed occurrence matches its exact instant",
			"spec": "Evaluation model - judgment at a point",
			"request": {
				"action": "eval",
				"document": {"version": "1.0", "timezone": "Asia/Tokyo", "schedules": [{"days": ["mon"], "times": ["10:00"]}]},
				"query": {"type": "point", "at": "2026-07-27T10:00:00+09:00"}
			},
			"response": {"result": true}
		}`)},
		"invalid/unknown-key.json": &fstest.MapFile{Data: []byte(`{
			"description": "Unknown document keys are rejected",
			"spec": "Document model - closed key set",
			"request": {
				"action": "eval",
				"document": {"version": "1.0", "extra": 1},
				"query": {"type": "point", "at": "2026-07-27T10:00:00+09:00"}
			},
			"response": {"invalid": true}
		}`)},
	}

	cases, err := LoadCases(fsys)
	if err != nil {
		t.Fatalf("LoadCases: %v", err)
	}

	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	// Sorted by name so runs are deterministic.
	if cases[0].Name != "invalid/unknown-key" || cases[1].Name != "point/timed-exact-instant" {
		t.Fatalf("unexpected names: %q, %q", cases[0].Name, cases[1].Name)
	}

	c := cases[1]
	if c.Description != "A timed occurrence matches its exact instant" {
		t.Errorf("description: %q", c.Description)
	}
	if c.Spec != "Evaluation model - judgment at a point" {
		t.Errorf("spec: %q", c.Spec)
	}
	if c.Request.Action != "eval" {
		t.Errorf("action: %q", c.Request.Action)
	}
	if c.Request.Query == nil || c.Request.Query.Type != "point" || c.Request.Query.At != "2026-07-27T10:00:00+09:00" {
		t.Errorf("query: %+v", c.Request.Query)
	}
	if c.Expected.Invalid {
		t.Error("a valid case must not be expected invalid")
	}
}

// The authored expectation decides how the case is judged, so a case
// carrying neither a result, nor invalid, is a kit-side authoring error
// and must fail loading, not run.
func TestLoadCasesRejectsACaseWithoutAnExpectation(t *testing.T) {
	fsys := fstest.MapFS{
		"broken.json": &fstest.MapFile{Data: []byte(`{
			"description": "no expectation",
			"spec": "nowhere",
			"request": {"action": "eval", "document": {}},
			"response": {}
		}`)},
	}

	if _, err := LoadCases(fsys); err == nil {
		t.Fatal("expected an error for a case without an expectation")
	}
}
