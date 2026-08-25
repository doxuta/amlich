package amlich

import "testing"

// Reference dates from the production TEdu almanac table (HOL_LUNAR), which
// has run against real users since 2026. These pin the holiday API to civil
// dates independently of the golden conversion files.

var tetVN = map[int]string{
	2024: "2024-02-10", 2025: "2025-01-29", 2026: "2026-02-17",
	2027: "2027-02-06", 2028: "2028-01-26", 2029: "2029-02-13",
	2030: "2030-02-02", 2031: "2031-01-23", 2032: "2032-02-11",
	2033: "2033-01-31", 2034: "2034-02-19", 2035: "2035-02-08",
}

var trungThuVN = map[int]string{
	2024: "2024-09-17", 2025: "2025-10-06", 2026: "2026-09-25",
	2027: "2027-09-15", 2028: "2028-10-03", 2029: "2029-09-22",
	2030: "2030-09-12", 2031: "2031-10-01", 2032: "2032-09-19",
	2033: "2033-09-08", 2034: "2034-09-26", 2035: "2035-09-16",
}

var gioToVN = map[int]string{
	2024: "2024-04-18", 2025: "2025-04-07", 2026: "2026-04-26",
	2027: "2027-04-16", 2028: "2028-04-04", 2029: "2029-04-23",
	2030: "2030-04-12",
}

func findHoliday(t *testing.T, year int, c Country, nameEN string) HolidayDate {
	t.Helper()
	for _, h := range Holidays(year, c) {
		if h.NameEN == nameEN {
			return h
		}
	}
	t.Fatalf("holiday %q not found in %v %d", nameEN, c, year)
	return HolidayDate{}
}

func TestTetReference(t *testing.T) {
	for year, want := range tetVN {
		got := findHoliday(t, year, VN, "Lunar New Year").ISO()
		if got != want {
			t.Errorf("Tết %d = %s, reference %s", year, got, want)
		}
	}
}

func TestTrungThuReference(t *testing.T) {
	for year, want := range trungThuVN {
		got := findHoliday(t, year, VN, "Mid-Autumn Festival").ISO()
		if got != want {
			t.Errorf("Trung thu %d = %s, reference %s", year, got, want)
		}
	}
}

func TestGioToReference(t *testing.T) {
	for year, want := range gioToVN {
		got := findHoliday(t, year, VN, "Hùng Kings' Commemoration").ISO()
		if got != want {
			t.Errorf("Giỗ tổ %d = %s, reference %s", year, got, want)
		}
	}
}

func TestCanChi(t *testing.T) {
	cases := []struct {
		year   int
		vi, ko string
	}{
		{1900, "Canh Tý", "경자년"},
		{2024, "Giáp Thìn", "갑진년"},
		{2025, "Ất Tỵ", "을사년"},
		{2026, "Bính Ngọ", "병오년"},
		{2027, "Đinh Mùi", "정미년"},
		{2028, "Mậu Thân", "무신년"},
	}
	for _, c := range cases {
		got := CanChi(c.year)
		if got.Vietnamese != c.vi {
			t.Errorf("CanChi(%d).Vietnamese = %q, want %q", c.year, got.Vietnamese, c.vi)
		}
		if got.Korean != c.ko {
			t.Errorf("CanChi(%d).Korean = %q, want %q", c.year, got.Korean, c.ko)
		}
	}
}

func TestHolidaysOrderedAndInYear(t *testing.T) {
	for _, c := range []Country{VN, KR} {
		for year := 2020; year <= 2040; year++ {
			hs := Holidays(year, c)
			if len(hs) == 0 {
				t.Fatalf("no holidays for %v %d", c, year)
			}
			prev := 0
			for _, h := range hs {
				if h.SolarY != year {
					t.Fatalf("%v %d: %s resolved outside year: %s", c, year, h.NameEN, h.ISO())
				}
				cur := h.SolarM*100 + h.SolarD
				if cur < prev {
					t.Fatalf("%v %d: holidays out of order", c, year)
				}
				prev = cur
			}
		}
	}
}

func TestMonthDaysRange(t *testing.T) {
	for ly := 1990; ly <= 2060; ly++ {
		for lm := 1; lm <= 12; lm++ {
			n, err := MonthDays(LunarDate{Year: ly, Month: lm}, Vietnam)
			if err != nil {
				t.Fatalf("MonthDays(%d/%d): %v", lm, ly, err)
			}
			if n != 29 && n != 30 {
				t.Fatalf("MonthDays(%d/%d) = %d", lm, ly, n)
			}
		}
	}
}

func TestDivergenceIsSelfConsistent(t *testing.T) {
	// Every reported divergence must reproduce via direct conversion, and
	// over three centuries the two calendars must diverge at least once —
	// otherwise the dual-zone machinery would be pointless.
	div := Divergence(1900, 2199)
	if len(div) == 0 {
		t.Fatal("no VN/KR divergence found in 300 years — implausible")
	}
	for _, d := range div {
		l := LunarDate{Year: d.LunarYear, Month: d.LunarMonth, Day: d.LunarDay}
		vy, vm, vd, _ := LunarToSolar(l, Vietnam)
		ky, km, kd, _ := LunarToSolar(l, Korea)
		vn := HolidayDate{SolarY: vy, SolarM: vm, SolarD: vd}.ISO()
		kr := HolidayDate{SolarY: ky, SolarM: km, SolarD: kd}.ISO()
		if vn != d.VN || kr != d.KR {
			t.Fatalf("divergence %+v does not reproduce: %s / %s", d, vn, kr)
		}
		if vn == kr {
			t.Fatalf("divergence %+v reports equal dates", d)
		}
	}
	t.Logf("VN/KR divergent observances 1900-2199: %d", len(div))
}

func TestInvalidInputs(t *testing.T) {
	if _, err := SolarToLunar(2026, 2, 30, Vietnam); err == nil {
		t.Error("Feb 30 accepted")
	}
	if _, err := SolarToLunar(2026, 13, 1, Vietnam); err == nil {
		t.Error("month 13 accepted")
	}
	if _, _, _, err := LunarToSolar(LunarDate{Year: 2026, Month: 0, Day: 1}, Vietnam); err == nil {
		t.Error("lunar month 0 accepted")
	}
	if _, _, _, err := LunarToSolar(LunarDate{Year: 2026, Month: 1, Day: 31}, Vietnam); err == nil {
		t.Error("lunar day 31 accepted")
	}
}
