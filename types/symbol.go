package types

// Symbol - mal symbol values
type Symbol struct {
	Name string
	Meta Map
}

// NewSymbol builds a new symbol
func NewSymbol(name string) Symbol {
	return Symbol{Name: name}
}

// ValueEquals compares symbols
func (symbol Symbol) ValueEquals(that MalType) bool {
	thatSymbol, valid := that.(Symbol)
	if !valid {
		return false
	}
	return symbol.Name == thatSymbol.Name
}

func (symbol Symbol) hashBytes() []byte {
	return append([]byte(symbol.Name), byte('\''))
}

// Metadata for a symbol
func (symbol Symbol) Metadata() Map {
	return symbol.Meta
}

// WithMetadata symbols
func (symbol Symbol) WithMetadata(m Map) HasMetadata {
	return Symbol{Name: symbol.Name, Meta: m}
}
