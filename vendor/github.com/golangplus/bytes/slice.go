// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytesp

import (
	"errors"
	"io"
	"strconv"
	"unicode/utf8"
)

// Slice is a wrapper type for []byte.
// Its pointer form, *Slice, implements io.Reader, io.Writer, io.ByteReader,
// io.ByteWriter, io.Closer, io.ReaderFrom, io.WriterTo and io.RuneReader
// interfaces.
//
// Benchmark shows *Slice is a better alternative for bytes.Buffer for writings and consumes less resource.
type Slice []byte

var (
	// Make sure *Slice implement the interfaces.
	_ io.Reader     = (*Slice)(nil)
	_ io.Writer     = (*Slice)(nil)
	_ io.ByteReader = (*Slice)(nil)
	_ io.ByteWriter = (*Slice)(nil)
	_ io.Closer     = (*Slice)(nil)
	_ io.ReaderFrom = (*Slice)(nil)
	_ io.WriterTo   = (*Slice)(nil)
	_ io.RuneReader = (*Slice)(nil)
)

// NewPSlice returns a *Slice with intialized contents.
func NewPSlice(bytes []byte) *Slice {
	return (*Slice)(&bytes)
}

// Read implements the io.Reader interface.
// After some bytes are read, the slice shrinks.
func (s *Slice) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if len(*s) == 0 {
		return 0, io.EOF
	}
	n = copy(p, *s)

	if n == len(*s) {
		*s = nil
	} else {
		*s = (*s)[n:]
	}

	return n, nil
}

// Reset sets the length of the slice to 0.
func (s *Slice) Reset() {
	*s = (*s)[:0]
}

// Skip skips n bytes.
func (s *Slice) Skip(n int64) (int64, error) {
	if n == 0 {
		return 0, nil
	}

	if len(*s) == 0 {
		return 0, io.EOF
	}
	if n >= int64(len(*s)) {
		n = int64(len(*s))
		*s = nil
	} else {
		*s = (*s)[int(n):]
	}
	return n, nil
}

// Write implements the io.Writer interface.
// Bytes are appended to the tail of the slice.
func (s *Slice) Write(p []byte) (n int, err error) {
	*s = append(*s, p...)
	return len(p), nil
}

// ReadByte implements the io.ByteReader interface.
func (s *Slice) ReadByte() (c byte, err error) {
	if len(*s) < 1 {
		return 0, io.EOF
	}

	c = (*s)[0]
	if len(*s) > 1 {
		*s = (*s)[1:]
	} else {
		*s = nil
	}
	return c, nil
}

// WriteByte implements the io.ByteWriter interface.
func (s *Slice) WriteByte(c byte) error {
	*s = append(*s, c)
	return nil
}

// Close implements the io.Closer interface.
// It does nothing.
func (s Slice) Close() error {
	return nil
}

// ReadFrom implements the io.ReaderFrom interface.
func (s *Slice) ReadFrom(r io.Reader) (n int64, err error) {
	const buf_SIZE = 32 * 1024
	buf := make([]byte, buf_SIZE)
	for {
		nRead, err := r.Read(buf)
		if nRead == 0 {
			if err != io.EOF {
				return n, err
			}
			break
		}
		n += int64(nRead)
		*s = append(*s, buf[:nRead]...)
		if err == io.EOF {
			break
		}

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// WriteTo implements the io.WriterTo interface.
func (s Slice) WriteTo(w io.Writer) (n int64, err error) {
	nWrite, err := w.Write(s)
	return int64(nWrite), err
}

// ReadRune implements the io.RuneReader interface.
func (s *Slice) ReadRune() (r rune, size int, err error) {
	if !utf8.FullRune(*s) {
		return utf8.RuneError, 0, io.ErrUnexpectedEOF
	}
	r, size = utf8.DecodeRune(*s)
	*s = (*s)[size:]

	return r, size, err
}

// error for a invalid rune
var ErrInvalidRune = errors.New("Slice: invalid rune")

var emptySlices = [...][]byte{
	nil,
	{0},
	{0, 0},
	{0, 0, 0},
	{0, 0, 0, 0},
}

// WriteRune writes a single Unicode code point, returning the number of bytes
// written and any error.
func (s *Slice) WriteRune(r rune) (size int, err error) {
	if r < utf8.RuneSelf {
		*s = append(*s, byte(r))
		return 1, nil
	}

	l := utf8.RuneLen(r)
	if l < 0 {
		return 0, ErrInvalidRune
	}

	*s = append(*s, emptySlices[l]...)
	utf8.EncodeRune((*s)[len(*s)-l:], r)
	return l, nil
}

// WriteString appends the contents of str to the slice, growing the slice as
// needed. The return value n is the length of str; err is always nil.
func (s *Slice) WriteString(str string) (size int, err error) {
	*s = append(*s, str...)
	return len(str), nil
}

// WriteItoa converts i into text of the specified base and write to s.
func (s *Slice) WriteItoa(i int64, base int) (size int, err error) {
	l := len(*s)
	*s = strconv.AppendInt([]byte(*s), i, base)
	return len(*s) - l, nil
}
