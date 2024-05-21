// SPDX-License-Identifier: MIT
//
// Copyright 2023 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package sets

import (
	"cmp"
	"slices"
	"sort"
)

// Sorted is a set whose elements are sorted.
// Elems and Range will return the elements in sorted order.
type Sorted[E any] interface {
	Set[E]

	search(E) (int, bool)
}

// A CmpFunc is a comparison function.
// It returns 1 if a is greater than b.
// It returns -1 if a is less than b.
// Otherwise, it returns 0.
type CmpFunc[E any] func(a, b E) int

// An EqFunc is an equality function.
// It returns true if and only if a and b are identical.
type EqFunc[E any] func(a, b E) bool

// NewSorted returns a sorted set initialized with the given elements.
func NewSorted[E cmp.Ordered](elems ...E) Sorted[E] {
	return &ordered[E]{
		elems: stableSortUniqCmpEq(slices.Clone(elems), cmp.Compare[E], equal[E]),
	}
}

// NewSortedCmpFunc returns a sorted set initialized with the given elements.
// The comparison function is used to order and identify elements.
func NewSortedCmpFunc[E any](cmp CmpFunc[E], elems ...E) Sorted[E] {
	return NewSortedCmpEqFunc(cmp, func(a, b E) bool { return cmp(a, b) == 0 }, elems...)
}

// NewSortedCmpEqFunc returns a sorted set initialized with the given elements.
// The comparison function is only used to order elements and the equality
// function is used to identify elements.
//
// It may contain unique elements for which cmp(a, b) == 0 and eq(a, b) == false.
func NewSortedCmpEqFunc[E any](cmp CmpFunc[E], eq EqFunc[E], elems ...E) Sorted[E] {
	return &sorted[E]{
		elems: stableSortUniqCmpEq(slices.Clone(elems), cmp, eq),
		cmp:   cmp,
		eq:    eq,
	}
}

type ordered[E cmp.Ordered] struct {
	elems []E
}

func (set *ordered[E]) Contains(elem E) bool {
	_, ok := set.search(elem)
	return ok
}

func (set *ordered[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if _, ok := set.search(e); !ok {
			return false
		}
	}
	return true
}

func (set *ordered[E]) ContainsSet(other Set[E]) bool {
	switch other := other.(type) {
	case table[E]:
		for e := range other {
			if _, ok := set.search(e); !ok {
				return false
			}
		}
		return true
	case *ordered[E]:
		a := set.elems
		b := other.elems
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				ai++
			case av == bv:
				ai++
				bi++
			default: // ab > bv:
				return false
			}
		}
		return bi == bn
	case *sorted[E]:
		return set.ContainsAll(other.elems...)
	default:
		ok := true
		other.Range(func(e E) bool {
			_, ok = set.search(e)
			return ok
		})
		return ok
	}
}

func (set *ordered[E]) Insert(elem E) {
	idx, found := set.search(elem)
	if found {
		set.elems[idx] = elem
		return
	}
	set.elems = append(set.elems, elem)      // grow slice
	copy(set.elems[idx+1:], set.elems[idx:]) // slide elements right
	set.elems[idx] = elem                    // overwrite target
}

func (set *ordered[E]) InsertAll(elems ...E) {
	set.insertAll(slices.Clone(elems))
}

func (set *ordered[E]) InsertSet(other Set[E]) {
	if set == other {
		return
	}
	set.insertAll(other.Elems()) // InsertAll without Clone.
}

func (set *ordered[E]) insertAll(elems []E) {
	elems = uniq(stableSort(elems))
	set.elems = mergeUniqSortedLists(set.elems, elems)
}

func (set *ordered[E]) Remove(elem E) {
	idx, found := set.search(elem)
	if !found {
		return
	}
	var zero E
	last := len(set.elems) - 1
	copy(set.elems[idx:], set.elems[idx+1:]) // slide elements left
	set.elems[last] = zero                   // zero out last element to prevent leaks
	set.elems = set.elems[0:last]            // shrink slice
}

func (set *ordered[E]) RemoveAll(elems ...E) {
	set.removeAll(slices.Clone(elems))
}

func (set *ordered[E]) RemoveSet(other Set[E]) {
	switch other := other.(type) {
	case *ordered[E]:
		set.elems = diffUniqSortedLists(set.elems, other.elems)
	default:
		set.removeAll(other.Elems()) // RemoveAll without clone.
	}
}

func (set *ordered[E]) removeAll(elems []E) {
	elems = stableSort(elems)
	set.elems = diffUniqSortedLists(set.elems, elems)
}

func (set *ordered[E]) Intersection(other Set[E]) Set[E] {
	s := &ordered[E]{}
	if other, ok := other.(*ordered[E]); ok {
		a, b := set.elems, other.elems
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				ai++
			case av > bv:
				bi++
			default: // av == bv:
				s.elems = append(s.elems, av)
				ai++
				bi++
			}
		}
		return s
	}
	for _, e := range set.elems {
		if other.Contains(e) {
			s.elems = append(s.elems, e)
		}
	}
	return s
}

func (set *ordered[E]) Union(other Set[E]) Set[E] {
	if other, ok := other.(*ordered[E]); ok {
		s := &ordered[E]{}
		a, b := set.elems, other.elems
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				s.elems = append(s.elems, av)
				ai++
			case av > bv:
				s.elems = append(s.elems, bv)
				bi++
			default: // av == bv:
				s.elems = append(s.elems, av)
				ai++
				bi++
			}
		}
		s.elems = append(s.elems, a[ai:]...)
		s.elems = append(s.elems, b[bi:]...)
		return s
	}
	elems := stableSort(other.Elems())
	elems = mergeUniqSortedLists(elems, set.elems)
	return &ordered[E]{elems: elems}
}

func (set *ordered[E]) Difference(other Set[E]) Set[E] {
	s := &ordered[E]{}
	if other, ok := other.(*ordered[E]); ok {
		a, b := set.elems, other.elems
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				s.elems = append(s.elems, av)
				ai++
			case av > bv:
				bi++
			default: // av == bv:
				ai++
				bi++
			}
		}
		s.elems = append(s.elems, a[ai:]...)
		return s
	}
	for _, e := range set.elems {
		if !other.Contains(e) {
			s.elems = append(s.elems, e)
		}
	}
	return s
}

func (set *ordered[E]) SymmetricDifference(other Set[E]) Set[E] {
	s := &ordered[E]{}
	if other, ok := other.(*ordered[E]); ok {
		a, b := set.elems, other.elems
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				s.elems = append(s.elems, av)
				ai++
			case av > bv:
				s.elems = append(s.elems, bv)
				bi++
			default: // av == bv:
				ai++
				bi++
			}
		}
		s.elems = append(s.elems, a[ai:]...)
		s.elems = append(s.elems, b[bi:]...)
		return s
	}
	for _, e := range set.elems {
		if !other.Contains(e) {
			s.elems = append(s.elems, e)
		}
	}
	var elems []E
	other.Range(func(e E) bool {
		if !set.Contains(e) {
			elems = append(elems, e)
		}
		return true
	})
	s.insertAll(elems)
	return s
}

func (set *ordered[E]) Len() int {
	return len(set.elems)
}

func (set *ordered[E]) Elems() []E {
	return slices.Clone(set.elems)
}

func (set *ordered[E]) Range(fn func(v E) bool) {
	for _, v := range set.elems {
		if !fn(v) {
			return
		}
	}
}

func (set *ordered[E]) Clone() Set[E] {
	return &ordered[E]{
		elems: slices.Clone(set.elems),
	}
}

func (set *ordered[E]) search(elem E) (idx int, found bool) {
	n := len(set.elems)
	idx = sort.Search(n, func(i int) bool { return elem <= set.elems[i] })
	return idx, (idx < n && elem == set.elems[idx])
}

type sorted[E any] struct {
	elems []E
	cmp   func(E, E) int
	eq    func(E, E) bool
}

func (set *sorted[E]) Contains(elem E) bool {
	_, ok := set.search(elem)
	return ok
}

func (set *sorted[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if _, ok := set.search(e); !ok {
			return false
		}
	}
	return true
}

func (set *sorted[E]) ContainsSet(other Set[E]) bool {
	switch other := other.(type) {
	case *sorted[E]:
		return set.ContainsAll(other.elems...)
	default:
		ok := true
		other.Range(func(e E) bool {
			_, ok = set.search(e)
			return ok
		})
		return ok
	}
}

func (set *sorted[E]) Insert(elem E) {
	idx, found := set.search(elem)
	if found {
		set.elems[idx] = elem
		return
	}
	set.elems = append(set.elems, elem)      // Grow slice.
	copy(set.elems[idx+1:], set.elems[idx:]) // Slide elements right.
	set.elems[idx] = elem                    // Overwrite target.
}

func (set *sorted[E]) InsertAll(elems ...E) {
	set.insertAll(slices.Clone(elems))
}

func (set *sorted[E]) InsertSet(other Set[E]) {
	if set == other {
		return
	}
	set.insertAll(other.Elems()) // InsertAll without Clone.
}

func (set *sorted[E]) insertAll(elems []E) {
	elems = stableSortUniqCmpEq(elems, set.cmp, set.eq)
	set.elems = mergeSortedLists(set.elems, elems, set.cmp, set.eq)
}

func (set *sorted[E]) Remove(elem E) {
	idx, found := set.search(elem)
	if !found {
		return
	}
	var zero E
	last := len(set.elems) - 1
	copy(set.elems[idx:], set.elems[idx+1:]) // Slide elements left.
	set.elems[last] = zero                   // Zero out last element to prevent leaks.
	set.elems = set.elems[0:last]            // Shrink slice.
}

func (set *sorted[E]) RemoveAll(elems ...E) {
	set.removeAll(slices.Clone(elems))
}

func (set *sorted[E]) RemoveSet(other Set[E]) {
	set.removeAll(other.Elems()) // RemoveAll without clone.
}

func (set *sorted[E]) removeAll(elems []E) {
	elems = stableSortCmp(elems, set.cmp)
	set.elems = diffSortedLists(set.elems, elems, set.cmp, set.eq)
}

func (set *sorted[E]) Intersection(other Set[E]) Set[E] {
	s := &sorted[E]{
		cmp: set.cmp,
		eq:  set.eq,
	}
	for _, v := range set.elems {
		if other.Contains(v) {
			s.elems = append(s.elems, v)
		}
	}
	return s
}

func (set *sorted[E]) Union(other Set[E]) Set[E] {
	v := set.Clone()
	v.InsertSet(other)
	return v
}

func (set *sorted[E]) Difference(other Set[E]) Set[E] {
	s := &sorted[E]{
		cmp: set.cmp,
		eq:  set.eq,
	}
	for _, e := range set.elems {
		if !other.Contains(e) {
			s.elems = append(s.elems, e)
		}
	}
	return s
}

func (set *sorted[E]) SymmetricDifference(other Set[E]) Set[E] {
	s := &sorted[E]{
		cmp: set.cmp,
		eq:  set.eq,
	}
	for _, e := range set.elems {
		if !other.Contains(e) {
			s.elems = append(s.elems, e)
		}
	}
	var elems []E
	other.Range(func(e E) bool {
		if !set.Contains(e) {
			elems = append(elems, e)
		}
		return true
	})
	s.insertAll(elems)
	return s
}

func (set *sorted[E]) Len() int {
	return len(set.elems)
}

func (set *sorted[E]) Elems() []E {
	return slices.Clone(set.elems)
}

func (set *sorted[E]) Range(fn func(v E) bool) {
	for _, v := range set.elems {
		if !fn(v) {
			return
		}
	}
}

func (set *sorted[E]) Clone() Set[E] {
	return &sorted[E]{
		elems: slices.Clone(set.elems),
		cmp:   set.cmp,
		eq:    set.eq,
	}
}

func (set *sorted[E]) search(elem E) (idx int, found bool) {
	n := len(set.elems)
	idx = sort.Search(n, func(i int) bool { return set.cmp(elem, set.elems[i]) <= 0 })
	// There are two options:
	// 	1. The sort key doesn't exist and idx is where it should be inserted.
	//	2. The sort key exists one or more times and idx is where it first appears.
	//
	// Iterate through elements as long as the sort key matches,
	// looking for a fully matching value.
	for ; idx < n && set.cmp(elem, set.elems[idx]) == 0; idx++ {
		if set.eq(set.elems[idx], elem) {
			return idx, true
		}
	}
	return idx, false
}

type insert[E any] struct {
	i int
	e E
}

// mergeUniqSortedLists merges B into A,
// both of which must be sorted and contain unique values.
func mergeUniqSortedLists[E cmp.Ordered](a, b []E) []E {
	var inserts []insert[E]
	ai, an := 0, len(a)
	bi, bn := 0, len(b)
	for ai < an && bi < bn {
		switch av, bv := a[ai], b[bi]; {
		case av < bv:
			ai++
		case av > bv:
			inserts = append(inserts, insert[E]{ai, bv})
			bi++
		default: // av == bv:
			a[ai] = bv // Overwrite existing value.
			ai++
			bi++
		}
	}
	return insertInto(a, b[bi:], inserts)
}

// mergeSortedLists merges B into A, both of which must be sorted.
func mergeSortedLists[E any](a, b []E, cmp CmpFunc[E], eq EqFunc[E]) []E {
	var inserts []insert[E]
	ai, an := 0, len(a)
	bi, bn := 0, len(b)
	for ai < an && bi < bn {
		switch c := cmp(a[ai], b[bi]); {
		case c < 0:
			ai++
		case c > 0:
			inserts = append(inserts, insert[E]{ai, b[bi]})
			bi++
		default: // case c == 0:
			ar := runEq(a[ai:], cmp)
			br := runEq(b[bi:], cmp)
			ai += len(ar) // Insert at the end of the run.
			for _, be := range br {
				if i := slices.IndexFunc(ar, func(ae E) bool { return eq(be, ae) }); i >= 0 {
					ar[i] = be // Overwrite existing values.
					continue
				}
				inserts = append(inserts, insert[E]{ai, be})
			}
			bi += len(br)
		}
	}
	return insertInto(a, b[bi:], inserts)
}

func insertInto[E any](a, tail []E, inserts []insert[E]) []E {
	// [ a, c, d, e, f, h, i, l, m ]
	//    + inserts[ B1, G5, J7, K7 ]
	//    + tail[ N, O, P ]
	//
	// [ a, c, d, e, f, h, i, l, m, _, _, _, _, _, _, _ ]
	//
	// [ a, c, d, e, f, h, i, l, m, _, _, _, _, N, O, P ]
	//                        x     i           k
	//
	// [ a, c, d, e, f, h, i, !, !, _, K, l, m, N, O, P ]
	//                       i,x       k
	//
	// [ a, c, d, e, f, h, i, !, !, J, K, l, m, N, O, P ]
	//                  x     i     k
	//
	// [ a, c, d, e, f, !, G, h, i, J, K, l, m, N, O, P ]
	//      x           i  k
	//
	// [ a, B, c, d, e, f, G, h, i, J, K, l, m, N, O, P ]

	extra := len(inserts) + len(tail)
	// Grow slice.
	i, k := len(a), len(a)+extra // b[i:k] is slice of invalid values
	a = slices.Grow(a, k)[0:k]
	// Add tail.
	k -= len(tail)
	copy(a[k:], tail)
	// Do inserts.
	for v := len(inserts) - 1; v >= 0; v-- {
		x := inserts[v].i
		n := i - x
		copy(a[k-n:k], a[x:x+n]) // Copy the valid elems starting at idx to the end of the invalid values.
		i = x                    // Set start of invalid values to idx, where the start of the copied values was.
		k -= n + 1               // Set end of invalid values to one left of where the valid values were copied.
		a[k] = inserts[v].e      // Insert new elem just before the start of the copied values.
	}
	return a
}

// diffUniqSortedLists diffs B from A (e.g. A - B),
// both of which must be sorted and contain unique values.
func diffUniqSortedLists[E cmp.Ordered](a, b []E) []E {
	var deletes []int
	ai, an := 0, len(a)
	bi, bn := 0, len(b)
	for ai < an && bi < bn {
		switch av, bv := a[ai], b[bi]; {
		case av < bv:
			ai++
		case av > bv:
			bi++
		default: // av == bv:
			deletes = append(deletes, ai)
			ai++
			bi++
		}
	}
	return deleteFrom(a, deletes)
}

// diffSortedLists diffs B from A (e.g. A - B), both of which must be sorted.
func diffSortedLists[E any](a, b []E, cmp CmpFunc[E], eq EqFunc[E]) []E {
	var deletes []int
	ai, an := 0, len(a)
	bi, bn := 0, len(b)
	for ai < an && bi < bn {
		switch c := cmp(a[ai], b[bi]); {
		case c < 0:
			ai++
		case c > 0:
			bi++
		default: // case c == 0:
			ar := runEq(a[ai:], cmp)
			br := runEq(b[bi:], cmp)
			for i, ae := range ar {
				if slices.ContainsFunc(br, func(be E) bool { return eq(ae, be) }) {
					deletes = append(deletes, ai+i)
				}
			}
			ai += len(ar)
			bi += len(br)
		}
	}
	return deleteFrom(a, deletes)
}

func deleteFrom[E any](a []E, deletes []int) []E {
	// [ a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p ]
	//   - [ b-1, g-6, j-9, k-10, n-13, o-14, p-15 ]
	//
	// [ a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p ]
	//      i  j           k
	//
	// [ a, c, d, e, f, _, _, h, i, j, k, l, m, n, o, p ]
	//                  i     j     k
	//
	// [ a, c, d, e, f, h, i, _, _, _, k, l, m, n, o, p ]
	//                        i       j,k
	//
	// [ a, c, d, e, f, h, i, _, _, _, _, l, m, n, o, p ]
	//                        i           j     k
	//
	// [ a, c, d, e, f, h, i, l, m, _, _, _, _, _, o, p ]
	//                              i             j,k
	//
	// [ a, c, d, e, f, h, i, l, m, _, _, _, _, _, _, p ]
	//                           i                   j,k
	//
	// [ a, c, d, e, f, h, i, l, m, _, _, _, _, _, _, _ ]
	//                           i                     j,k
	//
	// [ a, c, d, e, f, h, i, l, m ]
	//
	// i = dst; j = src; k = end

	n := len(a)
	d := len(deletes)
	if d == 0 {
		return a
	}
	dst := deletes[0] // Start of next overwrite.
	src := dst + 1    // Start of elems to keep.
	for x := 1; x < d; x++ {
		end := deletes[x]                // Set this end of elems to keep.
		dst += copy(a[dst:], a[src:end]) // Slide elems to keep left and increment next dst.
		src = end + 1                    // Set next start of elems to keep.
	}
	copy(a[dst:], a[src:]) // Slide elmes to keep left.
	zero(a[n-d:])
	return a[:n-d]
}

// runEq returns the run of equal elements starting at 0.
func runEq[E any](elems []E, cmp CmpFunc[E]) []E {
	if k := len(elems); k < 2 {
		return elems
	}
	n := 1
	v0 := elems[0]
	for _, v := range elems[1:] {
		if cmp(v0, v) != 0 {
			break
		}
		n++
	}
	return elems[:n]
}

// stableSortUniqCmpEq stable sorts the list and removes duplicates in place.
// It uses O(n*log(n)) compares and O(n*log(n)*log(n)) swaps for the sort.
// It uses O(n) compares and up to O(n^2) eqs for the uniq.
func stableSortUniqCmpEq[T any](list []T, cmp CmpFunc[T], eq EqFunc[T]) []T {
	return uniqCmpEq(stableSortCmp(list, cmp), cmp, eq)
}

// stableSort stable sorts the list using O(n*log(n)) compares and O(n*log(n)*log(n)) swaps.
func stableSort[E cmp.Ordered](list []E) []E {
	slices.SortStableFunc(list, cmp.Compare[E])
	return list
}

// stableSortCmp stable sorts the list using O(n*log(n)) compares and O(n*log(n)*log(n)) swaps.
func stableSortCmp[T any](list []T, cmp CmpFunc[T]) []T {
	slices.SortStableFunc(list, cmp)
	return list
}

// uniq removes item duplicates in place and preserves order using O(n) compares.
func uniq[E cmp.Ordered](list []E) []E {
	n := len(slices.Compact(list))
	zero(list[n:])
	return list[:n]
}

// uniqCmpEq removes item duplicates in place and preserves order using O(n) compares and up to O(n^2) eqs.
func uniqCmpEq[T any](list []T, cmp CmpFunc[T], eq EqFunc[T]) []T {
	n := len(list)
	if n == 0 {
		return nil
	}
	src := 0
	dst := 0
	prev := list[dst]
	for i := 1; i < n; i++ {
		next := list[i]
		if cmp(next, prev) == 0 {
			continue
		}
		dst += copy(list[dst:], uniqEqSlow(list[src:i], eq))
		src = i
		prev = next
	}
	dst += copy(list[dst:], uniqEqSlow(list[src:], eq))
	zero(list[dst:])
	return list[:dst]
}

// uniqEqSlow removes item duplicates in place and preserves order using up to O(n^2) eqs.
func uniqEqSlow[T any](list []T, eq EqFunc[T]) []T {
	n := len(list)
	if n < 2 {
		return list
	}
	src := 1
	dst := 1
	eqFn := func(a T) bool { return eq(list[src], a) }
	for ; src < n; src++ {
		if idx := slices.IndexFunc(list[:dst], eqFn); idx >= 0 {
			list[idx] = list[src] // Overwrite existing match.
			continue
		}
		list[dst] = list[src]
		dst++
	}
	zero(list[dst:])
	return list[:dst]
}

func zero[T any](s []T) {
	var empty T
	for i := range s {
		s[i] = empty
	}
}

func equal[T comparable](a, b T) bool { return a == b }
