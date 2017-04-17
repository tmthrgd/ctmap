// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package ctmap

import (
	"reflect"
	"testing"
)

func TestLen(t *testing.T) {
	for _, c := range []struct {
		m   [][]byte
		len int
	}{
		{nil, 0},
		{[][]byte{}, 0},
		{[][]byte{{0, 0}}, 1},
		{[][]byte{{0xa5, 0x5a}}, 1},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, 3},
		{[][]byte{{0xa5, 0x5a}, {0x5a, 0xa5}}, 2},
	} {
		m := &Map{m: c.m, keySize: 1, valSize: 1}

		if l := m.Len(); l != c.len {
			t.Error("Len failed")
			t.Logf("expected: %d", c.len)
			t.Logf("got:      %d", l)
			t.Fatal()
		}
	}
}

func TestAdd(t *testing.T) {
	for _, c := range []struct{ before, after [][]byte }{
		{nil, [][]byte{{0xa5, 0x5a}}},
		{[][]byte{{0, 0}}, [][]byte{{0, 0}, {0xa5, 0x5a}}},
		{[][]byte{{0xa5, 0x5a}}, [][]byte{{0xa5, 0x5a}, {0xa5, 0x5a}}},
	} {
		m := &Map{m: c.before, keySize: 1, valSize: 1}
		m.Add([]byte{0xa5}, []byte{0x5a})

		if !reflect.DeepEqual(m.m, c.after) {
			t.Error("Add failed")
			t.Logf("expected: %02x", c.after)
			t.Logf("got:      %02x", m.m)
			t.Fatal()
		}
	}
}

func TestSet(t *testing.T) {
	for _, c := range []struct{ before, after [][]byte }{
		{nil, nil},
		{[][]byte{{0, 0}}, [][]byte{{0, 0}}},
		{[][]byte{{0xa5, 0x5a}}, [][]byte{{0xa5, 0xff}}},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, [][]byte{{0, 0}, {0xa5, 0xff}, {0, 0}}},
		{[][]byte{{0xa5, 0x5a}, {0xa5, 0x5a}}, [][]byte{{0xa5, 0xff}, {0xa5, 0xff}}},
	} {
		m := &Map{m: c.before, keySize: 1, valSize: 1}
		m.Set([]byte{0xa5}, []byte{0xff})

		if !reflect.DeepEqual(m.m, c.after) {
			t.Error("Set failed")
			t.Logf("expected: %02x", c.after)
			t.Logf("got:      %02x", m.m)
			t.Fatal()
		}
	}
}

func TestReplace(t *testing.T) {
	for _, c := range []struct{ before, after [][]byte }{
		{nil, nil},
		{[][]byte{{0, 0}}, [][]byte{{0, 0}}},
		{[][]byte{{0xa5, 0x5a}}, [][]byte{{0x5a, 0xff}}},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, [][]byte{{0, 0}, {0x5a, 0xff}, {0, 0}}},
		{[][]byte{{0xa5, 0x5a}, {0xa5, 0x5a}}, [][]byte{{0x5a, 0xff}, {0xa5, 0x5a}}},
	} {
		m := &Map{m: c.before, keySize: 1, valSize: 1}
		m.Replace([]byte{0xa5}, []byte{0x5a}, []byte{0xff})

		if !reflect.DeepEqual(m.m, c.after) {
			t.Error("Replace failed")
			t.Logf("expected: %02x", c.after)
			t.Logf("got:      %02x", m.m)
			t.Fatal()
		}
	}
}

func TestRename(t *testing.T) {
	for _, c := range []struct{ before, after [][]byte }{
		{nil, nil},
		{[][]byte{{0, 0}}, [][]byte{{0, 0}}},
		{[][]byte{{0xa5, 0x5a}}, [][]byte{{0x5a, 0x5a}}},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, [][]byte{{0, 0}, {0x5a, 0x5a}, {0, 0}}},
		{[][]byte{{0xa5, 0x5a}, {0xa5, 0x5a}}, [][]byte{{0x5a, 0x5a}, {0xa5, 0x5a}}},
	} {
		m := &Map{m: c.before, keySize: 1, valSize: 1}
		m.Rename([]byte{0xa5}, []byte{0x5a})

		if !reflect.DeepEqual(m.m, c.after) {
			t.Error("Rename failed")
			t.Logf("expected: %02x", c.after)
			t.Logf("got:      %02x", m.m)
			t.Fatal()
		}
	}
}

func TestContains(t *testing.T) {
	for _, c := range []struct {
		m [][]byte
		v int
	}{
		{nil, 0},
		{[][]byte{{0, 0}}, 0},
		{[][]byte{{0xa5, 0x5a}}, 1},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, 1},
		{[][]byte{{0xa5, 0x5a}, {0xa5, 0x5a}}, 1},
	} {
		m := &Map{m: c.m, keySize: 1, valSize: 1}
		v := m.Contains([]byte{0xa5})

		if v != c.v {
			t.Error("Contains failed")
			t.Logf("expected: %d", c.v)
			t.Logf("got:      %d", v)
			t.Fatal()
		}
	}
}

func TestLookup(t *testing.T) {
	for _, c := range []struct {
		m   [][]byte
		v   int
		val []byte
	}{
		{nil, 0, []byte{0}},
		{[][]byte{{0, 0}}, 0, []byte{0}},
		{[][]byte{{0xa5, 0x5a}}, 1, []byte{0x5a}},
		{[][]byte{{0, 0}, {0xa5, 0x5a}, {0, 0}}, 1, []byte{0x5a}},
		{[][]byte{{0xa5, 0x11}, {0xa5, 0x22}}, 1, []byte{0x11}},
	} {
		m := &Map{m: c.m, keySize: 1, valSize: 1}

		var val [1]byte
		v := m.Lookup([]byte{0xa5}, val[:])

		if v != c.v || !reflect.DeepEqual(val[:], c.val) {
			t.Error("Lookup failed")
			t.Logf("expected: 0x%02x, %d", c.val, c.v)
			t.Logf("got:      0x%02x, %d", val[:], v)
			t.Fatal()
		}
	}
}
