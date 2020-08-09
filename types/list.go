package types

import "github.com/benbjohnson/immutable"

// List - sequences of mal values
type List struct {
	Imm  *immutable.List
	Meta Map
}

// NewList builds a new list
func NewList(items ...MalType) List {
	imm := immutable.NewList()
	if len(items) > 0 {
		b := immutable.NewListBuilder(imm)
		for _, v := range items {
			b.Append(v)
		}
		imm = b.List()
	}
	return List{Imm: imm}
}

// Sequential lists
func (List) Sequential() {}

// Applicable lists
func (List) Applicable() {}

// Seq traverses list items
func (list List) Seq() Seq {
	return ListIteratorSeq{Imm: list.Imm}
}

// Count counts list items
func (list List) Count() int {
	return list.Imm.Len()
}

// Conj prepends to lists
func (list List) Conj(value MalType) (Conjable, error) {
	return List{Imm: list.Imm.Prepend(value)}, nil
}

// Metadata for a list
func (list List) Metadata() Map {
	return list.Meta
}

// WithMetadata for a list
func (list List) WithMetadata(m Map) HasMetadata {
	return List{Imm: list.Imm, Meta: m}
}
