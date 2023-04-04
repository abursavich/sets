// SPDX-License-Identifier: MIT
//
// Copyright 2023 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package sets

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

func cmpRunePtrVal(a, b *rune) int {
	return compare(*a, *b)
}

func toRunePtrs(s string) []*rune {
	runes := []rune(s)
	ptrs := make([]*rune, len(s))
	for i := range s {
		ptrs[i] = &runes[i]
	}
	return ptrs
}

type externalSet[E comparable] struct {
	Set[E]
}

func TestOrderedSets(t *testing.T) {
	newSetTester(t, []rune("abcdefghijklmnop"), []*setType[rune]{
		{
			name:    "table",
			newSet:  New[rune],
			cmpFn:   compare[rune],
			eqFn:    equal[rune],
			sorted:  false,
			uniqCmp: true,
		},
		{
			name:    "ordered",
			newSet:  func(elems ...rune) Set[rune] { return NewSorted(elems...) },
			cmpFn:   compare[rune],
			eqFn:    equal[rune],
			sorted:  true,
			uniqCmp: true,
		},
		{
			name:    "sorted",
			newSet:  func(elems ...rune) Set[rune] { return NewSortedCmpFunc(compare[rune], elems...) },
			cmpFn:   compare[rune],
			eqFn:    equal[rune],
			sorted:  true,
			uniqCmp: true,
		},
		{
			name:   "external",
			newSet: func(elems ...rune) Set[rune] { return &externalSet[rune]{New(elems...)} },
			skip:   true,
		},
	}).test(t)
}

func TestUnorderedSets(t *testing.T) {
	newSetTester(t, toRunePtrs("aaabbbcccdddeee"), []*setType[*rune]{
		{
			name:    "table",
			newSet:  New[*rune],
			cmpFn:   cmpRunePtrVal,
			eqFn:    equal[*rune],
			sorted:  false,
			uniqCmp: true,
		},
		{
			name:    "sorted",
			newSet:  func(elems ...*rune) Set[*rune] { return NewSortedCmpEqFunc(cmpRunePtrVal, equal[*rune], elems...) },
			cmpFn:   cmpRunePtrVal,
			eqFn:    equal[*rune],
			sorted:  true,
			uniqCmp: false,
		},
		{
			name:   "external",
			newSet: func(elems ...*rune) Set[*rune] { return &externalSet[*rune]{New(elems...)} },
			skip:   true,
		},
	}).test(t)
}

type setType[E any] struct {
	name    string
	newSet  func(elems ...E) Set[E]
	cmpFn   CmpFunc[E]
	eqFn    EqFunc[E]
	sorted  bool
	uniqCmp bool
	skip    bool
}

func (typ *setType[E]) sort(elems []E) []E {
	slices.SortStableFunc(elems, func(a, b E) bool { return typ.cmpFn(a, b) < 0 })
	return elems
}

type setTester[E any] struct {
	setTypes []*setType[E]
	elems    []E
	full     int
	half     int
	quarter  int
}

func newSetTester[E any](t *testing.T, elems []E, setTypes []*setType[E]) *setTester[E] {
	seed := time.Now().UnixNano()
	t.Logf("random seed: %v", seed)

	elems = slices.Clone(elems)
	rand.New(rand.NewSource(seed)).Shuffle(len(elems), func(i, k int) {
		elems[i], elems[k] = elems[k], elems[i]
	})

	return &setTester[E]{
		setTypes: setTypes,
		elems:    elems,
		full:     len(elems),
		half:     len(elems) / 2,
		quarter:  len(elems) / 4,
	}
}

func (st *setTester[E]) test(t *testing.T) {
	for _, typ := range st.setTypes {
		if typ.skip {
			continue
		}
		typ := typ
		t.Run(typ.name, func(t *testing.T) {
			t.Parallel()
			t.Run("Contains", func(t *testing.T) { st.testContains(t, typ) })
			t.Run("ContainsAll", func(t *testing.T) { st.testContainsAll(t, typ) })
			t.Run("ContainsSet", func(t *testing.T) { st.testContainsSet(t, typ) })
			t.Run("Insert", func(t *testing.T) { st.testInsert(t, typ) })
			t.Run("InsertAll", func(t *testing.T) { st.testInsertAll(t, typ) })
			t.Run("InsertSet", func(t *testing.T) { st.testInsertSet(t, typ) })
			t.Run("Remove", func(t *testing.T) { st.testRemove(t, typ) })
			t.Run("RemoveAll", func(t *testing.T) { st.testRemoveAll(t, typ) })
			t.Run("RemoveSet", func(t *testing.T) { st.testRemoveSet(t, typ) })
			t.Run("Intersection", func(t *testing.T) { st.testIntersection(t, typ) })
			t.Run("Union", func(t *testing.T) { st.testUnion(t, typ) })
			t.Run("Difference", func(t *testing.T) { st.testDifference(t, typ) })
			t.Run("SymmetricDifference", func(t *testing.T) { st.testSymmetricDifference(t, typ) })
			t.Run("Range", func(t *testing.T) { st.testRange(t, typ) })
			t.Run("Elems", func(t *testing.T) { st.testElems(t, typ) })
			t.Run("Clone", func(t *testing.T) { st.testClone(t, typ) })
		})
	}
}

func (st *setTester[E]) testContains(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems[:st.half]...)
	if got, want := set.Len(), st.half; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	for i, e := range st.elems[st.half:] {
		if set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: true; want: false", i)
		}
	}
	for i, e := range st.elems[:st.half] {
		if !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
	}
}

func (st *setTester[E]) testContainsAll(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems[:st.half]...)
	if got, want := set.Len(), st.half; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	if !set.ContainsAll() {
		t.Fatalf("set.ContainsAll(); got: false; want: true")
	}
	if i, k := 0, st.quarter; !set.ContainsAll(st.elems[i:k]...) {
		t.Fatalf("set.ContainsAll(elems[%v:%v]...); got: false; want: true", i, k)
	}
	if i, k := st.quarter, st.half+st.quarter; set.ContainsAll(st.elems[i:k]...) {
		t.Fatalf("set.ContainsAll(elems[%v:%v]...); got: true; want: false", i, k)
	}
	if i, k := st.half, st.full; set.ContainsAll(st.elems[i:k]...) {
		t.Fatalf("set.ContainsAll(elems[%v:%v]...); got: true; want: false", i, k)
	}
}

func (st *setTester[E]) testContainsSet(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems[:st.half]...)
	if got, want := set.Len(), st.half; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			if !set.ContainsSet(otherTyp.newSet()) {
				t.Fatalf("set.ContainsSet(); got: false; want: true")
			}
			if i, k := 0, st.quarter; !set.ContainsSet(otherTyp.newSet(st.elems[i:k]...)) {
				t.Fatalf("set.ContainsSet(newSet(elems[%v:%v]...)); got: false; want: true", i, k)
			}
			if i, k := st.quarter, st.half+st.quarter; set.ContainsSet(otherTyp.newSet(st.elems[i:k]...)) {
				t.Fatalf("set.ContainsSet(newSet([%v:%v]...)); got: true; want: false", i, k)
			}
			if i, k := st.half, st.full; set.ContainsSet(otherTyp.newSet(st.elems[i:k]...)) {
				t.Fatalf("set.ContainsSet(newSet(elems[%v:%v]...)); got: true; want: false", i, k)
			}
		})
	}
}

func (st *setTester[E]) testInsert(t *testing.T, typ *setType[E]) {
	set := typ.newSet()
	// Insert each element one at a time.
	for i, e := range st.elems {
		// Insert the element when it doesn't exist.
		if set.Insert(e); !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
		// Insert the element when it does exist.
		if set.Insert(e); !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
		// Check the length.
		if got, want := set.Len(), i+1; got != want {
			t.Fatalf("set.Len(); got: %v; want: %v", got, want)
		}
	}
}

func (st *setTester[E]) testInsertAll(t *testing.T, typ *setType[E]) {
	set := typ.newSet()
	// Insert the first half of the elements.
	n := st.half
	set.InsertAll(st.elems[:n]...)
	if got, want := set.Len(), n; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	for i, e := range st.elems[:n] {
		if !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
	}
	// Insert the first quarter of the elements again.
	set.InsertAll(st.elems[:st.quarter]...)
	if got, want := set.Len(), n; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	for i, e := range st.elems[:n] {
		if !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
	}
	// Insert elements that are partially overlapping with the previous elements.
	n = st.half + st.quarter
	set.InsertAll(st.elems[st.quarter:n]...)
	if got, want := set.Len(), n; got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	for i, e := range st.elems[:n] {
		if !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", i)
		}
	}
}

func (st *setTester[E]) testInsertSet(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			set := typ.newSet()
			// Insert a set with the first half of the elements.
			n := st.half
			set.InsertSet(otherTyp.newSet(st.elems[:n]...))
			if got, want := set.Len(), n; got != want {
				t.Fatalf("set.Len(); got: %v; want: %v", got, want)
			}
			for i, e := range st.elems[:n] {
				if !set.Contains(e) {
					t.Fatalf("set.Contains(%v); got: false; want: true", i)
				}
			}
			// Insert set into itself.
			set.InsertSet(set)
			if got, want := set.Len(), n; got != want {
				t.Fatalf("set.Len(); got: %v; want: %v", got, want)
			}
			// Insert a set with elements that are partially overlapping with the previous elements.
			n = st.half + st.quarter
			set.InsertSet(otherTyp.newSet(st.elems[st.quarter:n]...))
			if got, want := set.Len(), n; got != want {
				t.Fatalf("set.Len(); got: %v; want: %v", got, want)
			}
			for i, e := range st.elems[:n] {
				if !set.Contains(e) {
					t.Fatalf("set.Contains(%v); got: false; want: true", i)
				}
			}
		})
	}
}

func (st *setTester[E]) testRemove(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems...)
	// Remove each element one at a time.
	for n := len(st.elems) - 1; n >= 0; n-- {
		e := st.elems[n]
		if !set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: false; want: true", n)
		}
		// Remove when it exists.
		if set.Remove(e); set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: true; want: false", n)
		}
		// Remove when it doesn't exist.
		if set.Remove(e); set.Contains(e) {
			t.Fatalf("set.Contains(%v); got: true; want: false", n)
		}
		// Check the length.
		if got, want := set.Len(), n; got != want {
			t.Fatalf("set.Len(); got: %v; want: %v", got, want)
		}
	}
}

func (st *setTester[E]) testRemoveAll(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems...)

	// Remove the second half of elements.
	set.RemoveAll(st.elems[st.half:]...)
	st.check(t, typ, set, st.elems[:st.half])

	// Remove the last quarter of elements again.
	set.RemoveAll(st.elems[st.half+st.quarter:]...)
	st.check(t, typ, set, st.elems[:st.half])

	// Remove elements that are partially overlapping with the previous elements.
	set.RemoveAll(st.elems[st.quarter : st.quarter+st.half]...)
	st.check(t, typ, set, st.elems[:st.quarter])
}

func (st *setTester[E]) testRemoveSet(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			// Create will all elements.
			set := typ.newSet(st.elems...)

			// Remove the second half of the elements.
			set.RemoveSet(otherTyp.newSet(st.elems[st.half:]...))
			st.check(t, typ, set, st.elems[:st.half])

			// Remove elements that are partially overlapping with the previous elements.
			set.RemoveSet(otherTyp.newSet(st.elems[st.quarter : st.quarter+st.half]...))
			st.check(t, typ, set, st.elems[:st.quarter])
		})
	}
}

func (st *setTester[E]) testElems(t *testing.T, typ *setType[E]) {
	got := typ.newSet(st.elems...).Elems()
	if !typ.sorted {
		got = typ.sort(got)
	}
	want := typ.sort(slices.Clone(st.elems))
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatal("Unexpected diff in set.Elems():\n", diff)
	}
}

func (st *setTester[E]) testRange(t *testing.T, typ *setType[E]) {
	// Make sure we iterate over all values (in order if sorted set).
	set := typ.newSet(st.elems...)
	seen := typ.newSet()
	var prev E
	set.Range(func(e E) bool {
		if seen.Contains(e) {
			t.Fatal("Range(...) called with already seen value")
		}
		if typ.sorted && seen.Len() > 0 {
			if cmp := typ.cmpFn(prev, e); cmp > 0 || (typ.uniqCmp && cmp == 0) {
				t.Fatalf("Range(...) called out of order; prev: %#v; next: %#v", prev, e)
			}
		}
		seen.Insert(e)
		prev = e
		return true
	})
	if got, want := seen.Len(), st.full; got != want {
		t.Fatalf("Range not called on all elements; got: %v; want: %v", got, want)
	}
	// Make sure stop works.
	i := 0
	set.Range(func(e E) bool {
		i++
		return i < st.half
	})
	if got, want := i, st.half; got != want {
		t.Fatalf("Range not stopped after half elements; got: %v; want: %v", got, want)
	}
}

func (st *setTester[E]) testClone(t *testing.T, typ *setType[E]) {
	set := typ.newSet(st.elems...)
	clone := set.Clone()
	st.check(t, typ, clone, st.elems)

	if typ.sorted {
		got, want := clone.Elems(), set.Elems()
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatal("Unexpected sorting diff in clone.Elems():\n", diff)
		}
	}
}

func (st *setTester[E]) testIntersection(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			// Intersection from empty.
			st.check(t, typ, typ.newSet().Intersection(otherTyp.newSet(st.elems...)), nil)
			// Intersection with empty.
			st.check(t, typ, typ.newSet(st.elems...).Intersection(otherTyp.newSet()), nil)
			// Intersection with same elems.
			st.check(t, typ, typ.newSet(st.elems...).Union(otherTyp.newSet(st.elems...)), st.elems)
			// Intersection with overlapping elems.
			st.check(t, typ,
				typ.newSet(st.elems[:st.half]...).Intersection(otherTyp.newSet(st.elems[st.quarter:]...)),
				st.elems[st.quarter:st.half],
			)
		})
	}
}

func (st *setTester[E]) testUnion(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			// Union from empty.
			st.check(t, typ, typ.newSet().Union(otherTyp.newSet(st.elems...)), st.elems)
			// Union with empty.
			st.check(t, typ, typ.newSet(st.elems...).Union(otherTyp.newSet()), st.elems)
			// Union with same elems.
			st.check(t, typ, typ.newSet(st.elems...).Union(otherTyp.newSet(st.elems...)), st.elems)
			// Union with overlapping elems.
			st.check(t, typ, typ.newSet(st.elems[:st.half]...).Union(otherTyp.newSet(st.elems[st.quarter:]...)), st.elems)
		})
	}
}

func (st *setTester[E]) testDifference(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			// Difference from empty.
			st.check(t, typ, typ.newSet().Difference(otherTyp.newSet(st.elems...)), nil)
			// Difference with empty.
			st.check(t, typ, typ.newSet(st.elems...).Difference(otherTyp.newSet()), st.elems)
			// Difference with same elems.
			st.check(t, typ, typ.newSet(st.elems...).Difference(otherTyp.newSet(st.elems...)), nil)
			// Difference with overlapping elems.
			st.check(t, typ,
				typ.newSet(st.elems[:st.half]...).Difference(otherTyp.newSet(st.elems[st.quarter:]...)),
				st.elems[:st.quarter],
			)
		})
	}
}

func (st *setTester[E]) testSymmetricDifference(t *testing.T, typ *setType[E]) {
	for _, otherTyp := range st.setTypes {
		t.Run(otherTyp.name, func(t *testing.T) {
			// SymmetricDifference from empty.
			st.check(t, typ, typ.newSet().SymmetricDifference(otherTyp.newSet(st.elems...)), st.elems)
			// SymmetricDifference with empty.
			st.check(t, typ, typ.newSet(st.elems...).SymmetricDifference(otherTyp.newSet()), st.elems)
			// SymmetricDifference with same elems.
			st.check(t, typ, typ.newSet(st.elems...).SymmetricDifference(otherTyp.newSet(st.elems...)), nil)
			// SymmetricDifference with overlapping elems.
			st.check(t, typ,
				typ.newSet(st.elems[:st.half]...).SymmetricDifference(otherTyp.newSet(st.elems[st.quarter:]...)),
				append(slices.Clone(st.elems[:st.quarter]), st.elems[st.half:]...),
			)
		})
	}
}

func (st *setTester[E]) check(t *testing.T, typ *setType[E], set Set[E], elems []E) {
	t.Helper()

	if got, want := reflect.TypeOf(set), reflect.TypeOf(typ.newSet()); got != want {
		t.Fatalf("Unexpected set type; got: %v; want: %v", got, want)
	}
	if got, want := set.Len(), len(elems); got != want {
		t.Fatalf("set.Len(); got: %v; want: %v", got, want)
	}
	if !set.ContainsAll(elems...) {
		t.Fatalf("set.ContainsAll(...); got: false; want: true")
	}
}
