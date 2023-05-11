// Package intset provides IntSet, a compact and fast representation
// for sets of non-negative int values.
//
// The time complexity of the operations Add, Remove and Has
// is in O(1) in practice those methods are faster and more
// space-efficient than equivalent operations on sets based on the Go
// map type. The Len, IsEmpty, Min, Max, and TakeMin operations
// require O(n).
package intset

import (
	"fmt"
	"math/bits"
	"strings"
	"unsafe"
)

// Limit values of implementation-specific int type.
const (
	intSize = 32 << (^uint(0) >> 63)
	MaxInt  = 1<<(intSize-1) - 1
	MinInt  = -MaxInt - 1
)

const (
	wordSize    = 32 << (^uint(0) >> 63)
	lg2WordSize = 4 + 1<<(^uint(0)>>63)
	bitmask     = 1<<lg2WordSize - 1
)

// wordMask returns the word index (in IntSet.words)
// and single-bit mask for the IntSet's ith bit.
func wordMask(x int) (w int, mask uint) {
	return x >> lg2WordSize, 1 << uint(x&bitmask)
}

// wordBit returns the word index (in IntSet.words)
// and bit index for the IntSet's ith bit.
func wordBit(x int) (w int, bit uint) {
	return x >> lg2WordSize, uint(x & bitmask)
}

// IntSet is a set of small non-negative int values.
//
// The zero value represents a valid empty set.
//
// IntSet must be copied using the Copy method, not by assigning
// a IntSet value.
type IntSet[E ~int] struct {
	words []uint
}

// Has reports whether the set s contains the non-negative value x.
func (s *IntSet[E]) Has(x E) bool {
	w, mask := wordMask(int(x))
	return w < len(s.words) && s.words[w]&mask != 0
}

// Add adds the non-negative value x to the set s, and reports whether the set grew.
func (s *IntSet[E]) Add(x E) bool {
	w, mask := wordMask(int(x))

	if w < len(s.words) && s.words[w]&mask != 0 {
		return false
	}

	for len(s.words) <= w {
		s.words = append(s.words, 0)
	}

	s.words[w] |= mask
	return true
}

// AddAll adds a group of non-negative value xs to the set.
func (s *IntSet[E]) AddAll(xs ...E) {
	for _, x := range xs {
		s.Add(x)
	}
}

// Remove remove x from the set s, and reports whether the set shrank.
func (s *IntSet[E]) Remove(x E) bool {
	w, mask := wordMask(int(x))
	if w >= len(s.words) || s.words[w]&mask == 0 {
		return false
	}

	s.words[w] &^= mask
	return true
}

// Len return the number of elements
func (s *IntSet[E]) Len() int {
	n := 0

	for _, w := range s.words {
		if w == 0 {
			continue
		}

		n += popcount(w)
	}

	return n
}

// IsEmpty reports whether the set s is empty.
func (s *IntSet[E]) IsEmpty() bool {
	for _, w := range s.words {
		if w != 0 {
			return false
		}
	}

	return true
}

// AppendTo returns the result of appending the elements of s to slice in order.
func (s *IntSet[E]) AppendTo(slice []E) []E {
	total := len(slice) + s.Len()
	if total > cap(slice) {
		newSlice := make([]E, total)
		n := copy(newSlice, slice)
		slice = newSlice[:n]
	}

	elems := slice[len(slice):total]
	i := 0
	s.forEach(func(x E) {
		elems[i] = x
		i++
	})

	return slice[:total]
}

// Elems return the elements of the set s in order.
func (s *IntSet[E]) Elems() []E {
	return s.AppendTo(nil)
}

// TakeMin sets *p to the minimum element of the set s,
// removes that element from the set and returns true If set s is non-empty.
// Otherwise, it returns false and *p is undefined.
//
// This method may be used for iteration over a worklist like so:
//
// var x int
// for worklist.TakeMin(&x) { use(x) }
func (s *IntSet[E]) TakeMin(p *E) bool {
	for i, w := range s.words {
		if w == 0 {
			continue
		}

		tz := ntz(w)
		s.words[i] &^= (1 << uint(tz))
		*p = E(wordSize*i + tz)
		return true
	}

	return false
}

// Clear remove all elements from the set s.
func (s *IntSet[E]) Clear() {
	s.words = nil
}

// Copy return a copy of the set s.
func (s *IntSet[E]) Copy() *IntSet[E] {
	sc := &IntSet[E]{
		words: make([]uint, len(s.words)),
	}
	copy(sc.words, s.words)

	return sc
}

// String returns a human-readable description of the set s.
func (s *IntSet[E]) String() string {
	var b strings.Builder

	b.WriteByte('{')
	s.forEach(func(x E) {
		if b.Len() > len("{") {
			b.WriteByte(' ')
		}

		var xi any = x
		if xs, ok := xi.(fmt.Stringer); ok {
			fmt.Fprint(&b, xs.String())
		} else {
			fmt.Fprintf(&b, "%d", int(x))
		}
	})
	b.WriteByte('}')

	return b.String()
}

// forEach applies function f to each element of the set s in order.
//
// f must not mutate s. Consequently, forEach is not to expose
// to clients. In any case, using "for x := range s.AppendTo(nil) { doSomethingWith(x) }" allows more
// natural control flow with continue/break/return.
func (s *IntSet[E]) forEach(f func(E)) {
	for i, w := range s.words {
		if w == 0 {
			continue
		}

		for j := 0; j < wordSize; j++ {
			if w&(1<<uint(j)) != 0 {
				f(E(wordSize*i + j))
			}
		}
	}
}

// BitString returns the set as a string of 1s and 0s denoting the sum
// of the x'th powers of 2, for each x in s.
//
// Examples:
//
//	        {}.BitString() = "0"
//	     {4,5}.BitString() = "110000"
//	{0,4,5}.BitString() = "110001"
func (s *IntSet[E]) BitString() string {
	if s.IsEmpty() {
		return "0"
	}

	n := int(s.Max())
	n++ // zero bit
	radix := n

	b := make([]byte, n)
	for i := range b {
		b[i] = '0'
	}

	s.forEach(func(x E) {
		b[radix-int(x)-1] = '1'
	})

	return *(*string)(unsafe.Pointer(&b))
}

// Equals reports whether the sets s and t have the same elements.
func (s *IntSet[E]) Equals(t *IntSet[E]) bool {
	if s == t {
		return true
	}

	if s.Len() != t.Len() {
		return false
	}

	for i, w := range s.words {
		if w == 0 {
			continue
		}

		if i >= len(t.words) {
			return false
		}

		if w != t.words[i] {
			return false
		}
	}

	return true
}

// LowerBound returns the smallest element >= x, or MaxInt if there is no such element.
func (s *IntSet[E]) LowerBound(x E) E {
	w, bit := wordBit(int(x))

	for i, word := range s.words {
		if i < w {
			continue
		}

		if word == 0 {
			continue
		}

		tz := ntz(word)
		if i > w {
			return E(wordSize*i + tz)
		}

		if uint(tz) >= bit {
			return E(wordSize*i + tz)
		}

		for j := bit; j < wordSize; j++ {
			if word&(1<<j) != 0 {
				return E(wordSize*i + int(j))
			}
		}
	}

	return MaxInt
}

// Max returns the maximum element of the set s, or MinInt if s is empty.
func (s *IntSet[E]) Max() E {
	for i := len(s.words) - 1; i > -1; i-- {
		w := s.words[i]
		if w == 0 {
			continue
		}

		return E(wordSize*(i+1) - nlz(w) - 1)
	}

	return MinInt
}

// Min returns the minimum element of the set s, or MaxInt if s is empty.
func (s *IntSet[E]) Min() E {
	for i, w := range s.words {
		if w == 0 {
			continue
		}

		return E(wordSize*i + ntz(w))
	}

	return MaxInt
}

// UnionWith sets s to the union s ∪ t.
func (s *IntSet[E]) UnionWith(t *IntSet[E]) {
	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] |= tword
		} else {
			s.words = append(s.words, tword)
		}
	}
}

// IntersectWith sets s to the intersection s ∩ t.
func (s *IntSet[E]) IntersectWith(t *IntSet[E]) {
	for i := range s.words {
		if i < len(t.words) {
			s.words[i] &= t.words[i]
		} else {
			s.words[i] = 0
		}
	}
}

// Intersects reports whether s ∩ x ≠ ∅.
func (s *IntSet[E]) Intersects(t *IntSet[E]) bool {
	for i, tword := range t.words {
		if tword == 0 {
			continue
		}

		if i >= len(s.words) {
			break
		}

		if s.words[i]&tword != 0 {
			return true
		}
	}

	return false
}

// DifferenceWith sets s to the difference s ∖ t.
func (s *IntSet[E]) DifferenceWith(t *IntSet[E]) {
	if s == t {
		s.Clear()
		return
	}

	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] &^= tword
		}
	}
}

// SymmetricDifference sets s to the symmetric difference s ∆ t.
func (s *IntSet[E]) SymmetricDifference(t *IntSet[E]) {
	for i, tword := range t.words {
		if i < len(s.words) {
			s.words[i] ^= tword
		} else {
			s.words = append(s.words, tword)
		}
	}
}

// SubsetOf reports whether s ∖ t = ∅.
func (s *IntSet[E]) SubsetOf(t *IntSet[E]) bool {
	for i, word := range s.words {
		if word == 0 {
			continue
		}

		if i >= len(t.words) {
			return false
		}

		if word&^t.words[i] != 0 {
			return false
		}
	}

	return true
}

// popcount returns the number of set bits in w.
func popcount(w uint) int {
	return bits.OnesCount(w)
}

// nlz returns the number of leading zeros of x.
func nlz(x uint) int {
	return bits.LeadingZeros(x)
}

// ntz returns the number of trailing zeros of x.
func ntz(x uint) int {
	return bits.TrailingZeros(x)
}
