package types

// Rune - mal rune values
type Rune rune

// ValueEquals compares runes
func (r Rune) ValueEquals(that MalType) bool {
	thatRune, valid := that.(Rune)
	if !valid {
		return false
	}
	return r == thatRune
}

func (r Rune) hashBytes() []byte {
	return []byte(string(r))
}
