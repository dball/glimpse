package types

// Boolean - mal boolean values
type Boolean bool

// ValueEquals compares booleans
func (boolean Boolean) ValueEquals(that MalType) bool {
	thatBoolean, valid := that.(Boolean)
	if !valid {
		return false
	}
	return boolean == thatBoolean
}

func (boolean Boolean) hashBytes() []byte {
	var byt byte
	if bool(boolean) {
		byt = byte(1)
	} else {
		byt = byte(2)
	}
	b := [1]byte{byt}
	return b[:]
}
