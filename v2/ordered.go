package sets

import (
	"cmp"
	"slices"

	"bursavich.dev/sets/internal/slicesx"
)

func NewSortedImmutable[E cmp.Ordered](elems ...E) SortedImmutable[E] {
	return &constSorted[E]{order(elems)}
}

func NewSortedMutable[E cmp.Ordered](elems ...E) SortedMutable[E] {
	return &varSorted[E]{order(elems)}
}

type ordered[E cmp.Ordered] struct {
	list []E
}

func order[E cmp.Ordered](elems []E) ordered[E] {
	list := slices.Clone(elems)
	slices.SortStableFunc(list, cmp.Compare)
	return ordered[E]{slices.Compact(list)}
}

func (s *ordered[E]) view() View[E]     { return s }
func (s *ordered[E]) data() []E         { return s.list }
func (s *ordered[E]) clone() ordered[E] { return ordered[E]{slices.Clone(s.list)} }

func (s *ordered[E]) Contains(e E) bool {
	_, ok := slices.BinarySearch(s.list, e)
	return ok
}
func (s *ordered[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if !s.Contains(e) {
			return false
		}
	}
	return true
}
func (s *ordered[E]) ContainsSet(other View[E]) bool {
	if other.Len() > len(s.list) {
		return false
	}
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			if !s.Contains(e) {
				return false
			}
		}
		return true
	case *ordered[E]:
		a, b := s.list, other.list
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
	case *funcSorted[E]:
		return s.ContainsAll(other.list...)
	default:
		ok := true
		other.Range(func(e E) bool {
			ok = s.Contains(e)
			return ok
		})
		return ok
	}
}

func (s *ordered[E]) Len() int   { return len(s.list) }
func (s *ordered[E]) Elems() []E { return ([]E)(slices.Clone(s.list)) }
func (s *ordered[E]) Range(fn func(e E) bool) {
	for _, e := range s.list {
		if !fn(e) {
			return
		}
	}
}

func (s *ordered[E]) intersection(other View[E]) ordered[E] {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	var list []E
	switch other := other.(type) {
	case *ordered[E]:
		a, b := s.list, other.list
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				ai++
			case av > bv:
				bi++
			default: // av == bv:
				list = append(list, av)
				ai++
				bi++
			}
		}
	default:
		for _, e := range s.list {
			if other.Contains(e) {
				list = append(list, e)
			}
		}
	}
	return ordered[E]{list}
}

func (s *ordered[E]) union(other View[E]) ordered[E] {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	var list []E
	switch other := other.(type) {
	case *ordered[E]:
		a, b := s.list, other.list
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				list = append(list, av)
				ai++
			case av > bv:
				list = append(list, bv)
				bi++
			default: // av == bv:
				list = append(list, av)
				ai++
				bi++
			}
		}
		list = append(list, a[ai:]...)
		list = append(list, b[bi:]...)
	default:
		list = other.Elems()
		slices.SortStableFunc(list, cmp.Compare)
		list = slicesx.MergeSortedUniq(slices.Compact(list), s.list)
	}
	return ordered[E]{list}
}

func (s *ordered[E]) difference(other View[E]) ordered[E] {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	var list []E
	switch other := other.(type) {
	case *ordered[E]:
		a, b := s.list, other.list
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				list = append(list, av)
				ai++
			case av > bv:
				bi++
			default: // av == bv:
				ai++
				bi++
			}
		}
		list = append(list, a[ai:]...)
	default:
		for _, e := range s.list {
			if !other.Contains(e) {
				list = append(list, e)
			}
		}
	}
	return ordered[E]{list}
}

func (s *ordered[E]) symmetricDifference(other View[E]) ordered[E] {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	var list []E
	switch other := other.(type) {
	case *ordered[E]:
		a, b := s.list, other.list
		ai, an := 0, len(a)
		bi, bn := 0, len(b)
		for ai < an && bi < bn {
			switch av, bv := a[ai], b[bi]; {
			case av < bv:
				list = append(list, av)
				ai++
			case av > bv:
				list = append(list, bv)
				bi++
			default: // av == bv:
				ai++
				bi++
			}
		}
		list = append(list, a[ai:]...)
		list = append(list, b[bi:]...)
	default:
		for _, e := range s.list {
			if !other.Contains(e) {
				list = append(list, e)
			}
		}
		var b []E
		other.Range(func(e E) bool {
			if !s.Contains(e) {
				b = append(b, e)
			}
			return true
		})
		slices.SortStableFunc(b, cmp.Compare)
		list = slicesx.MergeSortedUniq(list, slices.Compact(b))
	}
	return ordered[E]{list}
}

func (s *ordered[E]) insert(e E) {
	list := s.list
	i, ok := slices.BinarySearch(list, e)
	if ok {
		list[i] = e
		return
	}
	list = append(list, e)     // Drow slice.
	copy(list[i+1:], list[i:]) // Slide elements right.
	list[i] = e                // Overwrite target.
	s.list = list
}

func (s *ordered[E]) insertAll(unsorted []E) {
	slices.SortStableFunc(unsorted, cmp.Compare)
	s.list = slicesx.MergeSortedUniq(s.list, slices.Compact(unsorted))
}

func (s *ordered[E]) insertSet(other View[E]) {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	switch other := other.(type) {
	case *ordered[E]:
		s.list = slicesx.MergeSortedUniq(s.list, other.list)
	case SortedView[E]:
		s.list = slicesx.MergeSortedUniq(s.list, slices.Compact(other.Elems()))
	default:
		s.insertAll(other.Elems())
	}
}

func (s *ordered[E]) remove(e E) {
	list := s.list
	i, ok := slices.BinarySearch(list, e)
	if !ok {
		return
	}
	k := len(list) - 1
	copy(list[i:], list[i+1:]) // Slide elements left.
	clear(list[k:])            // Clear out last element to prevent leaks.
	s.list = list[:k]          // Shrink slice.
}

func (s *ordered[E]) removeAll(unsorted []E) {
	slices.SortStableFunc(unsorted, cmp.Compare)
	s.list = slicesx.DeleteSortedUniq(s.list, slices.Compact(unsorted))
}

func (s *ordered[E]) removeSet(other View[E]) {
	if o, ok := other.(viewer[E]); ok {
		other = o.view()
	}
	var elems []E
	switch other := other.(type) {
	case *ordered[E]:
		elems = other.list
	case SortedView[E]:
		elems = slices.Compact(other.Elems())
	default:
		elems = other.Elems()
		slices.SortStableFunc(elems, cmp.Compare)
		elems = slices.Compact(elems)
	}
	s.list = slicesx.DeleteSortedUniq(s.list, elems)
}

type constSorted[E cmp.Ordered] struct{ ordered[E] }

func (s *constSorted[E]) Intersection(other View[E]) SortedImmutable[E] {
	return &constSorted[E]{s.intersection(other)}
}
func (s *constSorted[E]) Union(other View[E]) SortedImmutable[E] {
	return &constSorted[E]{s.union(other)}
}
func (s *constSorted[E]) Difference(other View[E]) SortedImmutable[E] {
	return &constSorted[E]{s.difference(other)}
}
func (s *constSorted[E]) SymmetricDifference(other View[E]) SortedImmutable[E] {
	return &constSorted[E]{s.symmetricDifference(other)}
}

func (s *constSorted[E]) MutableCopy() SortedMutable[E] { return &varSorted[E]{s.clone()} }
func (s *constSorted[E]) ImmutableCopy() Mutable[E]     { return &varOrdered[E]{s.clone()} }
func (s *constSorted[E]) Immutable() Immutable[E]       { return (*constOrdered[E])(s) }

type constOrdered[E cmp.Ordered] struct{ ordered[E] }

func (s *constOrdered[E]) Intersection(other View[E]) Immutable[E] {
	return &constOrdered[E]{s.intersection(other)}
}
func (s *constOrdered[E]) Union(other View[E]) Immutable[E] {
	return &constOrdered[E]{s.union(other)}
}
func (s *constOrdered[E]) Difference(other View[E]) Immutable[E] {
	return &constOrdered[E]{s.difference(other)}
}
func (s *constOrdered[E]) SymmetricDifference(other View[E]) Immutable[E] {
	return &constOrdered[E]{s.symmetricDifference(other)}
}

func (s *constOrdered[E]) MutableCopy() Mutable[E] { return &varOrdered[E]{s.clone()} }

type varSorted[E cmp.Ordered] struct {
	ordered[E]
}

func (s *varSorted[E]) Intersection(o View[E]) SortedMutable[E] {
	return &varSorted[E]{s.intersection(o)}
}
func (s *varSorted[E]) Union(o View[E]) SortedMutable[E] {
	return &varSorted[E]{s.union(o)}
}
func (s *varSorted[E]) Difference(o View[E]) SortedMutable[E] {
	return &varSorted[E]{s.difference(o)}
}
func (s *varSorted[E]) SymmetricDifference(o View[E]) SortedMutable[E] {
	return &varSorted[E]{s.symmetricDifference(o)}
}

func (s *varSorted[E]) Insert(e E)          { s.insert(e) }
func (s *varSorted[E]) InsertAll(es ...E)   { s.insertAll(slices.Clone(es)) }
func (s *varSorted[E]) InsertSet(o View[E]) { s.insertSet(o) }

func (s *varSorted[E]) Remove(e E)          { s.remove(e) }
func (s *varSorted[E]) RemoveAll(es ...E)   { s.removeAll(slices.Clone(es)) }
func (s *varSorted[E]) RemoveSet(o View[E]) { s.removeSet(o) }

func (s *varSorted[E]) ImmutableCopy() SortedImmutable[E] { return &constSorted[E]{s.clone()} }
func (s *varSorted[E]) Clone() SortedMutable[E]           { return &varSorted[E]{s.clone()} }
func (s *varSorted[E]) Mutable() Mutable[E]               { return (*varOrdered[E])(s) }

type varOrdered[E cmp.Ordered] struct {
	ordered[E]
}

func (s *varOrdered[E]) Intersection(other View[E]) Mutable[E] {
	return &varOrdered[E]{s.intersection(other)}
}
func (s *varOrdered[E]) Union(other View[E]) Mutable[E] {
	return &varOrdered[E]{s.union(other)}
}
func (s *varOrdered[E]) Difference(other View[E]) Mutable[E] {
	return &varOrdered[E]{s.difference(other)}
}
func (s *varOrdered[E]) SymmetricDifference(other View[E]) Mutable[E] {
	return &varOrdered[E]{s.symmetricDifference(other)}
}

func (s *varOrdered[E]) Insert(e E)              { s.insert(e) }
func (s *varOrdered[E]) InsertAll(elems ...E)    { s.insertAll(slices.Clone(elems)) }
func (s *varOrdered[E]) InsertSet(other View[E]) { s.insertSet(other) }

func (s *varOrdered[E]) Remove(e E)              { s.remove(e) }
func (s *varOrdered[E]) RemoveAll(elems ...E)    { s.removeAll(slices.Clone(elems)) }
func (s *varOrdered[E]) RemoveSet(other View[E]) { s.removeSet(other) }

func (s *varOrdered[E]) ImmutableCopy() Immutable[E] { return &constOrdered[E]{s.clone()} }
func (s *varOrdered[E]) Clone() Mutable[E]           { return &varOrdered[E]{s.clone()} }
