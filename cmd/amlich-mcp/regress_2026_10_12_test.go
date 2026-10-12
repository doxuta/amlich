package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Regression tests for the 2026-10-12 audit. The server is exercised end to
// end — built, started, and spoken to over stdio — because the tool handlers
// live inside main and the stdio framing is part of what is under test.

func startServer(t *testing.T) (context.Context, *mcp.ClientSession) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "amlich-mcp")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	client := mcp.NewClient(&mcp.Implementation{Name: "regress", Version: "0"}, nil)
	sess, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.Command(bin)}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { sess.Close() })
	return ctx, sess
}

func call(t *testing.T, ctx context.Context, sess *mcp.ClientSession, name string, args map[string]any) (bool, string) {
	t.Helper()
	res, err := sess.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s %v: protocol error (server must answer, not fail the call): %v", name, args, err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return res.IsError, b.String()
}

// A year outside the supported window must come back as a tool error, not as
// an empty success. Holidays returns nil for such a year and the handler
// rendered that nil as "" with isError=false, so an agent asking for the
// holidays of year 5000 received an empty, authoritative-looking answer.
func TestRegress20261012_MCPHolidaysOutOfWindowIsToolError(t *testing.T) {
	ctx, sess := startServer(t)
	for _, year := range []int{5000, 1199, 3001, -1, 0} {
		isErr, txt := call(t, ctx, sess, "holidays_in_year", map[string]any{"year": year})
		if !isErr {
			t.Errorf("holidays_in_year year=%d: isError=false, text=%q; want a tool error", year, txt)
		}
	}
	// In-window years keep working.
	if isErr, txt := call(t, ctx, sess, "holidays_in_year", map[string]any{"year": 2025}); isErr || !strings.Contains(txt, "2025-01-29") {
		t.Errorf("holidays_in_year 2025: isError=%v text=%q", isErr, txt)
	}
}

// Divergence clamps its range to the window, so a range entirely outside it
// is empty — and the handler reported that emptiness as the factual claim
// "No VN/KR divergence in lunar years 5000-6000", which is not something the
// engine knows. Such a range must be a tool error.
func TestRegress20261012_MCPDivergenceOutOfWindowIsToolError(t *testing.T) {
	ctx, sess := startServer(t)
	cases := []map[string]any{
		{"from_year": 5000, "to_year": 6000},
		{"from_year": 1000, "to_year": 1199},
		{"from_year": 2031, "to_year": 2026}, // reversed: also an empty range
	}
	for _, args := range cases {
		isErr, txt := call(t, ctx, sess, "vn_kr_divergence", args)
		if !isErr {
			t.Errorf("vn_kr_divergence %v: isError=false, text=%q; want a tool error", args, txt)
		}
	}
	if isErr, txt := call(t, ctx, sess, "vn_kr_divergence", map[string]any{"from_year": 2026, "to_year": 2031}); isErr || !strings.Contains(txt, "2027-02-06") {
		t.Errorf("vn_kr_divergence 2026-2031: isError=%v text=%q", isErr, txt)
	}
	// A range that is genuinely in window but has no divergence is a plain answer.
	if isErr, txt := call(t, ctx, sess, "vn_kr_divergence", map[string]any{"from_year": 2026, "to_year": 2026}); isErr || !strings.Contains(txt, "No VN/KR divergence") {
		t.Errorf("vn_kr_divergence 2026-2026: isError=%v text=%q", isErr, txt)
	}
}
