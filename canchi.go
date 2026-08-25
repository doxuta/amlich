package amlich

import "fmt"

// The sexagenary cycle (can-chi / 간지 / 干支): ten heavenly stems combined
// with twelve earthly branches, naming years in a 60-year cycle. Both Vietnam
// and Korea use the same cycle; only the readings differ.

var (
	stemsVI    = [10]string{"Giáp", "Ất", "Bính", "Đinh", "Mậu", "Kỷ", "Canh", "Tân", "Nhâm", "Quý"}
	branchesVI = [12]string{"Tý", "Sửu", "Dần", "Mão", "Thìn", "Tỵ", "Ngọ", "Mùi", "Thân", "Dậu", "Tuất", "Hợi"}
	stemsKO    = [10]string{"갑", "을", "병", "정", "무", "기", "경", "신", "임", "계"}
	branchesKO = [12]string{"자", "축", "인", "묘", "진", "사", "오", "미", "신", "유", "술", "해"}
	animalsEN  = [12]string{"Rat", "Ox", "Tiger", "Cat", "Dragon", "Snake", "Horse", "Goat", "Monkey", "Rooster", "Dog", "Pig"}
)

// YearName is the sexagenary name of a lunar year in both readings.
type YearName struct {
	Vietnamese string // e.g. "Bính Ngọ"
	Korean     string // e.g. "병오년"
	AnimalEN   string // e.g. "Horse" — note: branch 4 (Mão) is Cat in Vietnam, Rabbit elsewhere
}

// CanChi returns the sexagenary name of the given lunar year.
func CanChi(lunarYear int) YearName {
	s := mod(lunarYear+6, 10)
	b := mod(lunarYear+8, 12)
	return YearName{
		Vietnamese: fmt.Sprintf("%s %s", stemsVI[s], branchesVI[b]),
		Korean:     stemsKO[s] + branchesKO[b] + "년",
		AnimalEN:   animalsEN[b],
	}
}

func mod(a, n int) int {
	m := a % n
	if m < 0 {
		m += n
	}
	return m
}
