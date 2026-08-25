package amlich

import "math"

// Astronomical core — a Go port of the lunisolar algorithm running in
// production inside TEdu (https://github.com/doxuta/tedu), which follows the
// classical method popularized for the Vietnamese calendar by Hồ Ngọc Đức,
// itself based on the truncated series of Jean Meeus, "Astronomical
// Algorithms" (2nd ed., 1998): new-moon instants from the lunation series and
// apparent solar longitude for the principal terms (trung khí).
//
// All floor operations deliberately mirror JavaScript Math.floor (toward -inf)
// so results are bit-identical with the production implementation they were
// cross-validated against (testdata/golden_*.txt).

const (
	jdEpoch1900   = 2415021.076998695 // JD of the 1900 lunation epoch
	synodicMonth  = 29.530588853      // mean synodic month length in days
	gregorianJD   = 2299161           // first JD of the Gregorian calendar
	radPerDeg     = math.Pi / 180
)

func fl(x float64) int { return int(math.Floor(x)) }

// jdFromDate returns the Julian day number of dd/mm/yy (proleptic Julian
// before the Gregorian switch, matching the reference implementation).
func jdFromDate(dd, mm, yy int) int {
	a := fl(float64(14-mm) / 12)
	y := yy + 4800 - a
	m := mm + 12*a - 3
	jd := dd + fl(float64(153*m+2)/5) + 365*y + fl(float64(y)/4) - fl(float64(y)/100) + fl(float64(y)/400) - 32045
	if jd < gregorianJD {
		jd = dd + fl(float64(153*m+2)/5) + 365*y + fl(float64(y)/4) - 32083
	}
	return jd
}

// jdToDate converts a Julian day number back to civil dd, mm, yy.
func jdToDate(jd int) (dd, mm, yy int) {
	var a, b, c int
	if jd > gregorianJD-1 {
		a = jd + 32044
		b = fl(float64(4*a+3) / 146097)
		c = a - fl(float64(b*146097)/4)
	} else {
		b = 0
		c = jd + 32082
	}
	d := fl(float64(4*c+3) / 1461)
	e := c - fl(float64(1461*d)/4)
	m := fl(float64(5*e+2) / 153)
	dd = e - fl(float64(153*m+2)/5) + 1
	mm = m + 3 - 12*fl(float64(m)/10)
	yy = b*100 + d - 4800 + fl(float64(m)/10)
	return
}

// newMoonDay returns the Julian day number of the k-th new moon after the
// 1900 epoch, as observed in a zone tz hours east of UTC.
func newMoonDay(k int, tz float64) int {
	kf := float64(k)
	t := kf / 1236.85
	t2 := t * t
	t3 := t2 * t
	jd1 := 2415020.75933 + 29.53058868*kf + 0.0001178*t2 - 0.000000155*t3
	jd1 += 0.00033 * math.Sin((166.56+132.87*t-0.009173*t2)*radPerDeg)
	m := 359.2242 + 29.10535608*kf - 0.0000333*t2 - 0.00000347*t3
	mpr := 306.0253 + 385.81691806*kf + 0.0107306*t2 + 0.00001236*t3
	f := 21.2964 + 390.67050646*kf - 0.0016528*t2 - 0.00000239*t3
	c1 := (0.1734-0.000393*t)*math.Sin(m*radPerDeg) + 0.0021*math.Sin(2*radPerDeg*m)
	c1 = c1 - 0.4068*math.Sin(mpr*radPerDeg) + 0.0161*math.Sin(radPerDeg*2*mpr)
	c1 = c1 - 0.0004*math.Sin(radPerDeg*3*mpr)
	c1 = c1 + 0.0104*math.Sin(radPerDeg*2*f) - 0.0051*math.Sin(radPerDeg*(m+mpr))
	c1 = c1 - 0.0074*math.Sin(radPerDeg*(m-mpr)) + 0.0004*math.Sin(radPerDeg*(2*f+m))
	c1 = c1 - 0.0004*math.Sin(radPerDeg*(2*f-m)) - 0.0006*math.Sin(radPerDeg*(2*f+mpr))
	c1 = c1 + 0.0010*math.Sin(radPerDeg*(2*f-mpr)) + 0.0005*math.Sin(radPerDeg*(2*mpr+m))
	var deltaT float64
	if t < -11 {
		deltaT = 0.001 + 0.000839*t + 0.0002261*t2 - 0.00000845*t3 - 0.000000081*t*t3
	} else {
		deltaT = -0.000278 + 0.000265*t + 0.000262*t2
	}
	return fl(jd1 + c1 - deltaT + 0.5 + tz/24)
}

// sunLongitude returns which of the 12 zodiacal sectors (0..11, each 30°)
// the apparent solar longitude falls in at local midnight of jdn.
func sunLongitude(jdn int, tz float64) int {
	t := (float64(jdn) - 2451545.5 - tz/24) / 36525
	t2 := t * t
	m := 357.52910 + 35999.05030*t - 0.0001559*t2 - 0.00000048*t*t2
	l0 := 280.46645 + 36000.76983*t + 0.0003032*t2
	dl := (1.914600 - 0.004817*t - 0.000014*t2) * math.Sin(radPerDeg*m)
	dl += (0.019993-0.000101*t)*math.Sin(radPerDeg*2*m) + 0.000290*math.Sin(radPerDeg*3*m)
	l := (l0 + dl) * radPerDeg
	l -= 2 * math.Pi * math.Floor(l/(2*math.Pi))
	return fl(l / math.Pi * 6)
}

// lunarMonth11 returns the JD of the first day of the lunar month containing
// the winter solstice (month 11) of Gregorian year yy.
func lunarMonth11(yy int, tz float64) int {
	off := jdFromDate(31, 12, yy) - 2415021
	k := fl(float64(off) / synodicMonth)
	nm := newMoonDay(k, tz)
	if sunLongitude(nm, tz) >= 9 {
		nm = newMoonDay(k-1, tz)
	}
	return nm
}

// leapMonthOffset locates the leap month: the offset (in months after month
// 11 starting at a11) of the first month that contains no principal term.
func leapMonthOffset(a11 int, tz float64) int {
	k := fl((float64(a11)-jdEpoch1900)/synodicMonth + 0.5)
	last := 0
	i := 1
	arc := sunLongitude(newMoonDay(k+i, tz), tz)
	for {
		last = arc
		i++
		arc = sunLongitude(newMoonDay(k+i, tz), tz)
		if arc == last || i >= 14 {
			break
		}
	}
	return i - 1
}
