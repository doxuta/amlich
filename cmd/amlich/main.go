// Command amlich is a small CLI for the amlich lunisolar engine.
//
//	amlich today [-kr]
//	amlich convert 2026-02-17 [-kr]
//	amlich holidays 2027 [-kr]
//	amlich canchi 2027
//	amlich divergence 2026 2060
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/doxuta/amlich"
)

func main() {
	args := os.Args[1:]
	c := amlich.VN
	// Strip a trailing -kr flag anywhere.
	out := args[:0]
	for _, a := range args {
		if a == "-kr" || a == "--kr" {
			c = amlich.KR
		} else {
			out = append(out, a)
		}
	}
	args = out
	if len(args) == 0 {
		usage()
	}

	switch args[0] {
	case "today":
		loc := time.FixedZone("", int(float64(c.Zone())*3600))
		now := time.Now().In(loc)
		l, err := amlich.SolarToLunar(now.Year(), int(now.Month()), now.Day(), c.Zone())
		die(err)
		cc := amlich.CanChi(l.Year)
		fmt.Printf("%s → âm lịch %s · %s (%s)\n", now.Format("2006-01-02"), l, cc.Vietnamese, cc.Korean)

	case "convert":
		need(args, 2)
		t, err := time.Parse("2006-01-02", args[1])
		die(err)
		l, err := amlich.SolarToLunar(t.Year(), int(t.Month()), t.Day(), c.Zone())
		die(err)
		cc := amlich.CanChi(l.Year)
		fmt.Printf("%s → âm lịch %s · %s (%s)\n", args[1], l, cc.Vietnamese, cc.Korean)

	case "holidays":
		need(args, 2)
		year := atoi(args[1])
		for _, h := range amlich.Holidays(year, c) {
			fmt.Printf("%s  %-22s %s (âm %d/%d)\n", h.ISO(), h.Name, h.NameEN, h.Day, h.Month)
		}

	case "canchi":
		need(args, 2)
		cc := amlich.CanChi(atoi(args[1]))
		fmt.Printf("%s · %s · %s\n", cc.Vietnamese, cc.Korean, cc.AnimalEN)

	case "divergence":
		need(args, 3)
		for _, d := range amlich.Divergence(atoi(args[1]), atoi(args[2])) {
			fmt.Printf("năm âm %d, ngày %d/%d: VN %s ≠ KR %s (%s / %s)\n",
				d.LunarYear, d.LunarDay, d.LunarMonth, d.VN, d.KR, d.NameVN, d.NameKR)
		}

	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  amlich today [-kr]
  amlich convert YYYY-MM-DD [-kr]
  amlich holidays YEAR [-kr]
  amlich canchi LUNAR_YEAR
  amlich divergence FROM_YEAR TO_YEAR`)
	os.Exit(2)
}

func need(args []string, n int) {
	if len(args) < n {
		usage()
	}
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	die(err)
	return n
}

func die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "amlich:", err)
		os.Exit(1)
	}
}
