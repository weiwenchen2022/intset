package intset_test

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/weiwenchen2022/intset"

	"github.com/google/go-cmp/cmp"
)

type IntSet struct {
	intset.IntSet[int]
}

func (s *IntSet) Copy() any {
	sc := &IntSet{*s.IntSet.Copy()}
	return sc
}

func (s *IntSet) UnionWith(a any) {
	t := a.(*IntSet)
	s.IntSet.UnionWith(&t.IntSet)
}

func (s *IntSet) IntersectWith(a any) {
	t := a.(*IntSet)
	s.IntSet.IntersectWith(&t.IntSet)
}

func (s *IntSet) Intersects(a any) bool {
	t := a.(*IntSet)
	return s.IntSet.Intersects(&t.IntSet)
}

func (s *IntSet) DifferenceWith(a any) {
	t := a.(*IntSet)
	s.IntSet.DifferenceWith(&t.IntSet)
}

func (s *IntSet) SymmetricDifference(a any) {
	t := a.(*IntSet)
	s.IntSet.SymmetricDifference(&t.IntSet)
}

func (s *IntSet) SubsetOf(a any) bool {
	t := a.(*IntSet)
	return s.IntSet.SubsetOf(&t.IntSet)
}

func (s *IntSet) Equals(a any) bool {
	t := a.(*IntSet)
	return s.IntSet.Equals(&t.IntSet)
}

type setOp string

const (
	opAdd    = setOp("Add")
	opAddAll = setOp("AddAll")

	opAppendTo = setOp("AppendTo")
	opElems    = setOp("Elems")

	opClear = setOp("Clear")
	opCopy  = setOp("Copy")

	opHas = setOp("Has")

	opIsEmpty = setOp("IsEmpty")
	opLen     = setOp("Len")

	opLowerBound = setOp("LowerBound")
	opMax        = setOp("Max")
	opMin        = setOp("Min")

	opRemove = setOp("Remove")

	opString    = setOp("String")
	opBitString = setOp("BitString")

	opTakeMin = setOp("TakeMin")

	opUnionWith           = setOp("UnionWith")
	opIntersectWith       = setOp("IntersectionWith")
	opIntersects          = setOp("Intersects")
	opDifferenceWith      = setOp("DifferenceWith")
	opSymmetricDifference = setOp("SymmetricDifference")
	opSubsetOf            = setOp("SubsetOf")
	opEquals              = setOp("Equals")
)

var setOps = [...]setOp{
	opAdd,
	opAddAll,
	opHas,
	opRemove,

	opAppendTo,
	opElems,

	opClear,
	opCopy,

	opIsEmpty,
	opLen,

	opLowerBound,
	opMax,
	opMin,

	opString,
	opBitString,

	opTakeMin,

	opUnionWith,
	opIntersectWith,
	opIntersects,
	opDifferenceWith,
	opSymmetricDifference,
	opSubsetOf,
	opEquals,
}

// setCall is a quick.Generator for calls on setInterface.
type setCall struct {
	op setOp
	x  int
	xs []int
	t  []int
}

func (setCall) Generate(r *rand.Rand, size int) reflect.Value {
	c := setCall{op: setOps[r.Intn(len(setOps))]}

	switch c.op {
	case opAdd, opHas, opLowerBound, opRemove:
		c.x = randValue(r)
	case opAddAll:
		xs := make([]int, r.Intn(4))
		for i := range xs {
			xs[i] = randValue(r)
		}
		c.xs = xs
	case opUnionWith, opIntersectWith, opIntersects,
		opDifferenceWith, opSymmetricDifference,
		opSubsetOf, opEquals:
		t := make([]int, r.Intn(4))
		for i := range t {
			t[i] = randValue(r)
		}
		c.t = t
	}

	return reflect.ValueOf(c)
}

func randValue(r *rand.Rand) int {
	return r.Intn(10000)
}

func randValues(r *rand.Rand) []int {
	s := make([]int, 1<<r.Intn(12))

	for i := range s {
		s[i] = r.Intn(1 << 12)
	}

	return s
}

func (c setCall) apply(s setInterface) (any, bool) {
	switch c.op {
	case opAdd:
		return c.x, s.Add(c.x)
	case opAddAll:
		s.AddAll(c.xs...)
		return s.String(), true
	case opAppendTo:
		return s.AppendTo(nil), true
	case opElems:
		return s.Elems(), true
	case opClear:
		s.Clear()
		return nil, false
	case opCopy:
		return s.Copy().(setInterface).String(), true
	case opHas:
		return c.x, s.Has(c.x)
	case opIsEmpty:
		return nil, s.IsEmpty()
	case opLen:
		return s.Len(), true
	case opLowerBound:
		return s.LowerBound(c.x), true
	case opMax:
		return s.Max(), true
	case opMin:
		return s.Min(), true
	case opRemove:
		return c.x, s.Remove(c.x)
	case opString:
		return s.String(), true
	case opBitString:
		return s.BitString(), true
	case opTakeMin:
		var x int
		ok := s.TakeMin(&x)
		return x, ok
	case opUnionWith, opIntersectWith, opIntersects,
		opDifferenceWith, opSymmetricDifference,
		opSubsetOf, opEquals:
		t := reflect.New(reflect.TypeOf(s).Elem()).Interface().(setInterface)
		for x := range c.t {
			t.Add(x)
		}

		switch c.op {
		case opUnionWith:
			s.UnionWith(t)
		case opIntersectWith:
			s.IntersectWith(t)
		case opIntersects:
			return nil, s.Intersects(t)
		case opDifferenceWith:
			s.DifferenceWith(t)
		case opSymmetricDifference:
			s.SymmetricDifference(t)
		case opSubsetOf:
			return nil, s.SubsetOf(t)
		case opEquals:
			return nil, s.Equals(t)
		}

		return s.String(), true
	default:
		panic("invalid setOp " + c.op)
	}
}

type setResult struct {
	Value any
	Ok    bool
}

func applyCalls(s setInterface, calls []setCall) (result []setResult, final []int) {
	for _, c := range calls {
		v, ok := c.apply(s)
		result = append(result, setResult{v, ok})
	}

	final = s.AppendTo(nil)
	return result, final
}

func applyIntSet(calls []setCall) ([]setResult, []int) {
	return applyCalls(&IntSet{}, calls)
}

func applyMapSet(calls []setCall) ([]setResult, []int) {
	return applyCalls(&MapSet{}, calls)
}

func TestIntSetMatchesMapSet(t *testing.T) {
	t.Parallel()

	if err := quick.CheckEqual(applyMapSet, applyIntSet, nil); err != nil {
		t.Error(err)
	}
}

func TestBasics(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]

	if l := s.Len(); l != 0 {
		t.Errorf("Len({}): got %d, want 0", l)
	}

	if s := s.String(); s != "{}" {
		t.Errorf("String({}): got %q, want \"{}\"", s)
	}

	if s.Has(3) {
		t.Errorf("Has(3): got true, want false")
	}

	if !s.Add(3) {
		t.Errorf("Add(3): got false, want true")
	}

	if max := s.Max(); max != 3 {
		t.Errorf("Max: got %d, want 3", max)
	}

	if !s.Add(435) {
		t.Errorf("Add(435): got false, want true")
	}

	if s := s.String(); s != "{3 435}" {
		t.Errorf("String({3 435}): got %q, want \"{3 435}\"", s)
	}

	if max := s.Max(); max != 435 {
		t.Errorf("Max: got %d, want 435", max)
	}

	if l := s.Len(); l != 2 {
		t.Errorf("Len: got %d, want 2", l)
	}

	if !s.Remove(435) {
		t.Errorf("Remove(435): got false, want true")
	}

	if s := s.String(); s != "{3}" {
		t.Errorf("String({3}): got %q, want \"{3}\"", s)
	}
}

// Add, Len, IsEmpty, Hash, Clear, Elems, AppendTo.
func TestBasicsMore(t *testing.T) {
	t.Parallel()

	s := new(intset.IntSet[int])

	s.Add(456)
	s.Add(123)
	s.Add(789)
	if s.Len() != 3 {
		t.Errorf("%s.Len: got %d, want 3", s, s.Len())
	}

	if s.IsEmpty() {
		t.Errorf("%s.IsEmpty: got true", s)
	}

	if !s.Has(123) {
		t.Errorf("%s.Has(123): got false", s)
	}

	if s.Has(1234) {
		t.Errorf("%s.Has(1234): got true", s)
	}

	want := []int{123, 456, 789}
	got := s.Elems()
	if !cmp.Equal(want, got) {
		t.Errorf("%s.Elems: got %v, want %v", s, got, want)
	}

	want = []int{1, 123, 456, 789}
	got = s.AppendTo([]int{1})
	if !cmp.Equal(want, got) {
		t.Errorf("%s.AppendTo: got %v, want %v", s, got, want)
	}

	s.Clear()

	if l := s.Len(); l != 0 {
		t.Errorf("Clear: got %d, want 0", l)
	}

	if !s.IsEmpty() {
		t.Errorf("IsEmpty: got false")
	}

	if s.Has(123) {
		t.Errorf("%s.Has(123): got false", s)
	}
}

func TestTakeMin(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]

	s.Add(456)
	s.Add(123)
	s.Add(789)

	var got int
	for i, want := range []int{123, 456, 789} {
		if !s.TakeMin(&got) || want != got {
			t.Errorf("TakeMin #%d: got %d, want %d", i, got, want)
		}
	}

	if s.TakeMin(&got) {
		t.Errorf("%s.TakeMin returned true", &s)
	}
}

func TestMinAndMax(t *testing.T) {
	t.Parallel()

	values := []int{0, 456, 123, 789} // elt 0 => empty set
	wantMax := []int{intset.MinInt, 456, 456, 789}
	wantMin := []int{intset.MaxInt, 456, 123, 123}

	var s intset.IntSet[int]

	for i, x := range values {
		if i != 0 {
			s.Add(x)
		}

		if want, got := wantMin[i], s.Min(); want != got {
			t.Errorf("Min #%d: got %d, want %d", i, got, want)
		}

		if want, got := wantMax[i], s.Max(); want != got {
			t.Errorf("Max #%d: got %d, want %d", i, got, want)
		}
	}
}

// intSetCall is a quick.Generator for calls on intset.IntSet.
type intSetCall struct {
	s, t []int
}

func (intSetCall) Generate(r *rand.Rand, size int) reflect.Value {
	c := intSetCall{s: randValues(r), t: randValues(r)}
	return reflect.ValueOf(c)
}

func applyIntSetCalls(f func(s, t *intset.IntSet[int]) (any, bool), calls []intSetCall) (results []setResult) {
	for _, c := range calls {
		s, t := &intset.IntSet[int]{}, &intset.IntSet[int]{}
		for x := range c.s {
			s.Add(x)
		}
		for x := range c.t {
			t.Add(x)
		}

		v, ok := f(s, t)
		results = append(results, setResult{v, ok})
	}

	return results
}

func TestIntersects(t *testing.T) {
	t.Parallel()

	f := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			s.IntersectWith(t)
			return nil, !s.IsEmpty()
		}, calls)
	}

	g := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			return nil, s.Intersects(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestSubsetOf(t *testing.T) {
	t.Parallel()

	f := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			s.DifferenceWith(t)
			return nil, s.IsEmpty()
		}, calls)
	}

	g := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			return nil, s.SubsetOf(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestEquals(t *testing.T) {
	t.Parallel()

	f := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			return nil, s.SubsetOf(t) && t.SubsetOf(s)
		}, calls)
	}

	g := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			return nil, s.Equals(t)
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestSymmetricDifference(t *testing.T) {
	t.Parallel()

	f := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			sc := s.Copy()
			sc.DifferenceWith(t)
			t.DifferenceWith(s)
			sc.UnionWith(t)

			return sc.String(), true
		}, calls)
	}

	g := func(calls []intSetCall) []setResult {
		return applyIntSetCalls(func(s, t *intset.IntSet[int]) (any, bool) {
			s.SymmetricDifference(t)
			return s.String(), true
		}, calls)
	}

	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestHas(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	if !s.Has(9) {
		t.Error("no element 9")
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	want := "{1 9 144}"
	got := s.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAddAll(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.AddAll(1, 144, 9)

	want := "{1 9 144}"
	got := s.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(9)

	if !s.Has(9) {
		t.Error("no element 9")
	}

	s.Remove(9)
	if s.Has(9) {
		t.Error("element 9 not remove")
	}
}

func TestLen(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	if n := s.Len(); n != 3 {
		t.Errorf("Len report %d elements, but it should be %d", n, 3)
	}
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	var s1, s2 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	if s1.IsEmpty() {
		t.Errorf("expected s1 not empty set")
	}

	if !s2.IsEmpty() {
		t.Errorf("expected s2 empty set")
	}
}

func TestElems(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	want := []int{1, 9, 144}
	got := s.Elems()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestClear(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	if s.Len() != 3 {
		t.Error("miss some elements")
	}

	s.Clear()
	if s.Len() != 0 {
		t.Error("clear set failed")
	}
}

func TestCopy(t *testing.T) {
	t.Parallel()

	var s1 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	s2 := s1.Copy()
	want := s1.String()
	got := s2.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	s.Add(1)
	s.Add(144)
	s.Add(9)

	want := "{1 9 144}"
	got := s.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestBitString(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		s    []int
		want string
	}{
		{nil, "0"},
		{[]int{0}, "1"},
		{[]int{0, 4, 5}, "110001"},
		{[]int{0, 7, 177}, "1" + strings.Repeat("0", 169) + "10000001"},
		{[]int{0, 4, 5}, "110001"},
	}

	for _, tc := range testcases {
		var s intset.IntSet[int]
		for _, x := range tc.s {
			s.Add(x)
		}

		got := s.BitString()
		if !cmp.Equal(tc.want, got) {
			t.Error(cmp.Diff(tc.want, got))
		}
	}
}

func TestLowerBound(t *testing.T) {
	t.Parallel()

	var s1 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	want := 1
	got := s1.LowerBound(1)
	if want != got {
		t.Errorf("expected lowerBound %d, got %d", want, got)
	}

	want = 9
	got = s1.LowerBound(8)
	if want != got {
		t.Errorf("expected lowerBound %d, got %d", want, got)
	}

	want = intset.MaxInt
	got = s1.LowerBound(145)
	if want != got {
		t.Errorf("expected lowerBound %d, got %d", want, got)
	}
}

func TestUnionWith(t *testing.T) {
	t.Parallel()

	var s1, s2 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	s2.Add(9)
	s2.Add(42)

	want := "{1 9 42 144}"
	s1.UnionWith(&s2)
	got := s1.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestIntersectWith(t *testing.T) {
	t.Parallel()

	var s1, s2 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	s2.Add(9)
	s2.Add(42)

	want := "{9}"
	s1.IntersectWith(&s2)
	got := s1.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDifferenceWith(t *testing.T) {
	t.Parallel()

	var s1, s2 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	s2.Add(9)
	s2.Add(42)

	want := "{1 144}"
	s1.DifferenceWith(&s2)
	got := s1.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
