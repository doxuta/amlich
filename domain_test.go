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

// The supported window is a window of CIVIL years: SolarToLunar accepts every
// civil day of MinYear, so — per the invariant stated on FuzzRoundTrip — every
// one of them must convert back. Civil January of MinYear belongs to lunar year
// MinYear-1, which the lunar-year gate used to refuse outright.
func TestWindowStartCivilDaysRoundTrip(t *testing.T) {
	for _, z := range []Zone{Vietnam, China, Korea} {
		for day := 1; day <= 31; day++ {
			l, err := SolarToLunar(MinYear, 1, day, z)
			if err != nil {
				t.Fatalf("SolarToLunar(%d-01-%02d, tz=%v) = %v", MinYear, day, z, err)
			}
			yy, mm, dd, err := LunarToSolar(l, z)
			if err != nil {
				t.Errorf("round trip %d-01-%02d tz=%v: SolarToLunar gave %v, LunarToSolar rejected it: %v",
					MinYear, day, z, l, err)
				continue
			}
			if yy != MinYear || mm != 1 || dd != day {
				t.Errorf("round trip %d-01-%02d tz=%v: got %04d-%02d-%02d", MinYear, day, z, yy, mm, dd)
			}
		}
	}
}

// Ông Táo (lunar 23/12) of lunar year 1199 falls on 1200-01-10, inside the
// supported window, so Holidays(1200, VN) must list it: the doc promises the
// holidays that fall inside the Gregorian year, and 1200 is in window.
func TestHolidaysCompleteAtWindowStart(t *testing.T) {
	if l, err := SolarToLunar(MinYear, 1, 10, Vietnam); err != nil || l.Day != 23 || l.Month != 12 {
		t.Fatalf("precondition: SolarToLunar(%d-01-10) = %v, %v", MinYear, l, err)
	}
	got := Holidays(MinYear, VN)
	if len(got) != len(holidaysVN) {
		t.Errorf("Holidays(%d, VN) returned %d of %d holidays", MinYear, len(got), len(holidaysVN))
	}
	for _, h := range got {
		if h.NameEN == "Kitchen Gods' Day" {
			if h.ISO() != "1200-01-10" {
				t.Errorf("Kitchen Gods' Day %d = %s, want 1200-01-10", MinYear, h.ISO())
			}
			return
		}
	}
	t.Errorf("Holidays(%d, VN) is missing \"Kitchen Gods' Day\", which falls on 1200-01-10", MinYear)
}
