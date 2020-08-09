package types

// String - mal string values
type String string

// ValueEquals compared strings
func (s String) ValueEquals(that MalType) bool {
	thatString, valid := that.(String)
	if !valid {
		return false
	}
	return s == thatString
}

func (s String) hashBytes() []byte {
	return []byte(s)
}

// Seq of string is a seq of runes
func (s String) Seq() Seq {
	// TODO is this the most efficient cast really?
	runes := []rune(s)
	items := make([]MalType, len(runes))
	for i, r := range runes {
		items[i] = Rune(r)
	}
	list := NewList(items...)
	return list.Seq()
}
