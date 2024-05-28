package sets

import (
	"slices"

	"bursavich.dev/sets/internal/slicesx"
)

func NewSortedImmutableFunc[E any](cmp func(E, E) int, elems ...E) SortedImmutable[E] {
	return NewSortedImmutableFuncs(cmp, func(a, b E) bool { return cmp(a, b) == 0 }, elems...)
}

func NewSortedImmutableFuncs[E any](cmp func(E, E) int, eq func(E, E) bool, elems ...E) SortedImmutable[E] {
	return &constFuncSorted[E]{funcSort(elems, cmp, eq)}
}

func NewSortedMutableFunc[E any](cmp func(E, E) int, elems ...E) SortedMutable[E] {
	return NewSortedMutableFuncs(cmp, func(a, b E) bool { return cmp(a, b) == 0 }, elems...)
}

func NewSortedMutableFuncs[E any](cmp func(E, E) int, eq func(E, E) bool, elems ...E) SortedMutable[E] {
	return &varFuncSorted[E]{funcSort(elems, cmp, eq)}
}

type funcSorted[E any] struct {
	list []E
	cmp  func(E, E) int
	eq   func(E, E) bool
}

func funcSort[E any](elems []E, cmp func(E, E) int, eq func(E, E) bool) funcSorted[E] {
	return funcSorted[E]{
		slicesx.StableSortUniqFuncs(slices.Clone(elems), cmp, eq),
		cmp,
		eq,
	}
}

func (s *funcSorted[E]) view() View[E] { return s }
func (s *funcSorted[E]) data() []E     { return s.list }
func (s *funcSorted[E]) clone() funcSorted[E] {
	return funcSorted[E]{slices.Clone(s.list), s.cmp, s.eq}
}

func (s *funcSorted[E]) search(e E) (int, bool) {
	i, ok := slices.BinarySearchFunc(s.list, e, s.cmp)
	if !ok {
		return i, false
	}
	// There are two options:
	// 	1. The sort key doesn't exist and idx is where it should be inserted.
	//	2. The sort key exists one or more times and idx is where it first appears.
	//
	// Iterate through elements as long as the sort key matches,
	// looking for a fully matching value.
	for j, v := range s.list[i:] {
		if s.cmp(e, v) != 0 {
			return i + j, false
		}
		if s.eq(e, v) {
			return i + j, true
		}
	}
	return len(s.list), false
}

func (s *funcSorted[E]) Contains(e E) bool {
	_, ok := s.search(e)
	return ok
}

func (s *funcSorted[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if !s.Contains(e) {
			return false
		}
	}
	return true
}

func (s *funcSorted[E]) ContainsSet(other View[E]) bool {
	if other.Len() > len(s.list) {
		return false
	}
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	switch other := other.(type) {
	case *funcSorted[E]:
		return s.ContainsAll(other.list...)
	default:
		ok := true
		other.Range(func(e E) bool {
			_, ok = s.search(e)
			return ok
		})
		return ok
	}
}

func (s *funcSorted[E]) Len() int   { return len(s.list) }
func (s *funcSorted[E]) Elems() []E { return ([]E)(slices.Clone(s.list)) }
func (s *funcSorted[E]) Range(fn func(e E) bool) {
	for _, e := range s.list {
		if !fn(e) {
			return
		}
	}
}

func (s *funcSorted[E]) intersection(other View[E]) funcSorted[E] {
	out := funcSorted[E]{nil, s.cmp, s.eq}
	for _, v := range s.list {
		if other.Contains(v) {
			out.list = append(out.list, v)
		}
	}
	return out
}

func (s *funcSorted[E]) union(other View[E]) funcSorted[E] {
	out := s.clone()
	out.insertSet(other)
	return out
}

func (s *funcSorted[E]) difference(other View[E]) funcSorted[E] {
	out := funcSorted[E]{nil, s.cmp, s.eq}
	for _, v := range s.list {
		if !other.Contains(v) {
			out.list = append(out.list, v)
		}
	}
	return out
}

func (s *funcSorted[E]) symmetricDifference(other View[E]) funcSorted[E] {
	out := funcSorted[E]{nil, s.cmp, s.eq}
	for _, v := range s.list {
		if !other.Contains(v) {
			out.list = append(out.list, v)
		}
	}
	var elems []E
	other.Range(func(e E) bool {
		if !s.Contains(e) {
			elems = append(elems, e)
		}
		return true
	})
	out.insertAll(elems)
	return out
}

func (s *funcSorted[E]) insert(e E) {
	i, ok := s.search(e)
	if ok {
		s.list[i] = e
		return
	}
	s.list = append(s.list, e)     // Grow slice.
	copy(s.list[i+1:], s.list[i:]) // Slide elements right.
	s.list[i] = e                  // Overwrite target.
}

func (s *funcSorted[E]) insertAll(unsorted []E) {
	s.list = slicesx.MergeSorted(
		s.list,
		slicesx.StableSortUniqFuncs(unsorted, s.cmp, s.eq),
		s.cmp,
		s.eq,
	)
}

func (s *funcSorted[E]) insertSet(other View[E]) {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	if s == other {
		return
	}
	s.insertAll(other.Elems())
}

func (s *funcSorted[E]) remove(e E) {
	i, ok := s.search(e)
	if !ok {
		return
	}
	k := len(s.list) - 1
	copy(s.list[i:], s.list[i+1:]) // Slide elements left.
	clear(s.list[k:])              // Clear out last element to prevent leaks.
	s.list = s.list[:k]            // Shrink slice.
}

func (s *funcSorted[E]) removeAll(unsorted []E) {
	s.list = slicesx.DeleteSorted(
		s.list,
		slicesx.StableSortFunc(unsorted, s.cmp),
		s.cmp,
		s.eq,
	)
}

func (s *funcSorted[E]) removeSet(other View[E]) {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	if s == other {
		clear(s.list)
		s.list = s.list[:0]
		return
	}
	s.removeAll(other.Elems())
}

type constFuncSorted[E any] struct{ funcSorted[E] }

func (s *constFuncSorted[E]) Intersection(other View[E]) SortedImmutable[E] {
	return &constFuncSorted[E]{s.intersection(other)}
}
func (s *constFuncSorted[E]) Union(other View[E]) SortedImmutable[E] {
	return &constFuncSorted[E]{s.union(other)}
}
func (s *constFuncSorted[E]) Difference(other View[E]) SortedImmutable[E] {
	return &constFuncSorted[E]{s.difference(other)}
}
func (s *constFuncSorted[E]) SymmetricDifference(other View[E]) SortedImmutable[E] {
	return &constFuncSorted[E]{s.symmetricDifference(other)}
}

func (s *constFuncSorted[E]) MutableCopy() SortedMutable[E] { return &varFuncSorted[E]{s.clone()} }
func (s *constFuncSorted[E]) Immutable() Immutable[E]       { return (*constFuncOrdered[E])(s) }

type constFuncOrdered[E any] struct{ funcSorted[E] }

func (s *constFuncOrdered[E]) Intersection(other View[E]) Immutable[E] {
	return &constFuncOrdered[E]{s.intersection(other)}
}
func (s *constFuncOrdered[E]) Union(other View[E]) Immutable[E] {
	return &constFuncOrdered[E]{s.union(other)}
}
func (s *constFuncOrdered[E]) Difference(other View[E]) Immutable[E] {
	return &constFuncOrdered[E]{s.difference(other)}
}
func (s *constFuncOrdered[E]) SymmetricDifference(other View[E]) Immutable[E] {
	return &constFuncOrdered[E]{s.symmetricDifference(other)}
}

func (s *constFuncOrdered[E]) MutableCopy() Mutable[E] { return &varFuncOrdered[E]{s.clone()} }

type varFuncSorted[E any] struct{ funcSorted[E] }

func (s *varFuncSorted[E]) Intersection(o View[E]) SortedMutable[E] {
	return &varFuncSorted[E]{s.intersection(o)}
}
func (s *varFuncSorted[E]) Union(o View[E]) SortedMutable[E] {
	return &varFuncSorted[E]{s.union(o)}
}
func (s *varFuncSorted[E]) Difference(o View[E]) SortedMutable[E] {
	return &varFuncSorted[E]{s.difference(o)}
}
func (s *varFuncSorted[E]) SymmetricDifference(o View[E]) SortedMutable[E] {
	return &varFuncSorted[E]{s.symmetricDifference(o)}
}

func (s *varFuncSorted[E]) Insert(e E)          { s.insert(e) }
func (s *varFuncSorted[E]) InsertAll(es ...E)   { s.insertAll(slices.Clone(es)) }
func (s *varFuncSorted[E]) InsertSet(o View[E]) { s.insertSet(o) }

func (s *varFuncSorted[E]) Remove(e E)          { s.remove(e) }
func (s *varFuncSorted[E]) RemoveAll(es ...E)   { s.removeAll(slices.Clone(es)) }
func (s *varFuncSorted[E]) RemoveSet(o View[E]) { s.removeSet(o) }

func (s *varFuncSorted[E]) ImmutableCopy() SortedImmutable[E] { return &constFuncSorted[E]{s.clone()} }
func (s *varFuncSorted[E]) Clone() SortedMutable[E]           { return &varFuncSorted[E]{s.clone()} }
func (s *varFuncSorted[E]) Mutable() Mutable[E]               { return (*varFuncOrdered[E])(s) }

type varFuncOrdered[E any] struct{ funcSorted[E] }

func (s *varFuncOrdered[E]) Intersection(other View[E]) Mutable[E] {
	return &varFuncOrdered[E]{s.intersection(other)}
}
func (s *varFuncOrdered[E]) Union(other View[E]) Mutable[E] {
	return &varFuncOrdered[E]{s.union(other)}
}
func (s *varFuncOrdered[E]) Difference(other View[E]) Mutable[E] {
	return &varFuncOrdered[E]{s.difference(other)}
}
func (s *varFuncOrdered[E]) SymmetricDifference(other View[E]) Mutable[E] {
	return &varFuncOrdered[E]{s.symmetricDifference(other)}
}

func (s *varFuncOrdered[E]) Insert(e E)              { s.insert(e) }
func (s *varFuncOrdered[E]) InsertAll(elems ...E)    { s.insertAll(slices.Clone(elems)) }
func (s *varFuncOrdered[E]) InsertSet(other View[E]) { s.insertSet(other) }

func (s *varFuncOrdered[E]) Remove(e E)              { s.remove(e) }
func (s *varFuncOrdered[E]) RemoveAll(elems ...E)    { s.removeAll(slices.Clone(elems)) }
func (s *varFuncOrdered[E]) RemoveSet(other View[E]) { s.removeSet(other) }

func (s *varFuncOrdered[E]) ImmutableCopy() Immutable[E] { return &constFuncOrdered[E]{s.clone()} }
func (s *varFuncOrdered[E]) Clone() Mutable[E]           { return &varFuncOrdered[E]{s.clone()} }
