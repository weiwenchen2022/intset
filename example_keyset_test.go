package intset_test

import (
	"fmt"

	"github.com/weiwenchen2022/intset"
)

type Key int

const (
	Copper Key = iota
	Jade
	Crystal

	MaxKey
)

var keys = [MaxKey]string{"copper", "jade", "crystal"}

// String implements the fmt.Stringer interface
func (k Key) String() string {
	if k < 0 || k >= MaxKey {
		return fmt.Sprintf("<unknown key: %d>", k)
	}

	return keys[k]
}

// KeySet is a set of keys in the game.
type KeySet = intset.IntSet[Key]

func Example_keySet() {
	var keys KeySet

	keys.Add(Copper)
	keys.Add(Jade)
	fmt.Println(&keys)

	keys.Remove(Copper)
	fmt.Println(keys.Has(Copper))
	fmt.Println(keys.Has(Jade))

	// Output:
	// {copper jade}
	// false
	// true
}
