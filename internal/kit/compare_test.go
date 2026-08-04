package kit

import (
	"encoding/json"
	"testing"
)

func resp(s string) Response {
	var r Response
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		panic(err)
	}
	return r
}

// The judgment queries answer a boolean; the comparison is equality.
func TestCompareEvalJudgsBooleans(t *testing.T) {
	if ok, _ := CompareEval(resp(`{"result": true}`), resp(`{"result": true}`)); !ok {
		t.Error("equal booleans must pass")
	}
	if ok, _ := CompareEval(resp(`{"result": true}`), resp(`{"result": false}`)); ok {
		t.Error("differing booleans must fail")
	}
}

// An enumeration answers dates (all-day) and instants (timed). Instants
// compare as moments, so the adapter's offset spelling is free; dates
// compare as the calendar days they are. Order and length are part of
// the answer.
func TestCompareEvalJudgsEnumerations(t *testing.T) {
	expected := resp(`{"result": ["2026-07-28", "2026-07-28T00:00:00+09:00"]}`)

	if ok, _ := CompareEval(expected, resp(`{"result": ["2026-07-28", "2026-07-28T00:00:00+09:00"]}`)); !ok {
		t.Error("an identical enumeration must pass")
	}
	// 2026-07-27T15:00:00Z is the same moment as 2026-07-28T00:00:00+09:00.
	if ok, _ := CompareEval(expected, resp(`{"result": ["2026-07-28", "2026-07-27T15:00:00Z"]}`)); !ok {
		t.Error("a differently spelled identical moment must pass")
	}
	if ok, _ := CompareEval(expected, resp(`{"result": ["2026-07-28T00:00:00+09:00", "2026-07-28"]}`)); ok {
		t.Error("a reordered enumeration must fail")
	}
	if ok, _ := CompareEval(expected, resp(`{"result": ["2026-07-28"]}`)); ok {
		t.Error("a shorter enumeration must fail")
	}
	// The two kinds never merge: a date is not the timed 00:00 of that day.
	if ok, _ := CompareEval(resp(`{"result": ["2026-07-28"]}`), resp(`{"result": ["2026-07-28T00:00:00+09:00"]}`)); ok {
		t.Error("a timed occurrence must not pass for an all-day expectation")
	}
}

// Invalid is a binary answer: the only passing response to an invalid
// case is the invalid flag, and answering invalid to a valid case fails.
func TestCompareEvalJudgsInvalid(t *testing.T) {
	if ok, _ := CompareEval(resp(`{"invalid": true}`), resp(`{"invalid": true}`)); !ok {
		t.Error("invalid expected, invalid answered: must pass")
	}
	if ok, _ := CompareEval(resp(`{"invalid": true}`), resp(`{"result": true}`)); ok {
		t.Error("invalid expected, result answered: must fail")
	}
	if ok, _ := CompareEval(resp(`{"result": true}`), resp(`{"invalid": true}`)); ok {
		t.Error("result expected, invalid answered: must fail")
	}
}

// The emitted document compares structurally: JSON key order and
// whitespace carry no meaning, so they must not fail the round-trip.
func TestCompareEmitComparesStructurally(t *testing.T) {
	authored := json.RawMessage(`{"version": "1.0", "timezone": "Asia/Tokyo", "schedules": [{"days": ["mon"], "times": ["10:00"]}]}`)

	reordered := resp(`{"document": {"timezone": "Asia/Tokyo", "schedules": [{"times": ["10:00"], "days": ["mon"]}], "version": "1.0"}}`)
	if ok, _ := CompareEmit(authored, false, reordered); !ok {
		t.Error("a reordered but structurally identical document must pass")
	}

	changed := resp(`{"document": {"version": "1.0", "timezone": "Asia/Tokyo", "schedules": [{"days": ["tue"], "times": ["10:00"]}]}}`)
	if ok, _ := CompareEmit(authored, false, changed); ok {
		t.Error("a structurally different document must fail")
	}

	if ok, _ := CompareEmit(authored, false, resp(`{"invalid": true}`)); ok {
		t.Error("invalid answered for a valid document must fail")
	}
	if ok, _ := CompareEmit(authored, true, resp(`{"invalid": true}`)); !ok {
		t.Error("invalid answered for an invalid document must pass")
	}
}
