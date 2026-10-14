package amlich

import "testing"

// Regression tests for the 2026-10-14 audit.

// SolarToLunar(1200-01-01) returns 14/12/1199 in every zone, and
// LunarToSolar accepts lunar year MinYear-1 for exactly that reason. MonthDays
// probed the month through its day 1, which falls in civil 1199, so the month
// that SolarToLunar itself emits was rejected as nonexistent. A month is
// computable when any of its days is inside the window; a month entirely
// before it (11/1199, 12/1198) stays rejected.
func TestRegress20261014_MonthDaysAcceptsWindowEdgeMonth(t *testing.T) {
	for _, z := range []Zone{Vietnam, Korea, China} {
		l, err := SolarToLunar(MinYear, 1, 1, z)
		if err != nil {
			t.Fatalf("SolarToLunar(%d-01-01, %v): %v", MinYear, z, err)
		}
		n, err := MonthDays(l, z)
		if err != nil {
			t.Errorf("MonthDays(%d/%d, %v): %v; SolarToLunar returned this month", l.Month, l.Year, z, err)
			continue
		}
		if n != 29 && n != 30 {
			t.Errorf("MonthDays(%d/%d, %v) = %d", l.Month, l.Year, z, n)
		}
		// Days before the window are rejected one by one; the rest exist.
		// So the month length is the first in-window day minus one, plus
		// the number of days LunarToSolar accepts.
		accepted := 0
		for d := 1; d <= 30; d++ {
			if _, _, _, err := LunarToSolar(LunarDate{Year: l.Year, Month: l.Month, Day: d}, z); err == nil {
				accepted++
			}
		}
		if want := l.Day - 1 + accepted; n != want {
			t.Errorf("MonthDays(%d/%d, %v) = %d, want %d (%d days before the window + %d accepted)",
				l.Month, l.Year, z, n, want, l.Day-1, accepted)
		}
		for _, m := range []LunarDate{{Year: MinYear - 1, Month: 11}, {Year: MinYear - 2, Month: 12}, {Year: MaxYear + 1, Month: 1}} {
			if n, err := MonthDays(m, z); err == nil {
				t.Errorf("MonthDays(%d/%d, %v) = %d, want ErrInvalidLunarDate (entirely outside the window)", m.Month, m.Year, z, n)
			}
		}
	}
}

// Holidays carried a "retry the adjacent lunar year" branch for a drift that
// cannot happen: lunar months 1–8 of year Y always fall in civil Y, and
// month 12 of Y-1 always falls in civil Y (it ends 7+ days before Tết, which
// is never before 21 January). The branch was removed; this pins the
// invariant it was guarding, across the whole window and both countries.
func TestRegress20261014_HolidaysResolveEveryDefinitionEveryYear(t *testing.T) {
	for _, c := range []Country{VN, KR} {
		defs := holidaysVN
		if c == KR {
			defs = holidaysKR
		}
		for y := MinYear; y <= MaxYear; y++ {
			hs := Holidays(y, c)
			if len(hs) != len(defs) {
				t.Fatalf("Holidays(%d, %v): %d holidays, want %d", y, c, len(hs), len(defs))
			}
			for _, h := range hs {
				if h.SolarY != y {
					t.Fatalf("Holidays(%d, %v): %s resolved to %s", y, c, h.Name, h.ISO())
				}
			}
		}
	}
}
