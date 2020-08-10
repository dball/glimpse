package types

import (
	"encoding/binary"
	"unsafe"
)

// Atom - mal atom values
type Atom struct {
	Value MalType
}

// Set the value of an atom
func (a *Atom) Set(value MalType) {
	a.Value = value
}

// ValueEquals checks pointer equality
func (a *Atom) ValueEquals(that MalType) bool {
	thatAtom, valid := that.(*Atom)
	if !valid {
		return false
	}
	return a == thatAtom
}

func (a *Atom) hashBytes() []byte {
	p := unsafe.Pointer(a)
	up := uintptr(p)
	ui := uint64(up)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, ui)
	return b[:]
}
