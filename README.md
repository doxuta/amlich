# amlich 🌖

**Astronomical Vietnamese & Korean lunisolar calendar engine in pure Go — with
an MCP server so AI agents stop guessing when Tết is.**

`amlich` computes the lunisolar calendar from first principles — new-moon
instants and solar longitude, not lookup tables — so it can convert any date,
resolve every lunisolar holiday (Tết, Giỗ tổ, Trung thu · 설날, 석가탄신일,
추석), name the sexagenary year in Vietnamese *and* Korean readings, and
answer a question almost nobody can: **in which years do Vietnam and Korea
celebrate the "same" lunar holiday on different days?**

```
$ amlich divergence 2026 2031
năm âm 2027, ngày 1/1: VN 2027-02-06 ≠ KR 2027-02-07 (Tết Nguyên Đán / 설날)
năm âm 2028, ngày 1/1: VN 2028-01-26 ≠ KR 2028-01-27 (Tết Nguyên Đán / 설날)
năm âm 2030, ngày 1/1: VN 2030-02-02 ≠ KR 2030-02-03 (Tết Nguyên Đán / 설날)
```

Same new moon — but Vietnam numbers its calendar at UTC+7 and Korea at UTC+9,
so a new moon between 23:00 ICT and 01:00 KST starts the month on different
civil days. Ask an LLM without a tool and it will happily hallucinate these
dates; give it this MCP server and it computes them from planetary motion.

- **Zero dependencies** in the engine (`import "github.com/doxuta/amlich"` pulls only the standard library)
- **Cross-validated**: 59,810 golden conversions + 1,732 astronomy samples against the implementation running in production inside [TEdu](https://github.com/doxuta/tedu), 1900–2199, both zones
- **Fuzzed**: native round-trip fuzz target, 38.9M executions clean; the inputs that broke earlier revisions live in `testdata/fuzz` as a regression corpus
- **Fast**: ~360 ns and 0 allocs per conversion (Apple M3 Pro)
- **Stricter than the reference algorithm**: rejects day 30 of a 29-day month and correctly handles a leap 12th month — two real defects of the classical implementation, documented in [DESIGN.md](DESIGN.md)

## Library

```bash
go get github.com/doxuta/amlich
```

```go
l, _ := amlich.SolarToLunar(2026, 2, 17, amlich.Vietnam)
fmt.Println(l)                            // 1/1/2026 — Tết Bính Ngọ
fmt.Println(amlich.CanChi(l.Year).Korean) // 병오년

y, m, d, _ := amlich.LunarToSolar(amlich.LunarDate{Year: 2026, Month: 8, Day: 15}, amlich.Korea)
// 2026-09-25 — Chuseok

for _, h := range amlich.Holidays(2027, amlich.VN) { fmt.Println(h.ISO(), h.Name) }
for _, div := range amlich.Divergence(2026, 2060) { fmt.Println(div.NameVN, div.VN, "≠", div.KR) }
```

API: `SolarToLunar` · `LunarToSolar` · `MonthDays` · `Holidays` · `CanChi` ·
`Divergence` — see [pkg.go.dev/github.com/doxuta/amlich](https://pkg.go.dev/github.com/doxuta/amlich).

## MCP server (for AI agents)

```bash
go install github.com/doxuta/amlich/cmd/amlich-mcp@latest
claude mcp add amlich -- amlich-mcp
```

Tools: `convert_date`, `lunar_today`, `holidays_in_year`, `vn_kr_divergence`.
Deterministic, offline, no API keys — an agent asked *"when is Seollal 2035?"*
answers from astronomy instead of vibes.

## CLI

```bash
go install github.com/doxuta/amlich/cmd/amlich@latest
amlich today                # hôm nay âm lịch bao nhiêu?
amlich convert 2026-02-17
amlich holidays 2027 -kr    # Korean holiday set at UTC+9
amlich canchi 2027          # Đinh Mùi · 정미년 · Goat
```

## Benchmarks

```
BenchmarkSolarToLunar-12    3405927    360.0 ns/op    0 B/op    0 allocs/op
BenchmarkLunarToSolar-12    3507360    339.8 ns/op    0 B/op    0 allocs/op
```

## Vietnamese — Tóm tắt

Thư viện Go thuần (không dependency) tính âm lịch Việt Nam và Hàn Quốc bằng
thiên văn — điểm sóc và kinh độ mặt trời — thay vì bảng tra, kèm CLI và MCP
server cho AI agent. Trích xuất từ thuật toán đang chạy thật trong
[TEdu](https://github.com/doxuta/tedu), đối chứng 59.810 phép đổi (1900–2199,
cả UTC+7 lẫn UTC+9), fuzz 38,9 triệu lượt. Điểm thú vị nhất: liệt kê những năm
Tết Việt Nam và Seollal Hàn Quốc **lệch nhau một ngày** vì cùng một trăng non
nhưng khác múi giờ pháp định (ví dụ 2027, 2028, 2030, 2053).

## 한국어 — 요약

베트남(UTC+7)과 한국(UTC+9)의 음력을 천문 계산으로 구하는 순수 Go
라이브러리입니다. 신월 시각과 태양 황경을 직접 계산하므로 조견표 없이 어떤
날짜든 변환할 수 있고, 설날·석가탄신일·추석 등 음력 명절 날짜와 간지(예:
2027년 정미년)를 제공합니다. 같은 신월이라도 두 나라의 표준시가 달라 설날과
Tết이 하루 어긋나는 해(2027, 2028, 2030, 2053...)를 계산으로 찾아내는
`Divergence` API가 특징입니다. MCP 서버를 통해 AI 에이전트가 음력 날짜를
추측하지 않고 계산하도록 할 수 있습니다.

## Provenance & AI disclosure

The astronomy follows the classical Meeus-series method popularized for the
Vietnamese calendar by Hồ Ngọc Đức; the porting, hardening, and documentation
were done with an AI coding agent under human review. Full lineage, the three
reference-implementation defects found along the way, and the validation
methodology are in [DESIGN.md](DESIGN.md).

MIT © Xuan Tai Doan
