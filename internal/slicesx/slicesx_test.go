// SPDX-License-Identifier: MIT
//
// Copyright 2023 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package slicesx

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	compare "github.com/google/go-cmp/cmp"
)

func cmpPtrVal[T cmp.Ordered](a, b *T) int {
	return cmp.Compare(*a, *b)
}

func equal[T comparable](a, b T) bool { return a == b }

func runePtrsFrom(data string) func(idx ...int) []*rune {
	r := []rune(data)
	return func(idx ...int) []*rune {
		s := make([]*rune, len(idx))
		for j, k := range idx {
			s[j] = &r[k]
		}
		return s
	}
}

func runePtrsString(s []*rune) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if s == nil {
			b.WriteString("<nil>")
		} else {
			b.WriteRune(*r)
		}
	}
	return b.String()
}

func TestMergeSortedLists(t *testing.T) {
	runes := runePtrsFrom("aaabbbcccdddeee")
	for _, tt := range []struct {
		a, b []*rune
		want []*rune
	}{
		{},
		{
			a:    runes(0),
			want: runes(0),
		},
		{
			b:    runes(0),
			want: runes(0),
		},
		{
			a:    runes(0),
			b:    runes(0),
			want: runes(0),
		},
		{
			a:    runes(0, 1),
			b:    runes(0, 1),
			want: runes(0, 1),
		},
		{
			a:    runes(0, 1),
			b:    runes(1, 2),
			want: runes(0, 1, 2),
		},
		{
			a:    runes(3, 4, 5),
			b:    runes(0, 1, 2),
			want: runes(0, 1, 2, 3, 4, 5),
		},
		{
			a:    runes(0, 1, 2),
			b:    runes(3, 4, 5),
			want: runes(0, 1, 2, 3, 4, 5),
		},
	} {
		t.Run(runePtrsString(tt.a), func(t *testing.T) {
			got := MergeSorted(slices.Clone(tt.a), slices.Clone(tt.b), cmpPtrVal[rune], equal[*rune])
			if diff := compare.Diff(got, tt.want); diff != "" {
				t.Fatal("Unexpected diff: \n", diff)
			}
		})
	}
}

func TestRunEq(t *testing.T) {
	for _, tt := range []struct {
		elems string
		want  string
	}{
		{},
		{
			elems: "a",
			want:  "a",
		},
		{
			elems: "aaa",
			want:  "aaa",
		},
		{
			elems: "abcdef",
			want:  "a",
		},
		{
			elems: "aaabcdef",
			want:  "aaa",
		},
	} {
		t.Run(tt.elems, func(t *testing.T) {
			if got := string(runEq([]rune(tt.elems), cmp.Compare[rune])); got != tt.want {
				t.Errorf("runEq(%q): got: %q; want: %q", tt.elems, got, tt.want)
			}
		})
	}
}

func TestStableSortUniq(t *testing.T) {
	runes := runePtrsFrom("aabbccddee")
	for _, tt := range []struct {
		elems []*rune
		want  []*rune
	}{
		{},
		{
			elems: runes(0),
			want:  runes(0),
		},
		{
			elems: runes(0, 0, 0),
			want:  runes(0),
		},
		{
			elems: runes(0, 1, 2, 3, 4, 5, 6, 7, 8, 9),
			want:  runes(0, 1, 2, 3, 4, 5, 6, 7, 8, 9),
		},
		{
			elems: runes(0, 0, 2, 4, 4, 6, 8, 8),
			want:  runes(0, 2, 4, 6, 8),
		},
	} {

		t.Run(runePtrsString(tt.elems), func(t *testing.T) {
			got := StableSortUniqFuncs(slices.Clone(tt.elems), cmpPtrVal[rune], equal[*rune])
			if diff := compare.Diff(got, tt.want); diff != "" {
				t.Fatal("Unexpected diff: \n", diff)
			}
		})
	}
}
