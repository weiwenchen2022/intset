// Package intset provides a set of integers based on a bit vector.
package intset

import (
	"fmt"
	"math/bits"
	"strings"
)

const (
	wordSize    = 32 << (^uint(0) >> 63)
	lg2WordSize = 4 + 1<<(^uint(0)>>63)
	bitmask     = 1<<lg2WordSize - 1
)

func split[E ~int](x E) (word int, bit uint) {
	return int(x) >> lg2WordSize, uint(x & bitmask)
}

// IntSet is a set of small non-negative integers.
// Its zero value represents the empty set.
type IntSet[E ~int] struct {
	words []uint
}

// Has reports whether the set contains the non-negative value x.
func (s *IntSet[E]) Has(x E) bool {
	word, bit := split(x)
	return word < len(s.words) && s.words[word]&(1<<bit) != 0
}

// Add adds the non-negative value x to the set.
func (s *IntSet[E]) Add(x E) {
	word, bit := split(x)
	for len(s.words) <= word {
		s.words = append(s.words, 0)
	}

	s.words[word] |= 1 << bit
}

// AddAll adds a group of non-negative value xs to the set.
func (s *IntSet[E]) AddAll(xs ...E) {
	for _, x := range xs {
		s.Add(x)
	}
}

// Remove remove x from the set
func (s *IntSet[E]) Remove(x E) {
	word, bit := split(x)
	if word >= len(s.words) {
		return
	}

	s.words[word] &= ^uint(1 << bit)
}

// Len return the number of elements
func (s *IntSet[E]) Len() int {
	n := 0

	for _, word := range s.words {
		if word == 0 {
			continue
		}

		// n += kernighan(uint64(word))
		// n += hackersdelight(uint64(word))
		n += bits.OnesCount64(uint64(word))
	}

	return n
}

func kernighan(x uint64) int {
	count := 0

	for ; x > 0; x &= (x - 1) {
		count++
	}

	return count
}

func hackersdelight(x uint64) int {
	const (
		m1  = 0x5555555555555555
		m2  = 0x3333333333333333
		m4  = 0x0f0f0f0f0f0f0f0f
		h01 = 0x0101010101010101
	)

	x -= (x >> 1) & m1
	x = (x & m2) + ((x >> 2) & m2)
	x = (x + (x >> 4)) & m4
	return int((x * h01) >> 56)
}

// Elems return the elements of set
func (s *IntSet[E]) Elems() []E {
	elems := make([]E, s.Len())
	n := 0

	for i, word := range s.words {
		if word == 0 {
			continue
		}

		for j := 0; j < wordSize; j++ {
			if word&(1<<uint(j)) != 0 {
				elems[n] = E(wordSize*i + j)
				n++
			}
		}
	}

	return elems
}

// Clear remove all elements from the set
func (s *IntSet[E]) Clear() {
	s.words = nil
}

// Copy return a copy of the set
func (s *IntSet[E]) Copy() *IntSet[E] {
	sc := &IntSet[E]{
		words: make([]uint, len(s.words)),
	}
	copy(sc.words, s.words)

	return sc
}

// String returns the set as a string of the form "{1 2 3}".
func (s *IntSet[E]) String() string {
	var b strings.Builder

	b.WriteByte('{')
	for i, word := range s.words {
		if word == 0 {
			continue
		}

		for j := 0; j < wordSize; j++ {
			if word&(1<<uint(j)) != 0 {
				if b.Len() > len("{") {
					b.WriteByte(' ')
				}

				var ei any = E(wordSize*i + j)
				if elem, ok := ei.(fmt.Stringer); ok {
					fmt.Fprint(&b, elem.String())
				} else {
					fmt.Fprintf(&b, "%d", wordSize*i+j)
				}
			}
		}
	}
	b.WriteByte('}')

	return b.String()
}

// UnionWith sets s to the union of s and t.
func (s *IntSet[E]) UnionWith(t *IntSet[E]) {
	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] |= tword
		} else {
			s.words = append(s.words, tword)
		}
	}
}

// IntersectWith sets s to the intersection of s and t.
func (s *IntSet[E]) IntersectWith(t *IntSet[E]) {
	for i := range s.words {
		if i < len(t.words) {
			s.words[i] &= t.words[i]
		} else {
			s.words[i] = 0
		}
	}
}

// DifferenceWith sets s to the difference of s and t.
func (s *IntSet[E]) DifferenceWith(t *IntSet[E]) {
	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] &^= tword
		}
	}
}

// SymmetricDifference sets s to the symmetric difference of s and t.
func (s *IntSet[E]) SymmetricDifference(t *IntSet[E]) {
	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] ^= tword
		} else {
			s.words = append(s.words, tword)
		}
	}
}
