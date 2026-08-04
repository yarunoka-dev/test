// Command yarunoka-test runs the embedded conformance cases against an
// implementation's adapter and reports the outcome.
//
//	yarunoka-test {eval|emit|all} <adapter command> [args...]
package main

import (
	"fmt"
	"os"

	testkit "github.com/yarunoka-dev/test"
	"github.com/yarunoka-dev/test/internal/kit"
)

func main() {
	mode, argv, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
		usage(os.Stderr)
		os.Exit(2)
	}

	cases, err := kit.LoadCases(testkit.Cases())
	if err != nil {
		fmt.Fprintf(os.Stderr, "loading the embedded cases (a kit-side defect): %v\n", err)
		os.Exit(2)
	}

	outcomes := kit.Run(cases, kit.Adapter{Argv: argv}, mode)
	kit.Report(os.Stdout, outcomes)
	kit.Summary(os.Stdout, outcomes)

	if !kit.AllPassed(outcomes) {
		os.Exit(1)
	}
}

// parseArgs reads the required mode and the adapter command. The mode
// has no default: a defaulted "all" would make a copied CI line run
// emit against eval-only implementations, and the command line would
// stop saying what was checked.
func parseArgs(args []string) (kit.Mode, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("a mode is required")
	}
	mode := kit.Mode(args[0])
	if mode != kit.ModeEval && mode != kit.ModeEmit && mode != kit.ModeAll {
		return "", nil, fmt.Errorf("unknown mode %q", args[0])
	}
	if len(args) < 2 {
		return "", nil, fmt.Errorf("an adapter command is required")
	}
	return mode, args[1:], nil
}

func usage(w *os.File) {
	fmt.Fprintf(w, `usage: yarunoka-test {eval|emit|all} <adapter command> [args...]

Runs the embedded conformance cases against the adapter and exits
non-zero unless every case passes.

  eval   the three queries (point, period, enumeration)
  emit   the round-trip spelling check
  all    both

The adapter command is started once per case, receives one request as
JSON on stdin, and answers as JSON on stdout.

example: yarunoka-test eval php adapter.php
`)
}
