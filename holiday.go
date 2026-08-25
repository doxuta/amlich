package amlich

import "fmt"

// Country selects a national lunisolar tradition: which holidays are observed
// and which civil clock (Zone) numbers the calendar.
type Country int

const (
	// VN — Vietnam: calendar at UTC+7.
	VN Country = iota
	// KR — Korea: calendar at UTC+9.
	KR
)

// Zone returns the civil time zone the country numbers its calendar in.
func (c Country) Zone() Zone {
	if c == KR {
		return Korea
	}
	return Vietnam
}

// Holiday is a lunisolar holiday definition.
type Holiday struct {
	Name    string // native name
	NameEN  string
	Day     int // lunar day
	Month   int // lunar month
	Country Country
}

// HolidayDate is a holiday resolved to a civil date for a specific year.
type HolidayDate struct {
	Holiday
	Year         int // Gregorian year the holiday falls in
	SolarY       int
	SolarM       int
	SolarD       int
}

// ISO renders the resolved civil date as YYYY-MM-DD.
func (h HolidayDate) ISO() string {
	return fmt.Sprintf("%04d-%02d-%02d", h.SolarY, h.SolarM, h.SolarD)
}

// The observed sets. Vietnam and Korea share the 1/1 new year and the 15/8
// full moon, but differ elsewhere — including Buddha's Birthday, which
// Vietnam keeps on 15/4 and Korea on 8/4.
var (
	holidaysVN = []Holiday{
		{"Tết Nguyên Đán", "Lunar New Year", 1, 1, VN},
		{"Tết Hàn thực", "Cold Food Festival", 3, 3, VN},
		{"Giỗ tổ Hùng Vương", "Hùng Kings' Commemoration", 10, 3, VN},
		{"Phật đản", "Buddha's Birthday", 15, 4, VN},
		{"Tết Đoan ngọ", "Double Fifth Festival", 5, 5, VN},
		{"Lễ Vu lan", "Ullambana (Ghost Festival)", 15, 7, VN},
		{"Tết Trung thu", "Mid-Autumn Festival", 15, 8, VN},
		{"Ông Táo chầu trời", "Kitchen Gods' Day", 23, 12, VN},
	}
	holidaysKR = []Holiday{
		{"설날", "Seollal (Lunar New Year)", 1, 1, KR},
		{"정월대보름", "Daeboreum (First Full Moon)", 15, 1, KR},
		{"석가탄신일", "Buddha's Birthday", 8, 4, KR},
		{"단오", "Dano (Double Fifth)", 5, 5, KR},
		{"추석", "Chuseok (Harvest Full Moon)", 15, 8, KR},
	}
)

// Holidays resolves the country's lunisolar holidays that fall inside
// Gregorian year year, in calendar order.
//
// Lunar months 1..8 of lunar year Y fall in Gregorian year Y; month 12 dates
// (Ông Táo) fall early in Gregorian year Y+1, so they are resolved from lunar
// year year-1.
func Holidays(year int, c Country) []HolidayDate {
	defs := holidaysVN
	if c == KR {
		defs = holidaysKR
	}
	z := c.Zone()
	out := make([]HolidayDate, 0, len(defs))
	for _, h := range defs {
		ly := year
		if h.Month >= 12 {
			ly = year - 1
		}
		y, m, d, err := LunarToSolar(LunarDate{Year: ly, Month: h.Month, Day: h.Day}, z)
		if err != nil {
			continue
		}
		if y != year {
			// Rare drift across the Gregorian boundary — retry adjacent year.
			y2, m2, d2, err2 := LunarToSolar(LunarDate{Year: ly - 1, Month: h.Month, Day: h.Day}, z)
			if err2 == nil && y2 == year {
				y, m, d = y2, m2, d2
			} else {
				continue
			}
		}
		out = append(out, HolidayDate{Holiday: h, Year: year, SolarY: y, SolarM: m, SolarD: d})
	}
	sortHolidayDates(out)
	return out
}

func sortHolidayDates(hs []HolidayDate) {
	for i := 1; i < len(hs); i++ {
		for j := i; j > 0; j-- {
			a, b := hs[j-1], hs[j]
			if a.SolarM > b.SolarM || (a.SolarM == b.SolarM && a.SolarD > b.SolarD) {
				hs[j-1], hs[j] = b, a
			} else {
				break
			}
		}
	}
}

// Divergent is a lunar anchor date whose civil date differs between the
// Vietnamese (UTC+7) and Korean (UTC+9) calendars in a given year.
type Divergent struct {
	LunarDay   int
	LunarMonth int
	LunarYear  int
	NameVN     string
	NameKR     string
	VN         string // YYYY-MM-DD as observed in Vietnam
	KR         string // YYYY-MM-DD as observed in Korea
}

// divergenceAnchors are the shared observances worth comparing.
var divergenceAnchors = []struct {
	d, m           int
	nameVN, nameKR string
}{
	{1, 1, "Tết Nguyên Đán", "설날 (Seollal)"},
	{15, 1, "Rằm tháng Giêng", "정월대보름 (Daeboreum)"},
	{5, 5, "Tết Đoan ngọ", "단오 (Dano)"},
	{15, 8, "Tết Trung thu", "추석 (Chuseok)"},
}

// Divergence enumerates, over lunar years [fromYear, toYear], the shared
// VN/KR observances whose civil dates differ because the same new moon falls
// on different sides of midnight at UTC+7 versus UTC+9.
func Divergence(fromYear, toYear int) []Divergent {
	var out []Divergent
	for ly := fromYear; ly <= toYear; ly++ {
		for _, a := range divergenceAnchors {
			l := LunarDate{Year: ly, Month: a.m, Day: a.d}
			vy, vm, vd, errV := LunarToSolar(l, Vietnam)
			ky, km, kd, errK := LunarToSolar(l, Korea)
			if errV != nil || errK != nil {
				continue
			}
			if vy != ky || vm != km || vd != kd {
				out = append(out, Divergent{
					LunarDay: a.d, LunarMonth: a.m, LunarYear: ly,
					NameVN: a.nameVN, NameKR: a.nameKR,
					VN: fmt.Sprintf("%04d-%02d-%02d", vy, vm, vd),
					KR: fmt.Sprintf("%04d-%02d-%02d", ky, km, kd),
				})
			}
		}
	}
	return out
}
