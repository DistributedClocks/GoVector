package fmtp

import (
	"bytes"
	"fmt"
)

func ExamplePrintfln() {
	Printfln("Hello: %d", 1234)
	Printfln("World: %s", "!")
	// OUTPUT:
	// Hello: 1234
	// World: !
}

func ExampleFprintfln() {
	var b bytes.Buffer

	Fprintfln(&b, "Hello: %d", 1234)
	Fprintfln(&b, "World: %s", "!")

	fmt.Print(b.String())
	// OUTPUT:
	// Hello: 1234
	// World: !
}
