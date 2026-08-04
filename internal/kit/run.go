package kit

// Mode selects what conformance is being checked. Eval and emit are
// independent claims ("kit vX.Y eval passed" / "emit passed"); all is
// the convenience of running both, not a third claim.
type Mode string

const (
	ModeEval Mode = "eval"
	ModeEmit Mode = "emit"
	ModeAll  Mode = "all"
)

// Status is the judgment of one case in one mode. Broken means the
// adapter itself misbehaved (crash, hang, non-JSON output) — the case
// was not checkable, which is different from checked-and-failed.
type Status int

const (
	StatusPass Status = iota
	StatusFail
	StatusBroken
)

// Outcome is one case judged in one mode, with the mismatch detail for
// FAIL reporting.
type Outcome struct {
	Case   Case
	Mode   Mode
	Status Status
	Detail string
}

// asker is what Run needs from an adapter; Adapter satisfies it with a
// real child process.
type asker interface {
	Ask(Request) (Response, error)
}

// Run judges every case in the selected mode(s), eval first. Emit
// derives from the same cases rather than a separate set: a valid case
// round-trips its own document (round-tripping is the identity, so the
// authored spelling is the expectation), an invalid case expects
// invalid. Authoring a valid case is authoring its emit expectation.
func Run(cases []Case, adapter asker, mode Mode) []Outcome {
	var outcomes []Outcome
	if mode == ModeEval || mode == ModeAll {
		for _, c := range cases {
			outcomes = append(outcomes, judge(c, ModeEval, adapter, c.Request))
		}
	}
	if mode == ModeEmit || mode == ModeAll {
		for _, c := range cases {
			outcomes = append(outcomes, judge(c, ModeEmit, adapter, Request{Action: "emit", Document: c.Request.Document}))
		}
	}
	return outcomes
}

func judge(c Case, mode Mode, adapter asker, req Request) Outcome {
	answer, err := adapter.Ask(req)
	if err != nil {
		return Outcome{Case: c, Mode: mode, Status: StatusBroken, Detail: err.Error()}
	}

	var ok bool
	var detail string
	if mode == ModeEval {
		ok, detail = CompareEval(c.Expected, answer)
	} else {
		ok, detail = CompareEmit(c.Request.Document, c.Expected.Invalid, answer)
	}
	if !ok {
		return Outcome{Case: c, Mode: mode, Status: StatusFail, Detail: detail}
	}
	return Outcome{Case: c, Mode: mode, Status: StatusPass}
}
