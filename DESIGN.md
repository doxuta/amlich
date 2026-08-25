# Design notes

## Lineage — what is borrowed, what is original

The astronomical core follows the classical method for the Vietnamese
lunisolar calendar popularized by **Hồ Ngọc Đức** (his widely-ported
`amlich` JavaScript), which in turn uses truncated series from **Jean Meeus,
_Astronomical Algorithms_ (2nd ed., 1998)**: a lunation-series approximation
for new-moon instants and a short series for apparent solar longitude. I first
implemented that method in JavaScript inside my production app
[TEdu](https://github.com/doxuta/tedu); this repository is my Go port of it,
cross-validated bit-for-bit against that production implementation.

Original to this library:

1. **Dual-zone model and the `Divergence` API.** The same astronomical events
   are numbered against different civil clocks: Vietnam (UTC+7) and Korea
   (UTC+9). When a new moon falls between 23:00 ICT and 01:00 KST, the two
   national calendars begin the month a day apart — so Tết 2027 is Feb 6 in
   Vietnam while Seollal 2027 is Feb 7 in Korea. `Divergence` enumerates every
   such year computationally instead of citing folklore.
2. **Stricter validation than the reference.** Three defects found while
   porting and fuzzing:
   - *Day-30 overflow*: the reference happily returns a date for day 30 of a
     29-day month — silently the 1st of the next month. This library returns
     `ErrInvalidLunarDate`.
   - *Leap-12 rejection*: the reference maps the leap month via
     `leapM = leapOff − 2 (mod 12)` and compares with months `1..12`, so a
     leap 12th month (`leapM = 0`) can never match and is always rejected.
     Fixed by treating 0 as 12.
   - *Lunation-estimate drift*: `k = ⌊(jd − epoch)/29.5306⌋` with a single
     step-down correction fails on rare boundary dates far from the epoch
     (found by fuzzing: 2185-03-31 at UTC+9). Fixed with a two-direction
     correction loop; the fuzz inputs are kept as a regression corpus.
3. **Sexagenary names in both readings** (Bính Ngọ / 병오년) and the national
   holiday sets, including the VN/KR split on Buddha's Birthday (15/4 vs 8/4).

## Numerical fidelity

Every floor operation mirrors JavaScript `Math.floor` (toward −∞) so the Go
port is bit-identical with the production JS: 59,810 conversions and 1,732
low-level astronomy samples in `testdata/` must match exactly in CI. This is
deliberate: correctness here is defined as agreement with the implementation
that real users already rely on, then hardened (see above) where the reference
is provably wrong.

## Accuracy bounds

The truncated series and the simple ΔT polynomial are good for roughly
1200–3000 CE; outside that, errors in ΔT can shift a new moon across midnight
and thus a month boundary by a day. Cross-validation coverage is 1900–2199.
Korean dates are computed by the same method at UTC+9; they match the KST
convention used for Seollal/Chuseok, but this library is not an official
KASI almanac.

## Testing

- Golden tests: bit-equality with production JS (both zones, 300 years).
- Round-trip property: `LunarToSolar(SolarToLunar(d)) == d` for every golden
  date, plus a native Go fuzz target (`FuzzRoundTrip`) — 38.9M executions
  clean at the time of writing; crashing inputs are committed as regression
  seeds.
- Reference pinning: Tết, Trung thu, and Giỗ tổ civil dates 2024–2035 from the
  production almanac table.
- `go test -race`, benchmarks with `-benchmem` (conversions are 0 allocs/op).

## AI-assisted workflow

This repository was built with an AI coding agent (Claude) under human
direction and review: the agent ported, tested, and documented; the algorithm
choice, API design, validation semantics, and every commit were reviewed by
me. The golden files were generated from my pre-existing production
implementation, which is the source of truth the port had to match.
