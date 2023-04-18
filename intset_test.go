package intset_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/weiwenchen2022/intset"

	"github.com/google/go-cmp/cmp"
)

func TestEmtpySet(t *testing.T) {
	t.Parallel()

	var s intset.IntSet[int]
	if n := s.Len(); 0 != n {
		t.Errorf("Empty set should contains 0 element, not %d", n)
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

	if n := s.Len(); 3 != n {
		t.Errorf("Len report %d elements, but it should be %d", n, 3)
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

	if 3 != s.Len() {
		t.Error("miss some elements")
	}

	s.Clear()
	if 0 != s.Len() {
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

func TestSymmetricDifference(t *testing.T) {
	t.Parallel()

	var s1, s2 intset.IntSet[int]
	s1.Add(1)
	s1.Add(144)
	s1.Add(9)

	s2.Add(9)
	s2.Add(42)

	want := "{1 42 144}"
	s1.SymmetricDifference(&s2)
	got := s1.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

// keySetInterface is the interface KeySet implements.
type keySetInterface interface {
	Add(int)
	Has(int) bool
	Len() int
}

// mapKeySet is an implementation of keySetInterface using a map.
type mapKeySet[K comparable] struct {
	keys map[K]bool
}

func (s *mapKeySet[K]) Add(x K) {
	if s.keys == nil {
		s.keys = make(map[K]bool)
	}

	s.keys[x] = true
}

func (s *mapKeySet[K]) Has(x K) bool {
	return s.keys[x]
}

func (s *mapKeySet[K]) Len() int {
	return len(s.keys)
}

// sliceKeySet is an implementation of keySetInterface using a slice.
type sliceKeySet[E comparable] struct {
	keys []E
}

func (s *sliceKeySet[E]) Add(x E) {
	if s.Has(x) {
		return
	}

	s.keys = append(s.keys, x)
}

func (s *sliceKeySet[E]) Has(x E) bool {
	for _, e := range s.keys {
		if x == e {
			return true
		}
	}

	return false
}

func (s sliceKeySet[E]) Len() int {
	return len(s.keys)
}

type bench struct {
	setup func(*testing.B, keySetInterface) *rand.Rand
	perG  func(b *testing.B, k keySetInterface, r *rand.Rand)
}

func benchKeySet(b *testing.B, bench bench) {
	for _, k := range [...]keySetInterface{
		&sliceKeySet[int]{},
		&mapKeySet[int]{},
		&intset.IntSet[int]{},
	} {
		name := fmt.Sprintf("%T", k)
		if index := strings.Index(name, "["); index > -1 {
			name = name[:index]
		}

		b.Run(name, func(b *testing.B) {
			k = reflect.New(reflect.TypeOf(k).Elem()).Interface().(keySetInterface)

			var r *rand.Rand
			if bench.setup != nil {
				r = bench.setup(b, k)
			}

			b.ResetTimer()
			bench.perG(b, k, r)
		})
	}
}

func BenchmarkAdd(b *testing.B) {
	benchKeySet(b, bench{
		setup: func(b *testing.B, ksi keySetInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			return r
		},

		perG: func(b *testing.B, k keySetInterface, r *rand.Rand) {
			const n = 100000

			for i := 0; i < b.N; i++ {
				k.Add(r.Intn(n))
			}
		},
	})
}

func BenchmarkHas(b *testing.B) {
	benchKeySet(b, bench{
		setup: func(b *testing.B, k keySetInterface) *rand.Rand {
			const n = 100000
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < b.N; i++ {
				k.Add(r.Intn(n))
			}

			return r
		},

		perG: func(b *testing.B, k keySetInterface, r *rand.Rand) {
			const n = 100000

			for i := 0; i < b.N; i++ {
				_ = k.Has(r.Intn(n))
			}
		},
	})
}

func BenchmarkLen(b *testing.B) {
	benchKeySet(b, bench{
		setup: func(b *testing.B, k keySetInterface) *rand.Rand {
			const n = 100000
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < n; i++ {
				k.Add(r.Intn(n))
			}

			return r
		},

		perG: func(b *testing.B, k keySetInterface, r *rand.Rand) {
			for i := 0; i < b.N; i++ {
				_ = k.Len()
			}
		},
	})
}
