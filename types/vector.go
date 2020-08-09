package types

import "github.com/benbjohnson/immutable"

// Vector - sequences of mal values
type Vector struct {
	Imm  *immutable.List
	Meta Map
}

// NewVector builds a new vector
func NewVector(items ...MalType) Vector {
	imm := immutable.NewList()
	if len(items) > 0 {
		b := immutable.NewListBuilder(imm)
		for _, v := range items {
			b.Append(v)
		}
		imm = b.List()
	}
	return Vector{Imm: imm}
}

// Sequential vectors
func (Vector) Sequential() {}

// Seq traverses vector items
func (vector Vector) Seq() Seq {
	return ListIteratorSeq{Imm: vector.Imm}
}

// Count counts vector items
func (vector Vector) Count() int {
	return vector.Imm.Len()
}

// Conj appends to a vector
func (vector Vector) Conj(value MalType) (Conjable, error) {
	return Vector{Imm: vector.Imm.Append(value)}, nil
}

// Lookup looks up in a vector by position
func (vector Vector) Lookup(index MalType) (MalType, bool) {
	i, valid := index.(Integer)
	if !valid {
		return nil, false
	}
	ii := int(i)
	if ii < 0 || ii >= vector.Imm.Len() {
		return nil, false
	}
	return vector.Imm.Get(ii), true
}

// Metadata for a vector
func (vector Vector) Metadata() Map {
	return vector.Meta
}

// WithMetadata for a vector
func (vector Vector) WithMetadata(m Map) HasMetadata {
	return Vector{Imm: vector.Imm, Meta: m}
}
