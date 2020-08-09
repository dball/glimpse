package types

// Concatenation is not empty
type Concatenation struct {
	Seqs []Seq
	Meta Map
}

// Seq of a concatenation is itself
func (c Concatenation) Seq() Seq {
	return c
}

// Next of a concatention finds the first nonempty value
func (c Concatenation) Next() (bool, MalType, Seq) {
	_, head, tail := c.Seqs[0].Next()
	tailEmpty, _, _ := tail.Next()
	if tailEmpty {
		if len(c.Seqs) == 1 {
			return false, head, Nil{}
		}
		if len(c.Seqs) == 2 {
			return false, head, c.Seqs[1]
		}
		return false, head, Concatenation{Seqs: c.Seqs[1:]}
	}
	seqs := make([]Seq, len(c.Seqs))
	seqs[0] = tail
	for i, seq := range c.Seqs {
		if i > 0 {
			seqs[i] = seq
		}
	}
	return false, head, Concatenation{Seqs: seqs}
}

// Sequential are concatenations
func (Concatenation) Sequential() {}

// Applicable are concatenations
func (Concatenation) Applicable() {}

// Conj to a concatenation prepends as a cons
func (c Concatenation) Conj(value MalType) (Conjable, error) {
	return ConsCell{Head: value, Tail: c}, nil
}

// Metadata for a concatenation
func (c Concatenation) Metadata() Map {
	return c.Meta
}

// WithMetadata for a concatenation
func (c Concatenation) WithMetadata(m Map) HasMetadata {
	return Concatenation{Seqs: c.Seqs, Meta: m}
}
