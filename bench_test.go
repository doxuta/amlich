package amlich

import "testing"

var sinkLunar LunarDate

func BenchmarkSolarToLunar(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l, _ := SolarToLunar(2026, 2, 17, Vietnam)
		sinkLunar = l
	}
}

var sinkInt int

func BenchmarkLunarToSolar(b *testing.B) {
	b.ReportAllocs()
	l := LunarDate{Year: 2026, Month: 1, Day: 1}
	for i := 0; i < b.N; i++ {
		_, _, d, _ := LunarToSolar(l, Vietnam)
		sinkInt = d
	}
}

func BenchmarkHolidaysYear(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkInt = len(Holidays(2026, VN))
	}
}

func BenchmarkDivergenceCentury(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkInt = len(Divergence(2000, 2100))
	}
}
