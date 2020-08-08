package runtime

import (
	"errors"

	"github.com/dball/glimpse/ex"
	"github.com/dball/glimpse/types"
)

var (
	invalidType  = ex.Ex{Code: "Invalid type"}
	invalidValue = ex.Ex{Code: "Invalid value"}
)

// Seq returns a seq for seq or seqable values
func Seq(value types.MalType) (types.Seq, error) {
	var seq types.Seq
	switch tvalue := value.(type) {
	case types.Seq:
		seq = tvalue
	case types.Seqable:
		seq = tvalue.Seq()
	default:
		return nil, invalidType
	}
	empty, _, _ := seq.Next()
	if empty {
		return types.Nil{}, nil
	}
	return seq, nil
}

// Empty returns true or false if its argument is seqable and empty or not
func Empty(value types.MalType) (types.Boolean, error) {
	seq, err := Seq(value)
	if err != nil {
		return false, err
	}
	empty, _, _ := seq.Next()
	return types.Boolean(empty), nil
}

// TakeDrop returns as many as n items from the front of the seqable argument
func TakeDrop(n types.MalType, value types.MalType) (types.SliceSeq, types.Seq, error) {
	intN, valid := n.(types.Integer)
	if !valid {
		return types.SliceSeq{}, nil, invalidType
	}
	in := int64(intN)
	if in < 0 {
		return types.SliceSeq{}, nil, invalidValue
	}
	seq, err := Seq(value)
	if err != nil {
		return types.SliceSeq{}, nil, err
	}
	var items []types.MalType
	var i int64
	for {
		if i == in {
			break
		}
		empty, head, tail := seq.Next()
		if empty {
			break
		}
		items = append(items, head)
		seq = tail
		i++
	}
	return types.SliceSeq{Items: items}, seq, nil
}

// Concat returns a seqable of the non-empty seqs
func Concat(values ...types.MalType) (types.MalType, error) {
	var seqs []types.Seq
	for _, value := range values {
		seq, err := Seq(value)
		if err != nil {
			return nil, err
		}
		empty, _, _ := seq.Next()
		if !empty {
			seqs = append(seqs, seq)
		}
	}
	if len(seqs) == 0 {
		return types.NewList(), nil
	} else if len(seqs) == 1 {
		return seqs[0], nil
	}
	return types.Concatenation{Seqs: seqs}, nil
}

// Conj conjoins values to a collection
func Conj(coll types.MalType, values ...types.MalType) (types.Conjable, error) {
	conjable, valid := coll.(types.Conjable)
	if !valid {
		return nil, errors.New("Invalid conj target")
	}
	for _, value := range values {
		newconj, err := conjable.Conj(value)
		if err != nil {
			return nil, err
		}
		conjable = newconj
	}
	return conjable, nil
}

// Into pours a seqable into a collection
func Into(coll types.MalType, value types.MalType) (types.Conjable, error) {
	var values []types.MalType
	seq, err := Seq(value)
	if err != nil {
		return nil, err
	}
	for {
		empty, head, tail := seq.Next()
		if empty {
			break
		}
		values = append(values, head)
		seq = tail
	}
	return Conj(coll, values...)
}

// IntoEmptyVector is a convenience fn
func IntoEmptyVector(value types.MalType) types.Vector {
	coll, _ := Into(types.NewVector(), value)
	return coll.(types.Vector)
}

// IntoSlice pours a seq into a slice
func IntoSlice(value types.MalType) ([]types.MalType, error) {
	var values []types.MalType
	seq, err := Seq(value)
	if err != nil {
		return nil, err
	}
	for {
		empty, head, tail := seq.Next()
		if empty {
			break
		}
		values = append(values, head)
		seq = tail
	}
	return values, nil
}

// Nth returns the nth value in a seqable, if any
func Nth(value types.MalType, n types.MalType) (types.MalType, error) {
	seq, err := Seq(value)
	if err != nil {
		return nil, err
	}
	nint, valid := n.(types.Integer)
	if !valid {
		return err, invalidValue
	}
	nint64 := int64(nint)
	if nint64 < 0 {
		return nil, invalidValue
	}
	var i int64
	for {
		empty, head, tail := seq.Next()
		if empty {
			return nil, invalidValue
		}
		if i == nint64 {
			return head, nil
		}
		seq = tail
		i++
	}
}

// Get looks up a key in an indexed collection
func Get(coll types.MalType, index types.MalType, notfound types.MalType) types.MalType {
	indexed, valid := coll.(types.Indexed)
	if !valid {
		return notfound
	}
	value, found := indexed.Lookup(index)
	if !found {
		return notfound
	}
	return value
}

// Contains tests the existence of a mapping for a key in an indexed collection
func Contains(coll types.MalType, index types.MalType) types.Boolean {
	indexed, valid := coll.(types.Indexed)
	if !valid {
		return types.Boolean(false)
	}
	_, found := indexed.Lookup(index)
	return types.Boolean(found)
}

// Keys returns a list of keys in a map
func Keys(coll types.MalType) (types.List, error) {
	m, valid := coll.(types.Map)
	if !valid {
		return types.NewList(), invalidType
	}
	imm := m.Imm
	items := make([]types.MalType, imm.Len())
	i := 0
	itr := imm.Iterator()
	for !itr.Done() {
		k, _ := itr.Next()
		items[i] = k
	}
	return types.NewList(items...), nil
}

// Vals returns a list of keys in a map
func Vals(coll types.MalType) (types.List, error) {
	m, valid := coll.(types.Map)
	if !valid {
		return types.NewList(), invalidType
	}
	imm := m.Imm
	items := make([]types.MalType, imm.Len())
	i := 0
	itr := imm.Iterator()
	for !itr.Done() {
		_, v := itr.Next()
		items[i] = v
	}
	return types.NewList(items...), nil
}

// WithMeta adds metadata to a container
func WithMeta(container types.MalType, metadata types.MalType) (types.MalType, error) {
	hm, valid := container.(types.HasMetadata)
	if !valid {
		return nil, invalidType
	}
	md, valid := metadata.(types.Map)
	if !valid {
		return nil, invalidType
	}
	return hm.WithMetadata(md), nil
}

// Meta returns a container's metadata
func Meta(container types.MalType) (types.MalType, error) {
	hm, valid := container.(types.HasMetadata)
	if !valid {
		return nil, invalidType
	}
	md := hm.Metadata()
	if md.Imm == nil {
		return types.Nil{}, nil
	}
	return hm.Metadata(), nil
}

func Range(constraints ...types.MalType) (types.MalType, error) {
	ints := make([]int64, len(constraints))
	for i, constraint := range constraints {
		in, valid := constraint.(types.Integer)
		if !valid {
			return nil, invalidType
		}
		ints[i] = int64(in)
	}
	switch len(constraints) {
	case 0:
		return types.Range{Step: 1}, nil
	case 1:
		return types.Range{Upper: ints[0], Step: 1, Finite: true}, nil
	case 2:
		return types.Range{Lower: ints[0], NextValue: ints[0], Upper: ints[1], Step: 1, Finite: true}, nil
	case 3:
		return types.Range{Lower: ints[0], NextValue: ints[0], Upper: ints[1], Step: ints[2], Finite: true}, nil
	default:
		return nil, invalidValue
	}
}
