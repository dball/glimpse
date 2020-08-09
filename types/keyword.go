package types

// Keyword - mal keyword values
type Keyword struct {
	Name string
}

// NewKeyword builds a new keyword
func NewKeyword(name string) Keyword {
	return Keyword{Name: name}
}

// ValueEquals compares keywords
func (keyword Keyword) ValueEquals(that MalType) bool {
	thatKeyword, valid := that.(Keyword)
	if !valid {
		return false
	}
	return keyword.Name == thatKeyword.Name
}

func (keyword Keyword) hashBytes() []byte {
	return append([]byte(keyword.Name), byte(':'))
}
