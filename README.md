# Sets
[![License][license-img]][license]
[![GoDev Reference][godev-img]][godev]
[![Go Report Card][goreportcard-img]][goreportcard]

Package sets provides generic implementations of sorted and unsorted sets,
or collections of unique elements.


## Interface

```go
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

// Sorted is a set whose elements are sorted.
// Elems and Range will return the elements in sorted order.
type Sorted[E any] interface {
	Set[E]
}
```


## Unsorted Sets

```go
// New returns a set initialized with the given elements.
func New[E comparable](elems ...E) Set[E]
```


## Sorted Sets

```go
// NewSorted returns a sorted set initialized with the given elements.
func NewSorted[E constraints.Ordered](elems ...E) Sorted[E]

// NewSortedCmpFunc returns a sorted set initialized with the given elements.
// The comparison function is used to order and identify elements.
func NewSortedCmpFunc[E any](cmp CmpFunc[E], elems ...E) Sorted[E]

// NewSortedCmpEqFunc returns a sorted set initialized with the given elements.
// The comparison function is only used to order elements and the equality
// function is used to identify elements.
//
// It may contain unique elements for which cmp(a, b) == 0 and eq(a, b) == false.
func NewSortedCmpEqFunc[E any](cmp CmpFunc[E], eq EqFunc[E], elems ...E) Sorted[E]

// A CmpFunc is a comparison function.
// It returns 1 if a is greater than b.
// It returns -1 if a is less than b.
// Otherwise, it returns 0.
type CmpFunc[E any] func(a, b E) int

// An EqFunc is an equality function.
// It returns true if and only if a and b are identical.
type EqFunc[E any] func(a, b E) bool
```

[license]: https://raw.githubusercontent.com/abursavich/sets/main/LICENSE
[license-img]: https://img.shields.io/badge/license-mit-blue.svg?style=for-the-badge

[godev]: https://pkg.go.dev/bursavich.dev/sets
[godev-img]: https://img.shields.io/static/v1?logo=go&logoColor=white&color=00ADD8&label=dev&message=reference&style=for-the-badge

[goreportcard]: https://goreportcard.com/report/bursavich.dev/sets
[goreportcard-img]: https://goreportcard.com/badge/bursavich.dev/sets?style=for-the-badge
