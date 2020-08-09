package types

// Nil - mal nil value
type Nil struct{}

// Seq traverses nothing
func (Nil) Seq() Seq {
	return Nil{}
}

// ValueEquals compares nils
func (malnil Nil) ValueEquals(that MalType) bool {
	_, valid := that.(Nil)
	return valid
}

func (malnil Nil) hashBytes() []byte {
	b := [1]byte{byte(0)}
	return b[:]
}

// Count of nothing is 0
func (malnil Nil) Count() int {
	return 0
}

// Next of a nil seq is empty
func (malnil Nil) Next() (bool, MalType, Seq) {
	return true, nil, nil
}

// Conj onto nil returns a list
func (malnil Nil) Conj(value MalType) (Conjable, error) {
	return NewList(value), nil
}

// Lookup in nil always fails
func (malnil Nil) Lookup(value MalType) (MalType, bool) {
	return nil, false
}
