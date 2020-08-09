package types

// Atom - mal atom values
type Atom struct {
	Value MalType
}

// Set the value of an atom
func (a *Atom) Set(value MalType) {
	a.Value = value
}
