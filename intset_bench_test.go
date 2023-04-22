package intset_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/weiwenchen2022/intset"
)

func benchmarkAddProbeIntSet(b *testing.B, size, spread int) {
	b.StopTimer()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Generate our insertions and probes beforehand (we don't want to benchmark
	// the r).
	add := make([]int, size)
	probe := make([]int, size*2)
	for i := range add {
		add[i] = r.Intn(spread)
	}
	for i := range probe {
		probe[i] = r.Intn(spread)
	}

	var s intset.IntSet[int]

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Clear()
		for _, x := range add {
			s.Add(x)
		}

		hits := 0
		for _, x := range probe {
			if s.Has(x) {
				hits++
			}
		}
		// Use the variable so it doesn't get optimized away.
		if hits > len(probe) {
			b.Fatalf("%d hits, only %d probes", hits, len(probe))
		}
	}
}

func BenchmarkAddProbeIntSet_2_10(b *testing.B) {
	benchmarkAddProbeIntSet(b, 2, 10)
}

func BenchmarkAddProbeIntSet_10_10(b *testing.B) {
	benchmarkAddProbeIntSet(b, 10, 10)
}

func BenchmarkAddProbeIntSet_10_1000(b *testing.B) {
	benchmarkAddProbeIntSet(b, 10, 1000)
}

func BenchmarkAddProbeIntSet_100_1000(b *testing.B) {
	benchmarkAddProbeIntSet(b, 100, 1000)
}

type bench struct {
	setup func(b *testing.B, s, t setInterface) *rand.Rand
	perG  func(b *testing.B, s, t setInterface, r *rand.Rand)
}

func benchSet(b *testing.B, bench bench) {
	for _, s := range [...]setInterface{&MapSet{}, &IntSet{}} {
		b.Run(fmt.Sprintf("%T", s), func(b *testing.B) {
			s = reflect.New(reflect.TypeOf(s).Elem()).Interface().(setInterface)
			t := reflect.New(reflect.TypeOf(s).Elem()).Interface().(setInterface)
			var r *rand.Rand
			if bench.setup != nil {
				r = bench.setup(b, s, t)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bench.perG(b, s, t, r)
			}
		})
	}
}

func BenchmarkAdd(b *testing.B) {
	const n = 100000

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			return r
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Add(r.Intn(n))
		},
	})
}

func BenchmarkRemove(b *testing.B) {
	const n = 100000

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(n))
			}

			return r
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			if s.Remove(r.Intn(n)) {
				s.Add(r.Intn(n))
			}
		},
	})
}

func BenchmarkHas(b *testing.B) {
	const n = 100000

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < n; i++ {
				s.Add(r.Intn(n))
			}

			return r
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Has(r.Intn(n))
		},
	})
}

func BenchmarkLen(b *testing.B) {
	const n = 100000

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < n; i++ {
				s.Add(r.Intn(n))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Len()
		},
	})
}

func BenchmarkIsEmpty(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(10000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.IsEmpty()
		},
	})
}

func BenchmarkAppendTo(b *testing.B) {
	var elems [1000]int

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(10000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.AppendTo(elems[:0])
		},
	})
}

func BenchmarkBitString(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(100000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.BitString()
		},
	})
}

func BenchmarkCopy(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(10000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Copy()
		},
	})
}

func BenchmarkLowerBound(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(100000))
			}

			return r
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.LowerBound(r.Intn(100000))
		},
	})
}

func BenchmarkMax(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(100000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Max()
		},
	})
}

func BenchmarkMin(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(100000))
			}

			return nil
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			s.Min()
		},
	})
}

func BenchmarkTakeMin(b *testing.B) {
	const n = 100000

	benchSet(b, bench{
		setup: func(b *testing.B, s, _ setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				s.Add(r.Intn(n))
			}

			return r
		},

		perG: func(b *testing.B, s, _ setInterface, r *rand.Rand) {
			var x int
			if s.TakeMin(&x) {
				s.Add(r.Intn(n))
			}
		},
	})
}

func BenchmarkEquals(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, t setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				x := r.Intn(100000)
				if i%2 == 0 {
					s.Add(x)
				} else {
					t.Add(x)
				}
			}

			return nil
		},

		perG: func(b *testing.B, s, t setInterface, r *rand.Rand) {
			s.Equals(t)
		},
	})
}

func BenchmarkUnionWith(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, t setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				x := r.Intn(100000)
				if i%2 == 0 {
					s.Add(x)
				} else {
					t.Add(x)
				}
			}

			return nil
		},

		perG: func(b *testing.B, s, t setInterface, r *rand.Rand) {
			sc := s.Copy().(setInterface)
			sc.UnionWith(t)
		},
	})
}

func BenchmarkIntersectWith(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, t setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				x := r.Intn(100000)
				if i%2 == 0 {
					s.Add(x)
				} else {
					t.Add(x)
				}
			}

			return nil
		},

		perG: func(b *testing.B, s, t setInterface, r *rand.Rand) {
			sc := s.Copy().(setInterface)
			sc.IntersectWith(t)
		},
	})
}

func BenchmarkIntersects(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, t setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				x := r.Intn(100000)
				if i%2 == 0 {
					s.Add(x)
				} else {
					t.Add(x)
				}
			}

			return nil
		},

		perG: func(b *testing.B, s, t setInterface, r *rand.Rand) {
			s.Intersects(t)
		},
	})
}

func BenchmarkDifferenceWith(b *testing.B) {
	benchSet(b, bench{
		setup: func(b *testing.B, s, t setInterface) *rand.Rand {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < 1000; i++ {
				x := r.Intn(100000)
				if i%2 == 0 {
					s.Add(x)
				} else {
					t.Add(x)
				}
			}

			return nil
		},

		perG: func(b *testing.B, s, t setInterface, r *rand.Rand) {
			sc := s.Copy().(setInterface)
			sc.DifferenceWith(t)
		},
	})
}
