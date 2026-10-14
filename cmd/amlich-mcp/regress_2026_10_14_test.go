package main

import (
	"strings"
	"testing"

	"github.com/doxuta/amlich"
)

// convert_date documents `leap` as "for lunar_to_solar", and on the
// solar_to_lunar direction it silently ignored it: an agent asking for the
// lunar date of 2025-07-25 "in the leap month" got a plain answer with no
// hint that the argument did nothing. A parameter that cannot apply is a tool
// error. lunar_to_solar keeps honouring it.
func TestRegress20261014_MCPConvertLeapOnSolarToLunarIsToolError(t *testing.T) {
	ctx, sess := startServer(t)
	isErr, txt := call(t, ctx, sess, "convert_date", map[string]any{
		"direction": "solar_to_lunar", "year": 2025, "month": 7, "day": 25, "leap": true,
	})
	if !isErr {
		t.Errorf("convert_date solar_to_lunar leap=true: isError=false, text=%q; want a tool error", txt)
	}
	// Without the flag the same date converts, and reports the leap month it
	// actually falls in (2025 has a leap 6th month; 2025-07-25 is its day 1).
	isErr, txt = call(t, ctx, sess, "convert_date", map[string]any{
		"direction": "solar_to_lunar", "year": 2025, "month": 7, "day": 25,
	})
	if isErr || !strings.Contains(txt, "month 6 (leap)") {
		t.Errorf("convert_date solar_to_lunar 2025-07-25: isError=%v text=%q", isErr, txt)
	}
	// lunar_to_solar with leap still works and still differs from the
	// non-leap month.
	y, m, d, err := amlich.LunarToSolar(amlich.LunarDate{Year: 2025, Month: 6, Day: 1, Leap: true}, amlich.Vietnam)
	if err != nil || y != 2025 || m != 7 || d != 25 {
		t.Fatalf("LunarToSolar(1/6+/2025) = %d-%02d-%02d, %v", y, m, d, err)
	}
	isErr, txt = call(t, ctx, sess, "convert_date", map[string]any{
		"direction": "lunar_to_solar", "year": 2025, "month": 6, "day": 1, "leap": true,
	})
	if isErr || !strings.Contains(txt, "2025-07-25") {
		t.Errorf("convert_date lunar_to_solar 1/6+/2025: isError=%v text=%q", isErr, txt)
	}
	isErr, txt = call(t, ctx, sess, "convert_date", map[string]any{
		"direction": "lunar_to_solar", "year": 2025, "month": 6, "day": 1,
	})
	if isErr || strings.Contains(txt, "2025-07-25") {
		t.Errorf("convert_date lunar_to_solar 1/6/2025 (not leap): isError=%v text=%q", isErr, txt)
	}
}
