package amlich_test

import (
	"fmt"

	"github.com/doxuta/amlich"
)

func ExampleSolarToLunar() {
	l, _ := amlich.SolarToLunar(2026, 2, 17, amlich.Vietnam)
	fmt.Println(l, amlich.CanChi(l.Year).Vietnamese)
	// Output: 1/1/2026 Bính Ngọ
}

func ExampleLunarToSolar() {
	// Trung thu (full moon of the 8th month), lunar year 2026.
	y, m, d, _ := amlich.LunarToSolar(amlich.LunarDate{Year: 2026, Month: 8, Day: 15}, amlich.Vietnam)
	fmt.Printf("%04d-%02d-%02d\n", y, m, d)
	// Output: 2026-09-25
}

func ExampleHolidays() {
	for _, h := range amlich.Holidays(2026, amlich.KR)[:2] {
		fmt.Println(h.ISO(), h.NameEN)
	}
	// Output:
	// 2026-02-17 Seollal (Lunar New Year)
	// 2026-03-03 Daeboreum (First Full Moon)
}

func ExampleCanChi() {
	n := amlich.CanChi(2027)
	fmt.Println(n.Vietnamese, "·", n.Korean, "·", n.AnimalEN)
	// Output: Đinh Mùi · 정미년 · Goat
}
