package types

import "github.com/benbjohnson/immutable"

// Map is an immutable map
type Map struct {
	Imm  *immutable.Map
	Meta *Map
}

// NewMap builds a new map
func NewMap(values ...MalType) Map {
	imm := immutable.NewMap(hasher{})
	if len(values) > 0 {
		b := immutable.NewMapBuilder(imm)
		for i := 0; i < len(values); i += 2 {
			b.Set(values[i], values[i+1])
		}
		imm = b.Map()
	}
	return Map{Imm: imm}
}

// Seq traverses map entries
func (m Map) Seq() Seq {
	entries := make([]MalType, m.Imm.Len())
	var i int64
	itr := m.Imm.Iterator()
	for !itr.Done() {
		k, v := itr.Next()
		entries[i] = NewVector(k, v)
		i++
	}
	return buildSeqFromSlice(entries)
}

// Count counts map entries
func (m Map) Count() int {
	return m.Imm.Len()
}

// Lookup in a map returns the value
func (m Map) Lookup(index MalType) (MalType, bool) {
	return m.Imm.Get(index)
}

// Metadata for a map
func (m Map) Metadata() Map {
	return *(m.Meta)
}

// WithMetadata for a map
func (m Map) WithMetadata(md Map) HasMetadata {
	return Map{Imm: m.Imm, Meta: &md}
}
