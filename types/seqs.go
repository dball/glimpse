package types

import "github.com/benbjohnson/immutable"

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
