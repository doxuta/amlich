// Package amlich computes the Vietnamese and Korean lunisolar calendars from
// first-principles astronomy — new-moon instants and apparent solar longitude
// — instead of lookup tables, so any date (past or future) can be converted.
//
// The same astronomical month runs on different civil clocks: Vietnam numbers
// its calendar at UTC+7, Korea at UTC+9 (China at UTC+8). When a new moon
// falls close to local midnight, the two national calendars can start a month
// on different days — which is why Tết and Seollal, or Trung thu and Chuseok,
// occasionally land on different dates. Divergence enumerates those years.
//
// Accuracy: the truncated series used here (Meeus, via the reference
// implementation by Hồ Ngọc Đức) is reliable for roughly 1200–3000 CE;
// results are cross-validated against the production implementation inside
// TEdu for 1900–2199 (see testdata/).
package amlich

import (
	"errors"
	"fmt"
)

// Zone is a civil time zone expressed in hours east of UTC, used as the
// reference clock for numbering the lunisolar calendar.
type Zone float64

const (
	// Vietnam numbers its lunisolar calendar at UTC+7 (since 1967-08-08 in
	// the North, unified 1975).
	Vietnam Zone = 7
	// Korea numbers its lunisolar calendar at UTC+9 (KST).
	Korea Zone = 9
	// China numbers its lunisolar calendar at UTC+8 (CST).
	China Zone = 8
)

// LunarDate is a date in a lunisolar calendar. Leap reports whether the date
// belongs to the intercalary (leap) month with the same Month number.
type LunarDate struct {
	Year  int
	Month int
	Day   int
	Leap  bool
}

// String renders the date in the conventional Vietnamese order, e.g.
// "15/8/2026" or "2/6+/2025" for a leap-month date.
func (l LunarDate) String() string {
	leap := ""
	if l.Leap {
		leap = "+"
	}
	return fmt.Sprintf("%d/%d%s/%d", l.Day, l.Month, leap, l.Year)
}

// ErrInvalidLunarDate is returned when a lunar date does not exist, e.g. day
// 30 of a 29-day month, or a leap-month flag on a year/month without one.
var ErrInvalidLunarDate = errors.New("amlich: lunar date does not exist")

// ErrInvalidSolarDate is returned for an impossible civil date.
var ErrInvalidSolarDate = errors.New("amlich: invalid solar date")

// SolarToLunar converts a civil (Gregorian) date to the lunisolar date as
// numbered in zone z.
func SolarToLunar(year, month, day int, z Zone) (LunarDate, error) {
	if !validSolar(year, month, day) {
		return LunarDate{}, ErrInvalidSolarDate
	}
	tz := float64(z)
	dayNumber := jdFromDate(day, month, year)
	// Estimate the lunation, then correct in both directions: monthStart must
	// be the latest new moon <= dayNumber. The reference algorithm's single
	// step-down misses rare boundary cases far from the epoch (found by fuzz:
	// 2185-03-31 at UTC+9).
	k := fl((float64(dayNumber) - jdEpoch1900) / synodicMonth)
	for newMoonDay(k+1, tz) <= dayNumber {
		k++
	}
	for newMoonDay(k, tz) > dayNumber {
		k--
	}
	monthStart := newMoonDay(k, tz)
	a11 := lunarMonth11(year, tz)
	b11 := a11
	lunarYear := 0
	if a11 >= monthStart {
		lunarYear = year
		a11 = lunarMonth11(year-1, tz)
	} else {
		lunarYear = year + 1
		b11 = lunarMonth11(year+1, tz)
	}
	lunarDay := dayNumber - monthStart + 1
	diff := fl(float64(monthStart-a11) / 29)
	lunarLeap := false
	lunarMonth := diff + 11
	if b11-a11 > 365 {
		leapMonthDiff := leapMonthOffset(a11, tz)
		if diff >= leapMonthDiff {
			lunarMonth = diff + 10
			if diff == leapMonthDiff {
				lunarLeap = true
			}
		}
	}
	if lunarMonth > 12 {
		lunarMonth -= 12
	}
	if lunarMonth >= 11 && diff < 4 {
		lunarYear--
	}
	return LunarDate{Year: lunarYear, Month: lunarMonth, Day: lunarDay, Leap: lunarLeap}, nil
}

// LunarToSolar converts a lunisolar date (as numbered in zone z) to the civil
// (Gregorian) date it falls on. Unlike the reference implementation, it
// validates month length: day 30 of a 29-day month is ErrInvalidLunarDate
// instead of silently overflowing into the next month.
func LunarToSolar(l LunarDate, z Zone) (year, month, day int, err error) {
	if l.Month < 1 || l.Month > 12 || l.Day < 1 || l.Day > 30 {
		return 0, 0, 0, ErrInvalidLunarDate
	}
	tz := float64(z)
	var a11, b11 int
	if l.Month < 11 {
		a11 = lunarMonth11(l.Year-1, tz)
		b11 = lunarMonth11(l.Year, tz)
	} else {
		a11 = lunarMonth11(l.Year, tz)
		b11 = lunarMonth11(l.Year+1, tz)
	}
	k := fl(0.5 + (float64(a11)-jdEpoch1900)/synodicMonth)
	off := l.Month - 11
	if off < 0 {
		off += 12
	}
	if b11-a11 > 365 {
		leapOff := leapMonthOffset(a11, tz)
		leapM := leapOff - 2
		if leapM < 0 {
			leapM += 12
		}
		if leapM == 0 {
			// Offset arithmetic yields 0 for a leap 12th month; the reference
			// implementation forgets this case and rejects every leap-12 date.
			leapM = 12
		}
		if l.Leap && l.Month != leapM {
			return 0, 0, 0, ErrInvalidLunarDate
		} else if l.Leap || off >= leapOff {
			off++
		}
	} else if l.Leap {
		return 0, 0, 0, ErrInvalidLunarDate
	}
	start := newMoonDay(k+off, tz)
	if l.Day > start2Len(start, k+off, tz) {
		return 0, 0, 0, ErrInvalidLunarDate
	}
	day, month, year = jdToDate(start + l.Day - 1)
	return year, month, day, nil
}

// MonthDays reports the length (29 or 30 days) of the given lunar month as
// numbered in zone z. The Day field of l is ignored.
func MonthDays(l LunarDate, z Zone) (int, error) {
	probe := l
	probe.Day = 1
	y, m, d, err := LunarToSolar(probe, z)
	if err != nil {
		return 0, err
	}
	start := jdFromDate(d, m, y)
	k := fl(0.5 + (float64(start)-jdEpoch1900)/synodicMonth)
	for newMoonDay(k, float64(z)) > start {
		k--
	}
	for newMoonDay(k+1, float64(z)) <= start {
		k++
	}
	return start2Len(start, k, float64(z)), nil
}

// start2Len returns the month length for the month starting at JD start,
// where k is that month's lunation number.
func start2Len(start, k int, tz float64) int {
	return newMoonDay(k+1, tz) - start
}

func validSolar(year, month, day int) bool {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	jd := jdFromDate(day, month, year)
	d, m, y := jdToDate(jd)
	return d == day && m == month && y == year
}
