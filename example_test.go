package intset_test

import (
	"fmt"

	"github.com/weiwenchen2022/intset"
)

func Example() {
	var s1, s2 intset.IntSet[int]

	s1.Add(1)
	s1.Add(144)
	s1.Add(9)
	fmt.Println(s1.String())

	s2.Add(9)
	s2.Add(42)
	fmt.Println(s2.String())

	s1.UnionWith(&s2)
	fmt.Println(s1.String())

	fmt.Println(s1.Has(9), s1.Has(123))

	// Output:
	// {1 9 144}
	// {9 42}
	// {1 9 42 144}
	// true false
}
