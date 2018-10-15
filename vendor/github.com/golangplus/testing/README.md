# testing
Plus to the standard `testing` package

Godoc: [testingp](http://godoc.org/github.com/golangplus/testing) [assert](http://godoc.org/github.com/golangplus/testing/assert)

## Featured
```go
// *WriterTB implements the testing.TB interface.
// An io.Writer can be specified as the destination of logging.
// This type is especially useful for writing testcases of tools for testing.
type WriterTB struct {
	io.Writer
    Suffix string
	...
}
```

## LICENSE
BSD license
