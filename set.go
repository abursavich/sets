// SPDX-License-Identifier: MIT
//
// Copyright 2023 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

// Package sets provides generic implementations of ordered and unordered sets,
// or collections of unique elements.
package sets

import (
	"golang.org/x/exp/maps"
)

// A Set is a collection of unique elements.
type Set[E any] interface {
	// Contains returns a value indicating if the given element is in the set.
	Contains(elem E) bool
	// ContainsAll returns a value indicating if all the given elements are in the set.
	ContainsAll(elems ...E) bool
	// ContainsSet returns a value indicating if all the elements of other are in the set.
	// It's semantically equivalent to calling ContainsAll(other.Elems())
	// but may be more efficient.
	ContainsSet(other Set[E]) bool

	// Insert adds the given element to the set if it's not in the set.
	Insert(elem E)
	// InsertAll adds the given elements to the set which are not in the set.
	// It's semantically equivalent to calling Insert with each of the elements,
	// but may be more efficient.
	InsertAll(elems ...E)
	// InsertSet adds the elements of other to the set which are not in the set.
	// It's semantically equivalent to calling InsertAll(other.Elems())
	// but may be more efficient.
	InsertSet(other Set[E])

	// Remove removes the given element from the set if it's in the set.
	Remove(elem E)
	// RemoveAll removes the given elements from the set which are in the set.
	// It's semantically equivalent to calling Remove with each of the elements,
	// but may be more efficient.
	RemoveAll(elems ...E)
	// RemoveSet removes the elements of other from the set which are in the set.
	// It's semantically equivalent to calling RemoveAll(other.Elems())
	// but may be more efficient.
	RemoveSet(other Set[E])

	// Intersection (A ∩ B) returns a new set that is the intersection of the set and other.
	Intersection(other Set[E]) Set[E]
	// Union (A ∪ B) returns a new set that is the union of the set and other.
	// It's semantically equivalent to cloning the set then calling InsertSet(other)
	// but may be more efficient.
	Union(other Set[E]) Set[E]
	// Difference (A − B) returns a new set that is the difference of the set and other.
	// It's semantically equivalent to cloning the set then calling RemoveSet(other)
	// but may be more efficient.
	Difference(other Set[E]) Set[E]
	// SymmetricDifference (A △ B) returns a new set that is the symmetric difference,
	// also known as disjunctive union, of the set and other.
	SymmetricDifference(other Set[E]) Set[E]

	// Len returns the size, also know as cardinality, of the set.
	Len() int
	// Elems returns a list of the elements in the set.
	Elems() []E
	// Range calls the given function with each element of the set until
	// there are no elements remaining or the function returns false.
	Range(fn func(elem E) bool)

	// Clone returns a copy of the set.
	Clone() Set[E]
}

// New returns a set initialized with the given elements.
func New[E comparable](elems ...E) Set[E] {
	set := make(table[E], len(elems))
	for _, elem := range elems {
		set[elem] = struct{}{}
	}
	return set
}

type table[E comparable] map[E]struct{}

func (set table[E]) Contains(elem E) bool {
	_, ok := set[elem]
	return ok
}

func (set table[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if _, ok := set[e]; !ok {
			return false
		}
	}
	return true
}

func (set table[E]) ContainsSet(other Set[E]) bool {
	switch other := other.(type) {
	case table[E]:
		for e := range other {
			if _, ok := set[e]; !ok {
				return false
			}
		}
		return true
	case *sorted[E]:
		for _, e := range other.elems {
			if _, ok := set[e]; !ok {
				return false
			}
		}
		return true
	default:
		ok := true
		other.Range(func(e E) bool {
			_, ok = set[e]
			return ok
		})
		return ok
	}
}

func (set table[E]) Insert(elem E) {
	set[elem] = struct{}{}
}

func (set table[E]) InsertAll(elems ...E) {
	for _, elem := range elems {
		set[elem] = struct{}{}
	}
}

func (set table[E]) InsertSet(other Set[E]) {
	switch other := other.(type) {
	case table[E]:
		for e := range other {
			set[e] = struct{}{}
		}
	case *sorted[E]:
		for _, e := range other.elems {
			set[e] = struct{}{}
		}
	default:
		other.Range(func(e E) bool {
			set[e] = struct{}{}
			return true
		})
	}
}

func (set table[E]) Remove(elem E) {
	delete(set, elem)
}

func (set table[E]) RemoveAll(elems ...E) {
	for _, e := range elems {
		delete(set, e)
	}
}

func (set table[E]) RemoveSet(other Set[E]) {
	switch other := other.(type) {
	case table[E]:
		for e := range other {
			delete(set, e)
		}
	case *sorted[E]:
		for _, e := range other.elems {
			delete(set, e)
		}
	default:
		other.Range(func(e E) bool {
			delete(set, e)
			return true
		})
	}
}

func (set table[E]) Intersection(other Set[E]) Set[E] {
	s := make(table[E])
	switch other := other.(type) {
	case table[E]:
		for e := range other {
			if _, ok := set[e]; ok {
				s[e] = struct{}{}
			}
		}
	case *sorted[E]:
		for _, e := range other.elems {
			if _, ok := set[e]; ok {
				s[e] = struct{}{}
			}
		}
	default:
		other.Range(func(e E) bool {
			if _, ok := set[e]; ok {
				s[e] = struct{}{}
			}
			return true
		})
	}
	return s
}

func (set table[E]) Union(other Set[E]) Set[E] {
	s := set.Clone()
	s.InsertSet(other)
	return s
}

func (set table[E]) Difference(other Set[E]) Set[E] {
	s := make(table[E])
	for e := range set {
		if !other.Contains(e) {
			s[e] = struct{}{}
		}
	}
	return s
}

func (set table[E]) SymmetricDifference(other Set[E]) Set[E] {
	s := make(table[E])
	switch other := other.(type) {
	case table[E]:
		for e := range set {
			if _, ok := other[e]; !ok {
				s[e] = struct{}{}
			}
		}
		for e := range other {
			if _, ok := set[e]; !ok {
				s[e] = struct{}{}
			}
		}
	case *sorted[E]:
		for e := range set {
			if !other.Contains(e) {
				s[e] = struct{}{}
			}
		}
		for _, e := range other.elems {
			if _, ok := set[e]; !ok {
				s[e] = struct{}{}
			}
		}
	default:
		for e := range set {
			if !other.Contains(e) {
				s[e] = struct{}{}
			}
		}
		other.Range(func(e E) bool {
			if _, ok := set[e]; !ok {
				s[e] = struct{}{}
			}
			return true
		})
	}
	return s
}

func (set table[E]) Len() int {
	return len(set)
}

func (set table[E]) Elems() []E {
	return maps.Keys(set)
}

func (set table[E]) Range(fn func(v E) bool) {
	for v := range set {
		if !fn(v) {
			return
		}
	}
}

func (set table[E]) Clone() Set[E] {
	return maps.Clone(set)
}
