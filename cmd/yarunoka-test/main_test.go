package main

import "testing"

// The mode is required and explicit: a defaulted "all" would make a
// copied CI line run emit against eval-only implementations, and the
// command line would stop saying what was checked.
func TestParseArgsRequiresAModeAndAnAdapterCommand(t *testing.T) {
	if _, _, err := parseArgs([]string{"eval", "php", "adapter.php"}); err != nil {
		t.Errorf("a mode with an adapter command must parse: %v", err)
	}
	if _, _, err := parseArgs([]string{"php", "adapter.php"}); err == nil {
		t.Error("a missing mode must be an error, not a default")
	}
	if _, _, err := parseArgs([]string{"all"}); err == nil {
		t.Error("a missing adapter command must be an error")
	}
	if _, _, err := parseArgs(nil); err == nil {
		t.Error("no arguments must be an error")
	}
}

func TestParseArgsKeepsTheAdapterCommandVerbatim(t *testing.T) {
	mode, argv, err := parseArgs([]string{"emit", "php", "-d", "memory_limit=1G", "adapter.php"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if string(mode) != "emit" {
		t.Errorf("mode: %q", mode)
	}
	if len(argv) != 4 || argv[0] != "php" || argv[3] != "adapter.php" {
		t.Errorf("argv must pass through untouched: %v", argv)
	}
}
