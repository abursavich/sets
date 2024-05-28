package sets

import (
	"golang.org/x/exp/maps"
)

type table[E comparable] struct {
	tbl map[E]struct{}
}

func newTable[E comparable](elems ...E) table[E] {
	tbl := make(map[E]struct{}, len(elems))
	for _, e := range elems {
		tbl[e] = struct{}{}
	}
	return table[E]{tbl}
}

func (s *table[E]) view() View[E]        { return s }
func (s *table[E]) data() map[E]struct{} { return s.tbl }
func (s *table[E]) clone() table[E]      { return table[E]{maps.Clone(s.tbl)} }

func (s *table[E]) Contains(elem E) bool {
	_, ok := s.tbl[elem]
	return ok
}

func (s *table[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if _, ok := s.tbl[e]; !ok {
			return false
		}
	}
	return true
}

func (s *table[E]) ContainsSet(other View[E]) bool {
	if other.Len() > len(s.tbl) {
		return false
	}
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			if _, ok := s.tbl[e]; !ok {
				return false
			}
		}
		return true
	case listView[E]:
		for _, e := range other.data() {
			if _, ok := s.tbl[e]; !ok {
				return false
			}
		}
		return true
	default:
		ok := true
		other.Range(func(e E) bool {
			_, ok = s.tbl[e]
			return ok
		})
		return ok
	}
}

func (s *table[E]) Len() int   { return len(s.tbl) }
func (s *table[E]) Elems() []E { return maps.Keys(s.tbl) }
func (s *table[E]) Range(fn func(v E) bool) {
	for v := range s.tbl {
		if !fn(v) {
			return
		}
	}
}

func (s *table[E]) intersection(other View[E]) table[E] {
	tbl := make(map[E]struct{})
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			if _, ok := s.tbl[e]; ok {
				tbl[e] = struct{}{}
			}
		}
	case listView[E]:
		for _, e := range other.data() {
			if _, ok := s.tbl[e]; ok {
				tbl[e] = struct{}{}
			}
		}
	default:
		other.Range(func(e E) bool {
			if _, ok := s.tbl[e]; ok {
				tbl[e] = struct{}{}
			}
			return true
		})
	}
	return table[E]{tbl}
}

func (s *table[E]) union(other View[E]) table[E] {
	tbl := maps.Clone(s.tbl)
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			tbl[e] = struct{}{}
		}
	case listView[E]:
		for _, e := range other.data() {
			tbl[e] = struct{}{}
		}
	default:
		other.Range(func(e E) bool {
			tbl[e] = struct{}{}
			return true
		})
	}
	return table[E]{tbl}
}

func (s *table[E]) difference(other View[E]) table[E] {
	tbl := make(map[E]struct{})
	for e := range s.tbl {
		if !other.Contains(e) {
			tbl[e] = struct{}{}
		}
	}
	return table[E]{tbl}
}

func (s *table[E]) symmetricDifference(other View[E]) table[E] {
	tbl := make(map[E]struct{})
	switch other := other.(type) {
	case tableView[E]:
		o := other.data()
		for e := range s.tbl {
			if _, ok := o[e]; !ok {
				tbl[e] = struct{}{}
			}
		}
		for e := range o {
			if _, ok := s.tbl[e]; !ok {
				tbl[e] = struct{}{}
			}
		}
	case listView[E]:
		for _, e := range other.data() {
			if _, ok := s.tbl[e]; !ok {
				tbl[e] = struct{}{}
			}
		}
		for e := range s.tbl {
			if !other.Contains(e) {
				tbl[e] = struct{}{}
			}
		}
	default:
		for e := range s.tbl {
			if !other.Contains(e) {
				tbl[e] = struct{}{}
			}
		}
		other.Range(func(e E) bool {
			if _, ok := s.tbl[e]; !ok {
				tbl[e] = struct{}{}
			}
			return true
		})
	}
	return table[E]{tbl}
}

// NewImmutable returns a new immutable s with the given elements.
func NewImmutable[E comparable](elems ...E) Immutable[E] {
	return &constTable[E]{newTable(elems...)}
}

type constTable[E comparable] struct{ table[E] }

func (s *constTable[E]) Intersection(other View[E]) Immutable[E] {
	return &constTable[E]{s.intersection(other)}
}
func (s *constTable[E]) Union(other View[E]) Immutable[E] {
	return &constTable[E]{s.union(other)}
}
func (s *constTable[E]) Difference(other View[E]) Immutable[E] {
	return &constTable[E]{s.difference(other)}
}
func (s *constTable[E]) SymmetricDifference(other View[E]) Immutable[E] {
	return &constTable[E]{s.symmetricDifference(other)}
}

func (s *constTable[E]) MutableCopy() Mutable[E] { return &varTable[E]{s.clone()} }

// NewMutable returns a new Mutable s with the given elements.
func NewMutable[E comparable](elems ...E) Mutable[E] {
	return &varTable[E]{newTable(elems...)}
}

type varTable[E comparable] struct{ table[E] }

func (s *varTable[E]) Intersection(other View[E]) Mutable[E] {
	return &varTable[E]{s.intersection(other)}
}
func (s *varTable[E]) Union(other View[E]) Mutable[E] {
	return &varTable[E]{s.union(other)}
}
func (s *varTable[E]) Difference(other View[E]) Mutable[E] {
	return &varTable[E]{s.difference(other)}
}
func (s *varTable[E]) SymmetricDifference(other View[E]) Mutable[E] {
	return &varTable[E]{s.symmetricDifference(other)}
}

func (s *varTable[E]) Insert(elem E) { s.tbl[elem] = struct{}{} }

func (s *varTable[E]) InsertAll(elems ...E) {
	for _, e := range elems {
		s.tbl[e] = struct{}{}
	}
}

func (s *varTable[E]) InsertSet(other View[E]) {
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			s.tbl[e] = struct{}{}
		}
	case listView[E]:
		for _, e := range other.data() {
			s.tbl[e] = struct{}{}
		}
	default:
		other.Range(func(e E) bool {
			s.tbl[e] = struct{}{}
			return true
		})
	}
}

func (s *varTable[E]) Remove(elem E) { delete(s.tbl, elem) }

func (s *varTable[E]) RemoveAll(elems ...E) {
	for _, e := range elems {
		delete(s.tbl, e)
	}
}

func (s *varTable[E]) RemoveSet(other View[E]) {
	switch other := other.(type) {
	case tableView[E]:
		for e := range other.data() {
			delete(s.tbl, e)
		}
	case listView[E]:
		for _, e := range other.data() {
			delete(s.tbl, e)
		}
	default:
		other.Range(func(e E) bool {
			delete(s.tbl, e)
			return true
		})
	}
}

func (s *varTable[E]) ImmutableCopy() Immutable[E] { return &constTable[E]{s.clone()} }
func (s *varTable[E]) Clone() Mutable[E]           { return &varTable[E]{s.clone()} }
