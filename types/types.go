package types

import (
	"errors"
	"fmt"
	"hash"

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
