package main

// Solution! c = 25734, a = 6

import (
	"fmt"
//	"os"
)

var a = uint32(4)
var b = uint32(1)
var c = uint32(0)

var cache = [32768*8]uint32{0}

func recurse() {
	key := (a * 32768) + b
	if cache[key] > 0 {
		a = cache[key]
//		fmt.Printf("%v\n", cache)
//		os.Exit(1)
		return
	}

	if a > 0 {
		if b > 0 {
			tmp := a
			b--
			recurse()
			b = a
			a = tmp
			a--
			recurse()
		} else {
			a--
			b = c
			recurse()
		}
	} else {
		a = (b+1) % 32768
	}

	cache[key] = a
}

func main() {
	for a != 6 && c < 32768 {
		c++
		a = 4
		b = 1
		if c % 100 == 0 { fmt.Printf("Trying: c=%d\n", c) }
		recurse()
//		fmt.Printf("After: %d, %d, %d\n", a, b, c)
		cache = [32768*8]uint32{}
	}

	fmt.Printf("Solution! c = %d, a = %d\n", c, a)
}
