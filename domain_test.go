package amlich

import (
	"testing"
	"time"
)

// Regression tests for the August 2026 audit finding: the lunation correction
// loops in SolarToLunar and MonthDays were uncapped, and the ΔT polynomial
// stops being monotonic in the lunation index far from the epoch, so
// SolarToLunar(-1000000, 1, 1, Vietnam) never returned. It was reachable from
// the MCP server, which passed a client-supplied year straight through.
//
// Every case here runs under a watchdog: a regression must fail the test, not
// hang the suite.
func within(t *testing.T, d time.Duration, name string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s did not return within %s — the correction loop is unbounded again", name, d)
	}
}

func TestOutOfWindowYearsReturnPromptly(t *testing.T) {
	years := []int{-1000000, -362994, 0, MinYear - 1, MaxYear + 1, 10000000, 30000000}
	for _, y := range years {
		y := y
		within(t, 2*time.Second, "SolarToLunar", func() {
			if _, err := SolarToLunar(y, 1, 1, Vietnam); err == nil {
				t.Errorf("SolarToLunar(%d) returned no error", y)
			}
		})
		within(t, 2*time.Second, "LunarToSolar", func() {
			if _, _, _, err := LunarToSolar(LunarDate{Year: y, Month: 1, Day: 1}, Korea); err == nil {
				t.Errorf("LunarToSolar(%d) returned no error", y)
			}
		})
		within(t, 2*time.Second, "MonthDays", func() {
			if _, err := MonthDays(LunarDate{Year: y, Month: 1}, Vietnam); err == nil {
				t.Errorf("MonthDays(%d) returned no error", y)
			}
		})
		within(t, 2*time.Second, "Holidays", func() {
			if got := Holidays(y, VN); got != nil {
				t.Errorf("Holidays(%d) returned %d entries", y, len(got))
			}
		})
	}
}

func TestDivergenceClampsItsRange(t *testing.T) {
	within(t, 10*time.Second, "Divergence", func() {
		// A caller asking for everything must be clamped, not hang.
		got := Divergence(-999999, 999999)
		if len(got) == 0 {
			t.Error("clamped Divergence returned nothing")
		}
		for _, d := range got {
			if d.LunarYear < MinYear || d.LunarYear > MaxYear {
				t.Errorf("divergence outside the supported window: %d", d.LunarYear)
			}
		}
	})
}

func TestWindowBoundariesStillCompute(t *testing.T) {
	for _, y := range []int{MinYear, MinYear + 1, MaxYear - 1, MaxYear} {
		within(t, 2*time.Second, "boundary", func() {
			l, err := SolarToLunar(y, 6, 15, Vietnam)
			if err != nil {
				t.Errorf("SolarToLunar(%d, 6, 15) = %v, want a date", y, err)
				return
			}
			yy, mm, dd, err := LunarToSolar(l, Vietnam)
			if err != nil {
				t.Errorf("round trip at %d rejected: %v", y, err)
				return
			}
			if yy != y || mm != 6 || dd != 15 {
				t.Errorf("round trip at %d = %d-%d-%d", y, yy, mm, dd)
			}
		})
	}
}

// The doc comment on Divergence claims a whole-lunation split really happens
// and cites lunar year 1985. If that stops being true, the documentation is
// wrong and this test says so.
func TestDivergenceDocumentedExampleHolds(t *testing.T) {
	var tet *Divergent
	for _, d := range Divergence(1985, 1985) {
		if d.LunarDay == 1 && d.LunarMonth == 1 {
			dd := d
			tet = &dd
		}
	}
	if tet == nil {
		t.Fatal("lunar year 1985 no longer diverges at 1/1")
	}
	if tet.VN != "1985-01-21" || tet.KR != "1985-02-20" {
		t.Fatalf("1985 Tết/Seollal = VN %s / KR %s, documented as VN 1985-01-21 / KR 1985-02-20", tet.VN, tet.KR)
	}
}
