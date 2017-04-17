// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// Package ctmap implements a constant-time key-value map.
package ctmap

import "crypto/subtle"

// Map is a constant-time key-value map.
type Map struct {
	m [][]byte

	keySize, valSize int
}

// New returns a new constant-time map with the
// given key and value sizes.
//
// Every key and value must be of equal size.
//
// For a constant-time equivalent of map[string]struct{},
// use 0 for valSize.
func New(keySize, valSize int) *Map {
	return &Map{
		keySize: keySize,
		valSize: valSize,
	}
}

// NewWithCapacity returns a new constant-time map with
// the given key and value sizes. It preallocates the
// map's backing storage, sized to fit capacity entries
// without reallocating.
//
// Every key and value must be of equal size.
//
// For a constant-time equivalent of map[string]struct{},
// use 0 for valSize.
func NewWithCapacity(keySize, valSize, capacity int) *Map {
	return &Map{
		m: make([][]byte, 0, capacity),

		keySize: keySize,
		valSize: valSize,
	}
}

// Len returns the number of entries in the map. It does
// not account for duplicates.
func (m *Map) Len() int {
	return len(m.m)
}

// Add appends a new entry to the map. It does not check
// for duplicates, nor does it handle them.
func (m *Map) Add(key, val []byte) {
	if len(key) != m.keySize {
		panic("key has invalid size")
	}

	if len(val) != m.valSize {
		panic("val has invalid size")
	}

	entry := make([]byte, m.keySize+m.valSize)
	copy(entry[:m.keySize], key)
	copy(entry[m.keySize:], val)

	m.m = append(m.m, entry)
}

// Set sets an existing map entry to val in constant-time.
// It returns 1 if the key was present and the val set, 0
// otherwise.
//
// If there are multiple entries with the same key, each
// will have it's value set to val.
func (m *Map) Set(key, val []byte) int {
	if len(key) != m.keySize {
		panic("key has invalid size")
	}

	if len(val) != m.valSize {
		panic("val has invalid size")
	}

	var v int

	for _, entry := range m.m {
		vv := subtle.ConstantTimeCompare(entry[:m.keySize], key)
		subtle.ConstantTimeCopy(vv, entry[m.keySize:], val)
		v |= vv
	}

	return v
}

// Replace replaces an existing entry in the map with a new
// entry in constant-time. It returns 1 if oldKey was present
// in the map and the entry replaced, 0 otherwise.
//
// If there are duplicate entries matching oldKey, only the
// first entry will be replaced.
func (m *Map) Replace(oldKey, newKey, val []byte) int {
	if len(oldKey) != m.keySize {
		panic("oldKey has invalid size")
	}

	if len(newKey) != m.keySize {
		panic("newKey has invalid size")
	}

	if len(val) != m.valSize {
		panic("val has invalid size")
	}

	var v int

	for _, entry := range m.m {
		vv := subtle.ConstantTimeCompare(entry[:m.keySize], oldKey) &^ v
		subtle.ConstantTimeCopy(vv, entry[:m.keySize], newKey)
		subtle.ConstantTimeCopy(vv, entry[m.keySize:], val)
		v |= vv
	}

	return v
}

// Rename is like Replace but it only changes they key and
// leaves the value untouched.
//
// If there are duplicate entries matching oldKey, only the
// first entry's key will be replaced.
func (m *Map) Rename(oldKey, newKey []byte) int {
	if len(oldKey) != m.keySize {
		panic("oldKey has invalid size")
	}

	if len(newKey) != m.keySize {
		panic("newKey has invalid size")
	}

	var v int

	for _, entry := range m.m {
		vv := subtle.ConstantTimeCompare(entry[:m.keySize], oldKey) &^ v
		subtle.ConstantTimeCopy(vv, entry[:m.keySize], newKey)
		v |= vv
	}

	return v
}

// Contains determines if a key is present in the map in
// constant-time. It returns 1 if the key is present, 0
// otherwise.
func (m *Map) Contains(key []byte) int {
	if len(key) != m.keySize {
		panic("key has invalid size")
	}

	var v int

	for _, entry := range m.m {
		v |= subtle.ConstantTimeCompare(entry[:m.keySize], key)
	}

	return v
}

// Lookup finds the value associated with a key in
// constant-time. The value is copied, in constant-time,
// into val which must be the correct length. It returns
// 1 if the key was present, 0 otherwise.
//
// If there are multiple entries matching key, only the
// first will be returned.
func (m *Map) Lookup(key, val []byte) int {
	if len(key) != m.keySize {
		panic("key has invalid size")
	}

	if len(val) != m.valSize {
		panic("val has invalid size")
	}

	var v int

	for _, entry := range m.m {
		vv := subtle.ConstantTimeCompare(entry[:m.keySize], key) &^ v
		subtle.ConstantTimeCopy(vv, val, entry[m.keySize:])
		v |= vv
	}

	return v
}

// Delete removes an entry with a given key from the map.
// It returns 1 if an entry was removed, 0 otherwise. The
// removed entry is zeroed.
//
// If the map contains multiple entries with the same key,
// only the first is removed.
//
// WARNING: Future calls to Add may leak the result of
// Delete. To avoid this leak, use Replace instead with
// zero, or sentinel, keys and values.
func (m *Map) Delete(key []byte) int {
	if len(key) != m.keySize {
		panic("key has invalid size")
	}

	if len(m.m) == 0 {
		return 0
	}

	var v int

	for i, entry := range m.m[:len(m.m)-1] {
		v |= subtle.ConstantTimeCompare(entry[:m.keySize], key)
		subtle.ConstantTimeCopy(v, entry, m.m[i+1])
	}

	last := m.m[len(m.m)-1]
	v |= subtle.ConstantTimeCompare(last[:m.keySize], key)

	for i := range last {
		last[i] &= byte(v - 1)
	}

	// The last entry in the list will not be garbage
	// collected until the next call to Add, this leaks
	// information about whether a key was removed or not.
	// Allowing the entry to be garbage collected now, by
	// setting the final entry to nil, would leak
	// information. It also cannot be done in constant-time.
	// Because of this, a memory leak is allowed to occur.
	// Even though m.m is truncated bellow, it still contains
	// a reference to the removed slice which prevents it
	// from being garbage collected. When Add is next called,
	// the append call will overwrite the reference and the
	// slice will be garbage collected resulting in two
	// separate timing leaks. One from the lack of need for
	// append to allocate a larger m.m slice and from the
	// eventual garbage collection. If Map was created with
	// NewWithCapacity, the append call in Add may not leak
	// any information in and of itself. That still leaves
	// the information leak when the garbage collector runs.

	// XXX: Hopefully this slice is constant-time.
	m.m = m.m[:len(m.m)-v]
	return v
}
