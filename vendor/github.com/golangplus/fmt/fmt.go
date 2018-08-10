package fmtp

import (
	"fmt"
	"io"
	"os"
)

// Printfln is similar to fmt.Printf but a newline is appended.
func Printfln(format string, a ...interface{}) (n int, err error) {
	return Fprintfln(os.Stdout, format, a...)
}

// Fprintfln is similar to fmt.Fprintf but a newline is appended.
func Fprintfln(w io.Writer, format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format+"\n", a...)
}

// Eprint is similar to fmt.Print but output to os.Stderr
func Eprint(a ...interface{}) (n int, err error) {
	return fmt.Fprint(os.Stderr, a...)
}

// Eprintf is similar to fmt.Printf but output to os.Stderr
func Eprintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}

// Eprintln is similar to fmt.Println but output to os.Stderr
func Eprintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stderr, a...)
}

// Eprintfln is similar to Printfln but output to os.Stderr
func Eprintfln(format string, a ...interface{}) (n int, err error) {
	return Fprintfln(os.Stderr, format, a...)
}
