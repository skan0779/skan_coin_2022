package utilities

import (
	"fmt"
)

func ExampleHash() {
	a := struct{ Test string }{Test: "hi"}
	b := Hash(a)
	fmt.Println(b)
	// Output: ~~
}
