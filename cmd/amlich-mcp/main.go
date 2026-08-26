// Command amlich-mcp exposes the amlich lunisolar engine as an MCP server
// (stdio transport), so AI agents can compute — not guess — Vietnamese and
// Korean lunar dates and holidays.
//
// Install:
//
//	go install github.com/doxuta/amlich/cmd/amlich-mcp@latest
//
// Register with Claude Code:
//
//	claude mcp add amlich -- amlich-mcp
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/doxuta/amlich"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func country(s string) (amlich.Country, error) {
	switch strings.ToLower(s) {
	case "", "vn", "vietnam":
		return amlich.VN, nil
	case "kr", "korea":
		return amlich.KR, nil
	}
	return amlich.VN, fmt.Errorf("unknown country %q (use \"vn\" or \"kr\")", s)
}

func text(format string, a ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, a...)}}}
}

type convertArgs struct {
	Direction string `json:"direction" jsonschema:"\"solar_to_lunar\" or \"lunar_to_solar\""`
	Year      int    `json:"year" jsonschema:"year of the input date (1200-3000)"`
	Month     int    `json:"month" jsonschema:"month of the input date (1-12)"`
	Day       int    `json:"day" jsonschema:"day of the input date"`
	Leap      bool   `json:"leap,omitempty" jsonschema:"for lunar_to_solar: the date is in the leap (intercalary) month"`
	Country   string `json:"country,omitempty" jsonschema:"calendar tradition: \"vn\" (UTC+7, default) or \"kr\" (UTC+9)"`
}

type yearArgs struct {
	Year    int    `json:"year" jsonschema:"Gregorian year (1200-3000)"`
	Country string `json:"country,omitempty" jsonschema:"\"vn\" (default) or \"kr\""`
}

type todayArgs struct {
	Country string `json:"country,omitempty" jsonschema:"\"vn\" (default) or \"kr\""`
}

type divergenceArgs struct {
	FromYear int `json:"from_year" jsonschema:"first lunar year to check (1200-3000)"`
	ToYear   int `json:"to_year" jsonschema:"last lunar year to check (1200-3000)"`
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "amlich", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "convert_date",
		Description: "Convert between the Gregorian calendar and the Vietnamese (UTC+7) or Korean (UTC+9) lunisolar calendar using astronomical computation. Exact for 1900-2199 and reliable well beyond.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, a convertArgs) (*mcp.CallToolResult, any, error) {
		c, err := country(a.Country)
		if err != nil {
			return nil, nil, err
		}
		z := c.Zone()
		switch a.Direction {
		case "solar_to_lunar":
			l, err := amlich.SolarToLunar(a.Year, a.Month, a.Day, z)
			if err != nil {
				return nil, nil, err
			}
			cc := amlich.CanChi(l.Year)
			return text("Solar %04d-%02d-%02d = lunar day %d, month %d%s, year %d (%s / %s). Zone UTC+%v.",
				a.Year, a.Month, a.Day, l.Day, l.Month, leapMark(l.Leap), l.Year, cc.Vietnamese, cc.Korean, float64(z)), nil, nil
		case "lunar_to_solar":
			y, m, d, err := amlich.LunarToSolar(amlich.LunarDate{Year: a.Year, Month: a.Month, Day: a.Day, Leap: a.Leap}, z)
			if err != nil {
				return nil, nil, err
			}
			return text("Lunar %d/%d%s/%d = solar %04d-%02d-%02d. Zone UTC+%v.",
				a.Day, a.Month, leapMark(a.Leap), a.Year, y, m, d, float64(z)), nil, nil
		}
		return nil, nil, fmt.Errorf("direction must be solar_to_lunar or lunar_to_solar")
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "lunar_today",
		Description: "Today's date in the Vietnamese or Korean lunisolar calendar, with the sexagenary (can-chi) year name.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, a todayArgs) (*mcp.CallToolResult, any, error) {
		c, err := country(a.Country)
		if err != nil {
			return nil, nil, err
		}
		loc := time.FixedZone("", int(float64(c.Zone())*3600))
		now := time.Now().In(loc)
		l, err := amlich.SolarToLunar(now.Year(), int(now.Month()), now.Day(), c.Zone())
		if err != nil {
			return nil, nil, err
		}
		cc := amlich.CanChi(l.Year)
		return text("Today %s = lunar day %d, month %d%s, year %d %s (%s).",
			now.Format("2006-01-02"), l.Day, l.Month, leapMark(l.Leap), l.Year, cc.Vietnamese, cc.Korean), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "holidays_in_year",
		Description: "All lunisolar holidays of Vietnam (Tết, Giỗ tổ, Trung thu...) or Korea (Seollal, Chuseok...) resolved to exact Gregorian dates for a year.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, a yearArgs) (*mcp.CallToolResult, any, error) {
		c, err := country(a.Country)
		if err != nil {
			return nil, nil, err
		}
		var b strings.Builder
		for _, h := range amlich.Holidays(a.Year, c) {
			fmt.Fprintf(&b, "%s  %s (%s) — lunar %d/%d\n", h.ISO(), h.Name, h.NameEN, h.Day, h.Month)
		}
		return text("%s", b.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "vn_kr_divergence",
		Description: "Years in which the Vietnamese (UTC+7) and Korean (UTC+9) lunisolar calendars place a shared observance (Lunar New Year, Mid-Autumn/Chuseok, Dano, Daeboreum) on different civil dates.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, a divergenceArgs) (*mcp.CallToolResult, any, error) {
		div := amlich.Divergence(a.FromYear, a.ToYear)
		if len(div) == 0 {
			return text("No VN/KR divergence in lunar years %d-%d.", a.FromYear, a.ToYear), nil, nil
		}
		var b strings.Builder
		for _, d := range div {
			fmt.Fprintf(&b, "Lunar %d/%d year %d: %s in VN on %s, %s in KR on %s\n",
				d.LunarDay, d.LunarMonth, d.LunarYear, d.NameVN, d.VN, d.NameKR, d.KR)
		}
		return text("%s", b.String()), nil, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func leapMark(leap bool) string {
	if leap {
		return " (leap)"
	}
	return ""
}
