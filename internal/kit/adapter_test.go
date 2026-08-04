package kit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// These tests exercise real child processes through sh, so they run on
// POSIX environments (the CI runner and the development machines).

func request() Request {
	return Request{Action: "eval", Document: json.RawMessage(`{}`)}
}

func TestAskWritesTheRequestAndReadsTheAnswer(t *testing.T) {
	// The script proves both directions: it drains stdin into the answer
	// it writes, so a wrong request would fail the assertion below.
	a := Adapter{Argv: []string{"sh", "-c", `echo "{\"result\": $(cat | wc -c | tr -d ' ') }"`}}

	req := request()
	answer, err := a.Ask(req)
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}

	sent, _ := json.Marshal(req)
	var got int
	if err := json.Unmarshal(answer.Result, &got); err != nil || got != len(sent) {
		t.Errorf("the adapter saw %v bytes on stdin, the runner sent %d", answer.Result, len(sent))
	}
}

func TestAskReportsANonZeroExitAsAdapterFailure(t *testing.T) {
	a := Adapter{Argv: []string{"sh", "-c", "echo boom >&2; exit 3"}}

	_, err := a.Ask(request())
	if err == nil {
		t.Fatal("expected an adapter failure")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("the failure must carry stderr for diagnosis: %v", err)
	}
}

func TestAskReportsNonJSONOutputAsAdapterFailure(t *testing.T) {
	a := Adapter{Argv: []string{"sh", "-c", "cat > /dev/null; echo 'PHP Warning: something'"}}

	if _, err := a.Ask(request()); err == nil {
		t.Fatal("expected an adapter failure for non-JSON output")
	}
}

func TestAskReportsAHangingAdapterAsAdapterFailure(t *testing.T) {
	a := Adapter{Argv: []string{"sh", "-c", "sleep 10"}, Timeout: 50 * time.Millisecond}

	start := time.Now()
	_, err := a.Ask(request())
	if err == nil {
		t.Fatal("expected an adapter failure for a hang")
	}
	if time.Since(start) > 5*time.Second {
		t.Error("the timeout did not cut the wait")
	}
}
