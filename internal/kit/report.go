package kit

import (
	"encoding/json"
	"fmt"
	"io"
)

// Report writes every failure and breakage with its authored context:
// what the case checks, under which normative text, the input that was
// sent, and how the answer differed. The detection is the kit's job and
// the fixing is the implementation's, so the report aims to hand the
// implementer everything they need. Passing cases stay quiet.
func Report(w io.Writer, outcomes []Outcome) {
	for _, o := range outcomes {
		switch o.Status {
		case StatusFail:
			fmt.Fprintf(w, "FAIL %s (%s)\n", o.Case.Name, o.Mode)
			fmt.Fprintf(w, "  checks: %s\n", o.Case.Description)
			fmt.Fprintf(w, "  spec:   %s\n", o.Case.Spec)
			fmt.Fprintf(w, "  sent:   %s\n", sentRequest(o))
			fmt.Fprintf(w, "  %s\n\n", o.Detail)
		case StatusBroken:
			fmt.Fprintf(w, "BROKEN %s (%s)\n", o.Case.Name, o.Mode)
			fmt.Fprintf(w, "  %s\n\n", o.Detail)
		}
	}
}

// Summary writes the pasteable claim: kit version, targeted spec
// version, case count, and the per-mode tally. Breakage appears in the
// tally because a run with uncheckable cases is not a complete claim.
func Summary(w io.Writer, outcomes []Outcome) {
	modes := []Mode{ModeEval, ModeEmit}
	perMode := map[Mode][3]int{}
	names := map[string]bool{}
	for _, o := range outcomes {
		tally := perMode[o.Mode]
		tally[o.Status]++
		perMode[o.Mode] = tally
		names[o.Case.Name] = true
	}

	fmt.Fprintf(w, "yarunoka-test %s (spec %s), %d cases\n", KitVersion, SpecVersion, len(names))
	for _, mode := range modes {
		tally, ran := perMode[mode]
		if !ran {
			continue
		}
		fmt.Fprintf(w, "%s: %d passed, %d failed", mode, tally[StatusPass], tally[StatusFail])
		if tally[StatusBroken] > 0 {
			fmt.Fprintf(w, ", %d broken", tally[StatusBroken])
		}
		fmt.Fprintln(w)
	}
}

// AllPassed is the exit criterion: anything but a full run of passes
// (a failure, or a case the breakage made uncheckable) is a red run.
func AllPassed(outcomes []Outcome) bool {
	for _, o := range outcomes {
		if o.Status != StatusPass {
			return false
		}
	}
	return true
}

func sentRequest(o Outcome) string {
	req := o.Case.Request
	if o.Mode == ModeEmit {
		req = Request{Action: "emit", Document: o.Case.Request.Document}
	}
	sent, err := json.Marshal(req)
	if err != nil {
		return fmt.Sprintf("(unmarshalable: %v)", err)
	}
	return string(sent)
}
