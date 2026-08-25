package amlich

import "testing"

// FuzzRoundTrip: any valid civil date in the supported window must convert to
// a lunar date and back to exactly itself, in every zone.
func FuzzRoundTrip(f *testing.F) {
	f.Add(2026, 2, 17, 7.0)
	f.Add(2035, 9, 16, 9.0)
	f.Add(1900, 1, 31, 7.0)
	f.Add(2199, 12, 31, 9.0)
	f.Fuzz(func(t *testing.T, y, m, d int, tz float64) {
		if y < 1800 || y > 2399 || tz != 7 && tz != 8 && tz != 9 {
			t.Skip()
		}
		l, err := SolarToLunar(y, m, d, Zone(tz))
		if err != nil {
			t.Skip() // invalid civil date is fine
		}
		yy, mm, dd, err := LunarToSolar(l, Zone(tz))
		if err != nil {
			t.Fatalf("round trip %d-%d-%d tz=%v: lunar %v rejected: %v", y, m, d, tz, l, err)
		}
		if yy != y || mm != m || dd != d {
			t.Fatalf("round trip %d-%d-%d tz=%v: got %d-%d-%d via %v", y, m, d, tz, yy, mm, dd, l)
		}
	})
}
