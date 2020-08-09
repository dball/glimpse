package types

// ConsCell - a seq of a head prepended to a tail seq
type ConsCell struct {
	Head MalType
	Tail Seq
	Meta Map
}

// Seq of a cons is itself
func (c ConsCell) Seq() Seq {
	return c
}

// Next of a cons just decomposes it
func (c ConsCell) Next() (bool, MalType, Seq) {
	return false, c.Head, c.Tail
}

// Sequential are cons
func (ConsCell) Sequential() {}

// Applicable are cons
func (ConsCell) Applicable() {}

// Conj prepends a cons
func (c ConsCell) Conj(value MalType) (Conjable, error) {
	return ConsCell{Head: value, Tail: c}, nil
}

// Metadata for a cons
func (c ConsCell) Metadata() Map {
	return c.Meta
}

// WithMetadata for a cons
func (c ConsCell) WithMetadata(m Map) HasMetadata {
	return ConsCell{Head: c.Head, Tail: c.Tail, Meta: m}
}
