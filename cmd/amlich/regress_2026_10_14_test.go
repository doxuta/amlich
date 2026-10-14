package main

import (
	"strings"
	"testing"

	"github.com/doxuta/amlich"
)

// The package's sentinel errors already carry the "amlich: " prefix and die()
// added another, so `amlich convert 5000-01-01` printed
// "amlich: amlich: invalid solar date". The prefix must appear once — and a
// well-formed date outside the window is not "invalid", it is unsupported, so
// the message names the window as `holidays` already does.
func TestRegress20261014_CLIErrorPrefixOnce(t *testing.T) {
	bin := buildCLI(t)
	for _, args := range [][]string{
		{"convert", "5000-01-01"},
		{"convert", "1199-12-31", "-kr"},
		{"convert", "2026-02-30"}, // time.Parse error: no package prefix of its own
		{"holidays", "5000"},
		{"canchi", "abc"},
	} {
		_, errOut, code := run(t, bin, args...)
		if code != 1 {
			t.Errorf("amlich %s: exit=%d, want 1 (stderr %q)", strings.Join(args, " "), code, errOut)
		}
		if !strings.HasPrefix(errOut, "amlich: ") || strings.Contains(errOut, "amlich: amlich:") {
			t.Errorf("amlich %s: stderr %q; want exactly one \"amlich: \" prefix", strings.Join(args, " "), errOut)
		}
	}
	_, errOut, _ := run(t, bin, "convert", "5000-01-01")
	if !strings.Contains(errOut, "5000") || !strings.Contains(errOut, "1200-3000") {
		t.Errorf("amlich convert 5000-01-01: stderr %q; want the year and the window %d-%d", errOut, amlich.MinYear, amlich.MaxYear)
	}
}
