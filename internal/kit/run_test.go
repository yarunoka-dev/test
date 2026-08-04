package kit

import (
	"encoding/json"
	"errors"
	"testing"
)

// A scripted stand-in for an adapter: it records what the runner sends
// and answers from a fixed table keyed by action.
type fakeAdapter struct {
	requests []Request
	answers  map[string]Response
	err      error
}

func (f *fakeAdapter) Ask(req Request) (Response, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return Response{}, f.err
	}
	return f.answers[req.Action], nil
}

func loaded(t *testing.T) []Case {
	t.Helper()
	return []Case{
		{
			Name:        "invalid/unknown-key",
			Description: "Unknown document keys are rejected",
			Request:     Request{Action: "eval", Document: json.RawMessage(`{"version": "1.0", "extra": 1}`), Query: &Query{Type: "point", At: "2026-07-27T10:00:00+09:00"}},
			Expected:    Response{Invalid: true},
		},
		{
			Name:        "point/timed-exact-instant",
			Description: "A timed occurrence matches its exact instant",
			Request:     Request{Action: "eval", Document: json.RawMessage(`{"version": "1.0"}`), Query: &Query{Type: "point", At: "2026-07-27T10:00:00+09:00"}, Bindings: map[string][]string{"company-closures": {"2026-08-05"}}},
			Expected:    Response{Result: json.RawMessage(`true`)},
		},
	}
}

func TestRunEvalSendsTheAuthoredRequestsAndJudges(t *testing.T) {
	adapter := &fakeAdapter{answers: map[string]Response{
		"eval": {Result: json.RawMessage(`true`)},
	}}

	outcomes := Run(loaded(t), adapter, ModeEval)

	if len(outcomes) != 2 {
		t.Fatalf("expected 2 outcomes, got %d", len(outcomes))
	}
	// The invalid case got a result answer, so it fails; the point case
	// got the right boolean, so it passes.
	if outcomes[0].Status != StatusFail || outcomes[1].Status != StatusPass {
		t.Fatalf("unexpected statuses: %+v", outcomes)
	}
	for _, req := range adapter.requests {
		if req.Action != "eval" || req.Query == nil {
			t.Errorf("eval mode must send the authored eval request: %+v", req)
		}
	}
}

// Emit derives from the same cases: a valid case round-trips its own
// document (the expectation is the authored spelling), an invalid case
// expects invalid. No separate emit cases exist.
func TestRunEmitDerivesRoundTripsFromTheSameCases(t *testing.T) {
	cases := loaded(t)
	adapter := &fakeAdapter{answers: map[string]Response{
		"emit": {Invalid: true},
	}}

	outcomes := Run(cases, adapter, ModeEmit)

	if len(outcomes) != 2 {
		t.Fatalf("expected 2 outcomes, got %d", len(outcomes))
	}
	// The invalid case expects invalid: pass. The valid case got
	// invalid instead of its document: fail.
	if outcomes[0].Status != StatusPass || outcomes[1].Status != StatusFail {
		t.Fatalf("unexpected statuses: %+v", outcomes)
	}
	for i, req := range adapter.requests {
		if req.Action != "emit" {
			t.Errorf("emit mode must send emit requests: %+v", req)
		}
		if req.Query != nil {
			t.Errorf("an emit request carries no query: %+v", req)
		}
		// Bindings ride along: parsing a document that declares resolvers
		// needs them (an unbound declared name is a validation error), and
		// emit starts with a parse.
		if want := cases[i].Request.Bindings; len(want) != len(req.Bindings) {
			t.Errorf("an emit request must keep the case's bindings: %+v", req)
		}
	}
}

func TestRunAllRunsBothModes(t *testing.T) {
	adapter := &fakeAdapter{answers: map[string]Response{}}

	outcomes := Run(loaded(t), adapter, ModeAll)

	if len(outcomes) != 4 {
		t.Fatalf("expected 2 cases x 2 modes = 4 outcomes, got %d", len(outcomes))
	}
}

func TestRunClassifiesAdapterBreakageApartFromFailures(t *testing.T) {
	adapter := &fakeAdapter{err: errors.New("the adapter exited abnormally")}

	outcomes := Run(loaded(t), adapter, ModeEval)

	for _, o := range outcomes {
		if o.Status != StatusBroken {
			t.Errorf("breakage must be its own status, got %+v", o)
		}
	}
}
