# fmt
Plus to standard fmt package.

[Godoc](http://godoc.org/github.com/golangplus/fmt)

## Featured
```go
// Printfln is similar to fmt.Printf but a newline is appended.
func Printfln(format string, a ...interface{}) (n int, err error) {}

// Fprintfln is similar to fmt.Fprintf but a newline is appended.
func Fprintfln(w io.Writer, format string, a ...interface{}) (n int, err error) {}

// Eprint is similar to fmt.Print but output to os.Stderr
func Eprint(a ...interface{}) (n int, err error) {}
```

## LICENSE
BSD license
