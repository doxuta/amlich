package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Regression tests for the 2026-10-12 audit, run against the built binary.

func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "amlich")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func run(t *testing.T, bin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		code = ee.ExitCode()
	default:
		t.Fatalf("%s %v: %v", bin, args, err)
	}
	return o.String(), e.String(), code
}

// `amlich holidays 5000` and `amlich divergence 5000 6000` printed nothing and
// exited 0: Holidays returns nil and Divergence clamps to an empty range for
// out-of-window years, and the CLI rendered both as success. Every other
// out-of-window input (convert) already exits 1 with a message.
func TestRegress20261012_CLIOutOfWindowExitsNonZero(t *testing.T) {
	bin := buildCLI(t)
	for _, args := range [][]string{
		{"holidays", "5000"},
		{"holidays", "1199", "-kr"},
		{"divergence", "5000", "6000"},
		{"divergence", "2031", "2026"},
	} {
		out, errOut, code := run(t, bin, args...)
		if code == 0 || out != "" || errOut == "" {
			t.Errorf("amlich %s: exit=%d stdout=%q stderr=%q; want exit!=0, empty stdout, a message on stderr",
				strings.Join(args, " "), code, out, errOut)
		}
	}
	if out, _, code := run(t, bin, "holidays", "2025"); code != 0 || !strings.Contains(out, "2025-01-29") {
		t.Errorf("amlich holidays 2025: exit=%d out=%q", code, out)
	}
}

// README.md opens with the exact output of `amlich divergence 2026 2031`;
// that block is a spec and the binary must reproduce it line for line.
func TestRegress20261012_ReadmeDivergenceExample(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	const cmd = "$ amlich divergence 2026 2031\n"
	i := strings.Index(string(readme), cmd)
	if i < 0 {
		t.Fatalf("README.md no longer contains %q", strings.TrimSpace(cmd))
	}
	rest := string(readme)[i+len(cmd):]
	want := rest[:strings.Index(rest, "```")]
	got, _, code := run(t, buildCLI(t), "divergence", "2026", "2031")
	if code != 0 || got != want {
		t.Errorf("amlich divergence 2026 2031 (exit %d) differs from the README example.\n--- got\n%s--- README\n%s", code, got, want)
	}
}
