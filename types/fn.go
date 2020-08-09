package types

// Function - functions of values to value
type Function struct {
	Fn      func(...MalType) (MalType, error)
	Body    MalType
	Binds   []MalType
	Env     *Env
	IsMacro bool
	Meta    Map
}

// Metadata for a fn
func (fn Function) Metadata() Map {
	return fn.Meta
}

// WithMetadata for a fn
func (fn Function) WithMetadata(m Map) HasMetadata {
	return Function{Fn: fn.Fn, Body: fn.Body, Binds: fn.Binds, Env: fn.Env, IsMacro: fn.IsMacro, Meta: m}
}
