package types

import "encoding/binary"

// Integer - mal integer values
type Integer int64

// ValueEquals compares integers
func (i Integer) ValueEquals(that MalType) bool {
	thatInt, valid := that.(Integer)
	if !valid {
		return false
	}
	return i == thatInt
}

func (i Integer) hashBytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b[:]
}
