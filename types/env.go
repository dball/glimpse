package types

import (
	"errors"

	"github.com/benbjohnson/immutable"
)

// Env binds names to values
type Env struct {
	Outer    *Env
	Bindings *immutable.Map
}

// Set sets the value of a symbol
func (env *Env) Set(name string, value MalType) {
	env.Bindings = env.Bindings.Set(name, value)
}

// Get gets the value of a symbol
func (env *Env) Get(name string) (MalType, error) {
	value, found := env.Bindings.Get(name)
	if !found {
		if env.Outer == nil {
			return nil, Undefined{Name: name}
		}
		outer := *env.Outer
		return outer.Get(name)
	}
	return value, nil
}

// BuildEnv builds a new env
func BuildEnv() *Env {
	return &Env{Bindings: immutable.NewMap(nil)}
}

// DeriveEnv derives an env
func DeriveEnv(Outer *Env, binds, exprs []MalType) (*Env, error) {
	env := BuildEnv()
	env.Bindings = Outer.Bindings
	env.Outer = Outer
	var bindSymbols []Symbol
	for _, bind := range binds {
		bindSymbol, valid := bind.(Symbol)
		if !valid {
			return nil, errors.New("binds must be symbols")
		}
		bindSymbols = append(bindSymbols, bindSymbol)
	}
	varargs := len(bindSymbols) >= 2 && bindSymbols[len(bindSymbols)-2].Name == "&"
	var varargSymbol Symbol
	if varargs {
		varargSymbol = bindSymbols[len(bindSymbols)-1]
		bindSymbols = bindSymbols[0 : len(bindSymbols)-2]
	}
	for i, bind := range bindSymbols {
		if i >= len(exprs) {
			return nil, errors.New("no expr for bind")
		}
		env.Set(bind.Name, exprs[i])
	}
	if varargs {
		list := NewList(exprs[len(bindSymbols):]...)
		env.Set(varargSymbol.Name, list)
	}
	return env, nil
}
