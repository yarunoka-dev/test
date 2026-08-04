package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// DefaultTimeout cuts a hanging adapter per case, so one stuck process
// cannot stall a whole CI run. Generous on purpose: a slow interpreter
// start-up is normal, an adapter that needs a minute for one case is not.
const DefaultTimeout = 60 * time.Second

// Adapter is the implementer-supplied command, started once per case
// with the request on stdin and the answer read from stdout.
type Adapter struct {
	Argv    []string
	Timeout time.Duration // zero means DefaultTimeout
}

// Ask runs one round: start, write the request, read the answer. An
// error is an adapter breakage (crash, hang, or non-JSON output) —
// infrastructure trouble to report apart from test results, never a FAIL.
func (a Adapter) Ask(req Request) (Response, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return Response{}, fmt.Errorf("marshaling the request (a kit-side error): %w", err)
	}

	timeout := a.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.Argv[0], a.Argv[1:]...)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Killing the adapter is not enough to unblock Wait: a grandchild
	// that inherited the stdout pipe keeps it open past the kill, and
	// Wait would sit on the pipe's EOF until that survivor exits. The
	// delay abandons the pipes and lets Wait return.
	cmd.WaitDelay = time.Second

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return Response{}, fmt.Errorf("the adapter did not answer within %s", timeout)
		}
		return Response{}, fmt.Errorf("the adapter exited abnormally: %w%s", err, diagnosis(&stderr))
	}

	var answer Response
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &answer); err != nil {
		return Response{}, fmt.Errorf("the adapter's output is not a JSON response: %q%s", stdout.String(), diagnosis(&stderr))
	}
	return answer, nil
}

func diagnosis(stderr *bytes.Buffer) string {
	if stderr.Len() == 0 {
		return ""
	}
	return fmt.Sprintf(" (stderr: %s)", bytes.TrimSpace(stderr.Bytes()))
}
