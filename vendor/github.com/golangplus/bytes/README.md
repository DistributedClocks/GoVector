# bytes [![GoSearch](http://go-search.org/badge?id=github.com%2Fgolangplus%2Fbytes)](http://go-search.org/view?id=github.com%2Fgolangplus%2Fbytes)
Plus to the standard `bytes` package.

## Featured
```go
// ByteSlice is a wrapper type for []byte.
// Its pointer form, *ByteSlice, implements io.Reader, io.Writer, io.ByteReader,
// io.ByteWriter, io.Closer, io.ReaderFrom, io.WriterTo and io.RuneReader
// interfaces.
//
// Benchmark shows *ByteSlice is a better alternative for bytes.Buffer for writings and consumes less resource.
type ByteSlice []byte
```
([blog about ByteSlice](http://daviddengcn.blogspot.com/2015/07/a-light-and-fast-type-for-serializing.html))

## LICENSE
BSD license
