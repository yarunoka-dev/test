package kit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// CompareEval judges an adapter's answer to an eval request against the
// authored expectation. The detail explains a mismatch for FAIL
// reporting; it is empty on a pass.
func CompareEval(expected, actual Response) (bool, string) {
	if expected.Invalid {
		if actual.Invalid {
			return true, ""
		}
		return false, fmt.Sprintf("expected invalid, got %s", describe(actual))
	}
	if actual.Invalid {
		return false, "expected a result, got invalid"
	}
	if actual.Result == nil {
		return false, fmt.Sprintf("expected a result, got %s", describe(actual))
	}

	var expectedList, actualList []string
	if err := json.Unmarshal(expected.Result, &expectedList); err != nil {
		// Not a list, so the expectation is a judgment boolean.
		return compareBooleans(expected.Result, actual.Result)
	}
	if err := json.Unmarshal(actual.Result, &actualList); err != nil {
		return false, fmt.Sprintf("expected an enumeration, got %s", actual.Result)
	}
	return compareEnumerations(expectedList, actualList)
}

// CompareEmit judges an adapter's answer to an emit request. The
// expectation is the authored document itself (round-tripping is the
// identity), or invalid when the authored document is invalid. The
// comparison is structural: JSON key order and whitespace carry no
// meaning, so encoder differences between languages never fail it.
func CompareEmit(authored json.RawMessage, authoredInvalid bool, actual Response) (bool, string) {
	if authoredInvalid {
		if actual.Invalid {
			return true, ""
		}
		return false, fmt.Sprintf("expected invalid, got %s", describe(actual))
	}
	if actual.Invalid {
		return false, "expected the round-tripped document, got invalid"
	}
	if actual.Document == nil {
		return false, fmt.Sprintf("expected a document, got %s", describe(actual))
	}

	var want, got any
	if err := json.Unmarshal(authored, &want); err != nil {
		return false, fmt.Sprintf("the authored document does not parse (a kit-side error): %v", err)
	}
	if err := json.Unmarshal(actual.Document, &got); err != nil {
		return false, fmt.Sprintf("the answered document does not parse: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		return false, fmt.Sprintf("the round-tripped document differs from the authored spelling: got %s", actual.Document)
	}
	return true, ""
}

func compareBooleans(expected, actual json.RawMessage) (bool, string) {
	var want, got bool
	if err := json.Unmarshal(expected, &want); err != nil {
		return false, fmt.Sprintf("the expected result is neither a boolean nor an enumeration (a kit-side error): %s", expected)
	}
	if err := json.Unmarshal(actual, &got); err != nil {
		return false, fmt.Sprintf("expected a boolean, got %s", actual)
	}
	if want != got {
		return false, fmt.Sprintf("expected %t, got %t", want, got)
	}
	return true, ""
}

// An enumeration entry is a date (all-day) or an instant (timed); the
// spelling tells them apart. Instants compare as moments so the offset
// spelling is the adapter's choice; dates compare as calendar days.
// The two kinds never merge, and order is part of the answer.
func compareEnumerations(expected, actual []string) (bool, string) {
	if len(expected) != len(actual) {
		return false, fmt.Sprintf("expected %d occurrences, got %d: %v", len(expected), len(actual), actual)
	}
	for i := range expected {
		want, wantInstant, err := parseOccurrence(expected[i])
		if err != nil {
			return false, fmt.Sprintf("the expected occurrence %q does not parse (a kit-side error): %v", expected[i], err)
		}
		got, gotInstant, err := parseOccurrence(actual[i])
		if err != nil {
			return false, fmt.Sprintf("occurrence %d %q does not parse: %v", i, actual[i], err)
		}
		if wantInstant != gotInstant {
			return false, fmt.Sprintf("occurrence %d: expected %q, got %q (an all-day occurrence and a timed occurrence never merge)", i, expected[i], actual[i])
		}
		if wantInstant {
			if !want.Equal(got) {
				return false, fmt.Sprintf("occurrence %d: expected the moment %q, got %q", i, expected[i], actual[i])
			}
		} else if expected[i] != actual[i] {
			return false, fmt.Sprintf("occurrence %d: expected the day %q, got %q", i, expected[i], actual[i])
		}
	}
	return true, ""
}

func parseOccurrence(s string) (t time.Time, instant bool, err error) {
	if len(s) == len("2006-01-02") {
		_, err = time.Parse("2006-01-02", s)
		return time.Time{}, false, err
	}
	t, err = time.Parse(time.RFC3339, s)
	return t, true, err
}

func describe(r Response) string {
	switch {
	case r.Result != nil:
		return fmt.Sprintf("the result %s", r.Result)
	case r.Document != nil:
		return "a document"
	default:
		return "an empty response"
	}
}
