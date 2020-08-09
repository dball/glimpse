package types

// Range is a seq over integers
type Range struct {
	Lower  int64
	Upper  int64
	Step   int64
	Finite bool
}

// Seq of a range is itself
func (r Range) Seq() Seq {
	return r
}

// Next of a range increments Lower unless it's finite and finished
func (r Range) Next() (bool, MalType, Seq) {
	if r.Finite && r.Lower >= r.Upper {
		return true, nil, nil
	}
	head := Integer(r.Lower)
	tail := Range{Lower: r.Lower + r.Step, Upper: r.Upper, Step: r.Step, Finite: r.Finite}
	return false, head, tail
}
