package kit

import (
	"encoding/json"
	"strings"
	"testing"
)

func outcomes() []Outcome {
	failing := Case{
		Name:        "point/timed-exact-instant",
		Description: "A timed occurrence matches its exact instant",
		Spec:        "Evaluation model - judgment at a point",
		Request:     Request{Action: "eval", Document: json.RawMessage(`{"version": "1.0"}`), Query: &Query{Type: "point", At: "2026-07-27T10:00:00+09:00"}},
		Expected:    Response{Result: json.RawMessage(`true`)},
	}
	return []Outcome{
		{Case: Case{Name: "invalid/unknown-key"}, Mode: ModeEval, Status: StatusPass},
		{Case: failing, Mode: ModeEval, Status: StatusFail, Detail: "expected true, got false"},
		{Case: Case{Name: "period/boundary"}, Mode: ModeEmit, Status: StatusBroken, Detail: "the adapter exited abnormally"},
	}
}

// A FAIL shows what the implementer needs to fix their side: the case,
// what it checks and under which normative text, the input that was
// sent, and how the answer differed. Passing cases stay quiet.
func TestReportShowsFailuresWithTheirAuthoredContext(t *testing.T) {
	var b strings.Builder
	Report(&b, outcomes())
	out := b.String()

	for _, want := range []string{
		"FAIL point/timed-exact-instant (eval)",
		"A timed occurrence matches its exact instant",
		"Evaluation model - judgment at a point",
		`"at":"2026-07-27T10:00:00+09:00"`,
		"expected true, got false",
		"BROKEN period/boundary (emit)",
		"the adapter exited abnormally",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the report must contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "invalid/unknown-key") {
		t.Error("passing cases must stay quiet")
	}
}

// The summary is made to be pasted into an implementations-list PR:
// kit version, targeted spec version, case count, and the per-mode
// tally, each on its own line.
func TestSummaryCarriesTheVersionsAndPerModeTallies(t *testing.T) {
	var b strings.Builder
	Summary(&b, outcomes())
	out := b.String()

	for _, want := range []string{
		"yarunoka-test " + KitVersion,
		"spec " + SpecVersion,
		"eval: 1 passed, 1 failed",
		"emit: 0 passed, 0 failed, 1 broken",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the summary must contain %q, got:\n%s", want, out)
		}
	}
}

func TestAllPassedIsTheExitCriterion(t *testing.T) {
	if AllPassed(outcomes()) {
		t.Error("failures and breakage must not count as passed")
	}
	if !AllPassed([]Outcome{{Status: StatusPass}}) {
		t.Error("a run of passes must count as passed")
	}
}
