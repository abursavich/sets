package slicesx

import (
	"cmp"
	"slices"
)

// A CmpFunc is a comparison function.
// It returns 1 if a is greater than b.
// It returns -1 if a is less than b.
// Otherwise, it returns 0.
type CmpFunc[E any] func(a, b E) int

// An EqFunc is an equality function.
// It returns true if and only if a and b are identical.
type EqFunc[E any] func(a, b E) bool

type insert[E any] struct {
	i int
	e E
}

// MergeSortedUniq merges B into A, both of which must be sorted and contain unique values.
func MergeSortedUniq[E cmp.Ordered](a, b []E) []E {
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

// MergeSorted merges B into A, both of which must be sorted.
func MergeSorted[E any](a, b []E, cmp CmpFunc[E], eq EqFunc[E]) []E {
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

// DeleteSortedUniq deletes B from A (e.g. A - B),
// both of which must be sorted and contain unique values.
func DeleteSortedUniq[E cmp.Ordered](a, b []E) []E {
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

// DeleteSorted deletes B from A (e.g. A - B), both of which must be sorted.
func DeleteSorted[E any](a, b []E, cmp CmpFunc[E], eq EqFunc[E]) []E {
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
	clear(a[n-d:])
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

// StableSortUniqFuncs stable sorts the list and removes duplicates in place.
// It uses O(n*log(n)) compares and O(n*log(n)*log(n)) swaps for the sort.
// It uses O(n) compares and up to O(n^2) eqs for the uniq
// Elements may be ordered the same but unequal (e.g. cmp(a, b) == 0 && !eq(a, b))..
func StableSortUniqFuncs[T any](list []T, cmp func(T, T) int, eq func(T, T) bool) []T {
	return UniqSortedFuncs(StableSortFunc(list, cmp), cmp, eq)
}

// StableSort stable sorts the list using O(n*log(n)) compares and O(n*log(n)*log(n)) swaps.
func StableSort[E cmp.Ordered](list []E) []E {
	slices.SortStableFunc(list, cmp.Compare[E])
	return list
}

// StableSortFunc stable sorts the list using O(n*log(n)) compares and O(n*log(n)*log(n)) swaps.
func StableSortFunc[T any](list []T, cmp func(T, T) int) []T {
	slices.SortStableFunc(list, cmp)
	return list
}

// UniqSorted removes duplicate elements from the sorted list in place
// and preserves order using O(n) compares.
func UniqSorted[E comparable](sorted []E) []E {
	return slices.Compact(sorted)
}

// UniqSortedFunc removes duplicate elements from the sorted list in place
// and preserves order using O(n) compares.
func UniqSortedFunc[T any](sorted []T, eq func(T, T) bool) []T {
	return slices.CompactFunc(sorted, eq)
}

// UniqSortedFuncs removes duplicate elements from the sorted list in place
// and preserves order using O(n) compares and up to O(n^2) eqs.
// Elements may be ordered the same but unequal (e.g. cmp(a, b) == 0 && !eq(a, b)).
func UniqSortedFuncs[T any](sorted []T, cmp func(T, T) int, eq func(T, T) bool) []T {
	n := len(sorted)
	if n == 0 {
		return nil
	}
	src := 0
	dst := 0
	prev := sorted[dst]
	for i := 1; i < n; i++ {
		next := sorted[i]
		if cmp(next, prev) == 0 {
			continue
		}
		dst += copy(sorted[dst:], uniqEqSlow(sorted[src:i], eq))
		src = i
		prev = next
	}
	dst += copy(sorted[dst:], uniqEqSlow(sorted[src:], eq))
	clear(sorted[dst:])
	return sorted[:dst]
}

// uniqEqSlow removes duplicate elements in place and preserves order using up to O(n^2) eqs.
func uniqEqSlow[T any](list []T, eq func(T, T) bool) []T {
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
	clear(list[dst:])
	return list[:dst]
}
