package intset_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"unsafe"

	"github.com/google/go-cmp/cmp"
)

// This file contains reference set implementations for unit-tests.

// setInterface is the interface IntSet implements.
type setInterface interface {
	Add(int) bool
	AddAll(...int)
	Has(int) bool
	Remove(int) bool

	AppendTo([]int) []int
	Elems() []int

	Clear()

	Copy() any

	IsEmpty() bool
	Len() int

	LowerBound(int) int
	Max() int
	Min() int

	String() string
	BitString() string

	TakeMin(*int) bool

	UnionWith(any)
	IntersectWith(any)
	Intersects(any) bool
	DifferenceWith(any)
	SymmetricDifference(any)
	SubsetOf(any) bool
	Equals(any) bool
}

var _ setInterface = &MapSet{}

const (
	intSize = 32 << (^uint(0) >> 63)
	MaxInt  = 1<<(intSize-1) - 1
	MinInt  = -MaxInt - 1
)

// MapSet is an implementation of setInterface using a Go map.
type MapSet struct {
	m map[int]bool
}

func (s *MapSet) Add(x int) bool {
	if s.m[x] {
		return false
	}

	s.init()
	s.m[x] = true
	return true
}

func (s *MapSet) AddAll(xs ...int) {
	for _, x := range xs {
		s.Add(x)
	}
}

func (s *MapSet) AppendTo(slice []int) []int {
	total := len(slice) + len(s.m)
	if total > cap(slice) {
		newSlice := make([]int, total)
		n := copy(newSlice, slice)
		slice = newSlice[:n]
	}

	elems := slice[len(slice):total]
	i := 0
	s.forEach(func(x int) {
		elems[i] = x
		i++
	})

	return slice[:total]
}

func (s *MapSet) Elems() []int {
	return s.AppendTo(nil)
}

func (s *MapSet) Clear() {
	s.m = nil
}

func (s *MapSet) Copy() any {
	sc := &MapSet{
		m: make(map[int]bool, len(s.m)),
	}

	s.forEach(func(x int) {
		sc.m[x] = true
	})

	return sc
}

func (s *MapSet) Has(x int) bool {
	return s.m[x]
}

func (s *MapSet) IsEmpty() bool {
	return len(s.m) == 0
}

func (s *MapSet) Len() int {
	return len(s.m)
}

func (s *MapSet) LowerBound(x int) int {
	min := MaxInt

	s.forEach(func(k int) {
		if k >= x && k < min {
			min = k
		}
	})

	return min
}

func (s *MapSet) Max() int {
	if s.IsEmpty() {
		return MinInt
	}

	return s.AppendTo(nil)[s.Len()-1]
}

func (s *MapSet) Min() int {
	if s.IsEmpty() {
		return MaxInt
	}

	return s.AppendTo(nil)[0]
}

func (s *MapSet) Remove(x int) bool {
	if !s.m[x] {
		return false
	}

	delete(s.m, x)
	return true
}

func (s *MapSet) String() string {
	elems := make([]int, len(s.m))
	i := 0
	s.forEach(func(x int) {
		elems[i] = x
		i++
	})

	var b strings.Builder

	b.WriteByte('{')
	for _, x := range elems {
		if b.Len() > len("{") {
			b.WriteByte(' ')
		}

		fmt.Fprintf(&b, "%d", x)
	}
	b.WriteByte('}')

	return b.String()
}

func (s *MapSet) BitString() string {
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

	s.forEach(func(x int) {
		b[radix-x-1] = '1'
	})

	return *(*string)(unsafe.Pointer(&b))
}

func (s *MapSet) TakeMin(p *int) bool {
	if s.IsEmpty() {
		return false
	}

	*p = s.AppendTo(nil)[0]
	delete(s.m, *p)
	return true
}

func (s *MapSet) UnionWith(t any) {
	s.init()

	for x := range t.(*MapSet).m {
		s.m[x] = true
	}
}

func (s *MapSet) IntersectWith(a any) {
	t := a.(*MapSet)

	for x := range s.m {
		if !t.m[x] {
			delete(s.m, x)
		}
	}
}

func (s *MapSet) Intersects(a any) bool {
	t := a.(*MapSet)

	for x := range s.m {
		if t.m[x] {
			return true
		}
	}

	return false
}

func (s *MapSet) DifferenceWith(a any) {
	t := a.(*MapSet)

	for x := range s.m {
		if t.m[x] {
			delete(s.m, x)
		}
	}
}

func (s *MapSet) SymmetricDifference(a any) {
	t := a.(*MapSet)

	for x := range s.m {
		if t.m[x] {
			// mark later delete
			s.m[x] = false
		}
	}

	s.init()
	for x := range t.m {
		if _, ok := s.m[x]; !ok {
			s.m[x] = true
		} else {
			delete(s.m, x)
		}
	}
}

func (s *MapSet) SubsetOf(a any) bool {
	t := a.(*MapSet)

	for x := range s.m {
		if !t.m[x] {
			return false
		}
	}

	return true
}

func (s *MapSet) Equals(a any) bool {
	t := a.(*MapSet)

	if len(s.m) != len(t.m) {
		return false
	}

	for x := range s.m {
		if !t.m[x] {
			return false
		}
	}

	return true
}

func (s *MapSet) init() {
	if s.m == nil {
		s.m = make(map[int]bool)
	}
}

// forEach applies function f to each element of the set s in order.
func (s *MapSet) forEach(f func(int)) {
	xs := make([]int, len(s.m))
	i := 0
	for x := range s.m {
		xs[i] = x
		i++
	}
	sort.Ints(xs)

	for _, x := range xs {
		f(x)
	}
}

// mapSetCall is a quick.Generator for calls on MapSet.
type mapSetCall struct {
	s, t []int
}

func (mapSetCall) Generate(r *rand.Rand, size int) reflect.Value {
	c := mapSetCall{s: randValues(r), t: randValues(r)}
	return reflect.ValueOf(c)
}

func applyMapSetCalls(f func(*MapSet, any) (any, bool), calls []mapSetCall) (results []setResult) {
	for _, c := range calls {
		s, t := &MapSet{}, &MapSet{}
		for _, x := range c.s {
			s.Add(x)
		}
		for _, x := range c.t {
			t.Add(x)
		}

		v, ok := f(s, t)
		results = append(results, setResult{v, ok})
	}

	return results
}

func TestIntersectsMatches(t *testing.T) {
	t.Parallel()

	f := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			s.IntersectWith(t)
			return nil, !s.IsEmpty()
		}, calls)
	}

	g := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			return nil, s.Intersects(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestSubsetOfMatches(t *testing.T) {
	t.Parallel()

	f := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			s.DifferenceWith(t)
			return nil, s.IsEmpty()
		}, calls)
	}

	g := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			return nil, s.SubsetOf(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestEqualsMatches(t *testing.T) {
	t.Parallel()

	f := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			return nil, s.SubsetOf(t) && t.(*MapSet).SubsetOf(s)
		}, calls)
	}

	g := func(calls []mapSetCall) []setResult {
		return applyMapSetCalls(func(s *MapSet, t any) (any, bool) {
			return nil, s.Equals(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		cerr := err.(*quick.CheckEqualError)
		t.Error(cmp.Diff(cerr.Out1, cerr.Out2))
		t.Error(cerr.In)
	}
}
