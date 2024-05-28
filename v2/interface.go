package sets

type viewer[E any] interface {
	view() View[E]
}

type tableView[E comparable] interface {
	View[E]
	data() map[E]struct{}
}

type listView[E any] interface {
	View[E]
	data() []E
}

// A View is a collection of unique elements.
type View[E any] interface {
	// Contains returns a value indicating if the given element is in the set.
	Contains(elem E) bool
	// ContainsAll returns a value indicating if all the given elements are in the set.
	ContainsAll(elems ...E) bool
	// ContainsSet returns a value indicating if all the elements of other are in the set.
	// It's semantically equivalent to calling ContainsAll(other.Elems())
	// but may be more efficient.
	ContainsSet(other View[E]) bool

	// Len returns the size, also known as cardinality, of the set.
	Len() int
	// Elems returns a list of the elements in the set.
	Elems() []E
	// Range calls the given function with each element of the set until
	// there are no elements remaining or the function returns false.
	Range(fn func(elem E) bool)
}

// ImmutableOperations are operations for immutable sets.
type ImmutableOperations[E any, Self View[E]] interface {
	// Intersection (A ∩ B) returns a new set that is the intersection of the set and other.
	Intersection(other View[E]) Self
	// Union (A ∪ B) returns a new set that is the union of the set and other.
	// It's semantically equivalent to cloning the set then calling InsertSet(other)
	// but may be more efficient.
	Union(other View[E]) Self
	// Difference (A − B) returns a new set that is the difference of the set and other.
	// It's semantically equivalent to cloning the set then calling RemoveSet(other)
	// but may be more efficient.
	Difference(other View[E]) Self
	// SymmetricDifference (A △ B) returns a new set that is the symmetric difference,
	// also known as disjunctive union, of the set and other.
	SymmetricDifference(other View[E]) Self
}

// ImmutableSet defines the shared features of immutable sets.
type ImmutableSet[E any, Self View[E], Mutable View[E]] interface {
	View[E]
	ImmutableOperations[E, Self]

	// Mutable returns a mutable copy of the set.
	MutableCopy() Mutable
}

// MutableOperations are operations for mutable sets.
type MutableOperations[E any] interface {
	// Insert adds the given element to the set if it's not in the set.
	Insert(elem E)
	// InsertAll adds the given elements to the set which are not in the set.
	// It's semantically equivalent to calling Insert with each of the elements,
	// but may be more efficient.
	InsertAll(elems ...E)
	// InsertSet adds the elements of other to the set which are not in the set.
	// It's semantically equivalent to calling InsertAll(other.Elems())
	// but may be more efficient.
	InsertSet(other View[E])

	// Remove removes the given element from the set if it's in the set.
	Remove(elem E)
	// RemoveAll removes the given elements from the set which are in the set.
	// It's semantically equivalent to calling Remove with each of the elements,
	// but may be more efficient.
	RemoveAll(elems ...E)
	// RemoveSet removes the elements of other from the set which are in the set.
	// It's semantically equivalent to calling RemoveAll(other.Elems())
	// but may be more efficient.
	RemoveSet(other View[E])
}

// MutableSet defines the shared features of mutable sets.
type MutableSet[E any, Self View[E], Immutable View[E]] interface {
	View[E]
	ImmutableOperations[E, Self]
	MutableOperations[E]

	// ImmutableCopy returns an immutable copy of the set.
	ImmutableCopy() Immutable
	// Clone returns a copy of the mutable set.
	Clone() Self
}

// Immutable is an immutable set of unique elements.
type Immutable[E any] interface {
	ImmutableSet[E, Immutable[E], Mutable[E]]
}

// Mutable is a mutable set of unique elements.
type Mutable[E any] interface {
	MutableSet[E, Mutable[E], Immutable[E]]
}

// A SortedView is a view whose elements are sorted.
// Elems and Range will return the elements in sorted order.
type SortedView[E any] interface {
	View[E]

	listView[E]
}

// SortedImmutable is an immutable set of sorted unique elements.
type SortedImmutable[E any] interface {
	SortedView[E]
	ImmutableSet[E, SortedImmutable[E], SortedMutable[E]]

	// Immutable returns the underlying immutable set without the sorted interface.
	Immutable() Immutable[E]
}

// SortedMutable is a mutable set of sorted unique elements.
type SortedMutable[E any] interface {
	SortedView[E]
	MutableSet[E, SortedMutable[E], SortedImmutable[E]]

	// Mutable returns the underlying mutable set without the sorted interface.
	Mutable() Mutable[E]
}
