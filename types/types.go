package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"

	"github.com/benbjohnson/immutable"
	"github.com/spaolacci/murmur3"
)

// MalType - the root type of all mal values
type MalType interface{}

// Counted - collections that have finite, known size
type Counted interface {
	Count() int
}

// Seqable - collections that can produce a traversing sequence, or are empty
type Seqable interface {
	Seq() Seq
}

// Sequential - collections that have an order
type Sequential interface {
	Seqable
	Sequential()
}

// Seq - a traversal of a sequence
type Seq interface {
	// Next returns a tuple of emptiness, the first item if non-empty, and the tail seq
	Next() (bool, MalType, Seq)
}

// HasSimpleValueEquality - is a type which can compare itself to other values
type HasSimpleValueEquality interface {
	ValueEquals(MalType) bool
	hashBytes() []byte
}

// HasMetadata - the type has metadata
type HasMetadata interface {
	Metadata() Map
	WithMetadata(m Map) HasMetadata
}

// Applicable - a list-y form that applies the rest to the first when evaluated
type Applicable interface {
	Sequential
	Applicable()
}

// Conjable - collection supports conjoining
type Conjable interface {
	Conj(MalType) (Conjable, error)
}

// Indexed - collection supports lookup by key
type Indexed interface {
	Lookup(MalType) (MalType, bool)
}

// SliceSeq traverses slices of items
type SliceSeq struct {
	Items []MalType
	Meta  Map
}

// Next from a slice seq
func (seq SliceSeq) Next() (bool, MalType, Seq) {
	if len(seq.Items) == 0 {
		return true, nil, nil
	}
	return false, seq.Items[0], SliceSeq{Items: seq.Items[1:]}
}

// Count counts the items in a slice
func (seq SliceSeq) Count() int {
	return len(seq.Items)
}

// Metadata seqs have metadata
func (seq SliceSeq) Metadata() Map {
	return seq.Meta
}

// WithMetadata seqs have metadata
func (seq SliceSeq) WithMetadata(m Map) HasMetadata {
	return SliceSeq{Items: seq.Items, Meta: m}
}

func buildSeqFromSlice(items []MalType) Seq {
	if len(items) == 0 {
		return Nil{}
	}
	return SliceSeq{Items: items}
}

// ListIteratorSeq seqs over immutable Lists
type ListIteratorSeq struct {
	Imm       *immutable.List
	NextIndex int
	Meta      Map
}

// Next for a list
func (seq ListIteratorSeq) Next() (bool, MalType, Seq) {
	if seq.NextIndex >= seq.Imm.Len() {
		return true, nil, nil
	}
	head := seq.Imm.Get(seq.NextIndex)
	tail := ListIteratorSeq{Imm: seq.Imm, NextIndex: seq.NextIndex + 1}
	return false, head, tail
}

// Metadata for a list
func (seq ListIteratorSeq) Metadata() Map {
	return seq.Meta
}

// WithMetadata seqs have metadata
func (seq ListIteratorSeq) WithMetadata(m Map) HasMetadata {
	return ListIteratorSeq{Imm: seq.Imm, NextIndex: seq.NextIndex, Meta: m}
}

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

// Integer - mal integer values
type Integer int64

// ValueEquals compares integers
func (i Integer) ValueEquals(that MalType) bool {
	thatInt, valid := that.(Integer)
	if !valid {
		return false
	}
	return i == thatInt
}

func (i Integer) hashBytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b[:]
}

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

// Keyword - mal keyword values
type Keyword struct {
	Name string
}

// NewKeyword builds a new keyword
func NewKeyword(name string) Keyword {
	return Keyword{Name: name}
}

// ValueEquals compares keywords
func (keyword Keyword) ValueEquals(that MalType) bool {
	thatKeyword, valid := that.(Keyword)
	if !valid {
		return false
	}
	return keyword.Name == thatKeyword.Name
}

func (keyword Keyword) hashBytes() []byte {
	return append([]byte(keyword.Name), byte(':'))
}

// String - mal string values
type String string

// ValueEquals compared strings
func (s String) ValueEquals(that MalType) bool {
	thatString, valid := that.(String)
	if !valid {
		return false
	}
	return s == thatString
}

func (s String) hashBytes() []byte {
	return []byte(s)
}

func hashAnyValue(hash *hash.Hash32, value *MalType) {
	switch cast := (*value).(type) {
	case HasSimpleValueEquality:
		(*hash).Write(cast.hashBytes())
	case Sequential:
		seq := cast.Seq()
		for {
			empty, head, tail := seq.Next()
			if empty {
				break
			}
			hashAnyValue(hash, &head)
			seq = tail
		}
	case Map:
		(*hash).Write([]byte("{}"))
		seq := cast.Seq()
		for {
			empty, head, tail := seq.Next()
			if empty {
				break
			}
			hashAnyValue(hash, &head)
			seq = tail
		}
	default:
		// TODO hash the pointer address for instance identity
	}
}

// Hash computes a murmur3 hash of the given value
func Hash(value MalType) uint32 {
	hash := murmur3.New32()
	hashAnyValue(&hash, &value)
	return hash.Sum32()
}

type hasher struct{}

func (h hasher) Hash(key interface{}) uint32 {
	return Hash(key)
}

func (h hasher) Equal(a, b interface{}) bool {
	return Equals(a, b)
}

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

// Atom - mal atom values
type Atom struct {
	Value MalType
}

// Set the value of an atom
func (a *Atom) Set(value MalType) {
	a.Value = value
}

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

// Boolean - mal boolean values
type Boolean bool

// ValueEquals compares booleans
func (boolean Boolean) ValueEquals(that MalType) bool {
	thatBoolean, valid := that.(Boolean)
	if !valid {
		return false
	}
	return boolean == thatBoolean
}

func (boolean Boolean) hashBytes() []byte {
	var byt byte
	if bool(boolean) {
		byt = byte(1)
	} else {
		byt = byte(2)
	}
	b := [1]byte{byt}
	return b[:]
}

// Nil - mal nil value
type Nil struct{}

// Seq traverses nothing
func (Nil) Seq() Seq {
	return Nil{}
}

// ValueEquals compares nils
func (malnil Nil) ValueEquals(that MalType) bool {
	_, valid := that.(Nil)
	return valid
}

func (malnil Nil) hashBytes() []byte {
	b := [1]byte{byte(0)}
	return b[:]
}

// Count of nothing is 0
func (malnil Nil) Count() int {
	return 0
}

// Next of a nil seq is empty
func (malnil Nil) Next() (bool, MalType, Seq) {
	return true, nil, nil
}

// Conj onto nil returns a list
func (malnil Nil) Conj(value MalType) (Conjable, error) {
	return NewList(value), nil
}

// Lookup in nil always fails
func (malnil Nil) Lookup(value MalType) (MalType, bool) {
	return nil, false
}

type Range struct {
	Lower     int64
	Upper     int64
	Step      int64
	Finite    bool
	NextValue int64
}

func (r Range) Seq() Seq {
	return r
}

func (r Range) Next() (bool, MalType, Seq) {
	if r.Finite && r.NextValue >= r.Upper {
		return true, nil, nil
	}
	head := Integer(r.NextValue)
	tail := Range{Lower: r.Lower, Upper: r.Upper, Step: r.Step, Finite: r.Finite, NextValue: r.NextValue + r.Step}
	return false, head, tail
}

// Equals compares values
func Equals(this MalType, that MalType) bool {
	switch cast := this.(type) {
	case HasSimpleValueEquality:
		return cast.ValueEquals(that)
	case Map:
		thatMap, valid := that.(Map)
		if !valid {
			return false
		}
		thisImm := cast.Imm
		thatImm := thatMap.Imm
		if thisImm.Len() != thatImm.Len() {
			return false
		}
		itr := thisImm.Iterator()
		for !itr.Done() {
			k, v := itr.Next()
			v2, found := thatImm.Get(k)
			if !found {
				return false
			}
			if !Equals(v, v2) {
				return false
			}
		}
		return true
	case Sequential:
		thatSequential, valid := that.(Sequential)
		if !valid {
			return false
		}
		thisSeq := cast.Seq()
		thatSeq := thatSequential.Seq()
		for {
			thisEmpty, thisHead, thisTail := thisSeq.Next()
			thatEmpty, thatHead, thatTail := thatSeq.Next()
			if thisEmpty && thatEmpty {
				return true
			}
			if thisEmpty || thatEmpty {
				return false
			}
			if !Equals(thisHead, thatHead) {
				return false
			}
			thisSeq = thisTail
			thatSeq = thatTail
		}
	default:
		return false
	}
}

// Compare compares values
func Compare(this MalType, that MalType) (int8, error) {
	switch cast := this.(type) {
	case Integer:
		thatInt, valid := that.(Integer)
		if !valid {
			return 0, errors.New("Incomparable values")
		}
		delta := cast - thatInt
		if delta > 0 {
			return 1, nil
		} else if delta == 0 {
			return 0, nil
		}
		return -1, nil
	default:
		return 0, errors.New("Incomparable values")
	}
}

// MalError contains any mal reason
type MalError struct {
	Reason MalType
}

func (err MalError) Error() string {
	return fmt.Sprintf("%v", err.Reason)
}

// Undefined errors
type Undefined struct {
	Name string
}

func (err Undefined) Error() string {
	return fmt.Sprintf("'%v' not found", err.Name)
}

// Env binds names to values
type Env struct {
	Outer    *Env
	Bindings map[string]MalType
}

// Set sets the value of a symbol
func (env Env) Set(name string, value MalType) {
	env.Bindings[name] = value
}

// Get gets the value of a symbol
func (env Env) Get(name string) (MalType, error) {
	value, found := env.Bindings[name]
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
	return &Env{Bindings: make(map[string]MalType)}
}

// DeriveEnv derives an env
func DeriveEnv(Outer *Env, binds, exprs []MalType) (*Env, error) {
	env := BuildEnv()
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
