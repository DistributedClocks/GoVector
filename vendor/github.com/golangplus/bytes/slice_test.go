// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytesp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/golangplus/testing/assert"
)

func ExampleSlice() {
	var b Slice
	b.WriteByte(65)
	b.WriteString("bc")

	fmt.Println(b)
	fmt.Println(string(b))
	// OUTPUT:
	// [65 98 99]
	// Abc
}

func TestSlice(t *testing.T) {
	var bs Slice
	assert.Equal(t, "len(bs)", len(bs), 0)
	assert.StringEqual(t, "bs", bs, "[]")

	n, err := bs.Skip(1)
	assert.Equal(t, "err", err, io.EOF)
	assert.Equal(t, "n", n, int64(0))

	c, err := bs.ReadByte()
	assert.Equal(t, "err", err, io.EOF)

	bs.Write([]byte{1})
	c, err = bs.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, "c", c, byte(1))
	assert.Equal(t, "len(bs)", len(bs), 0)

	bs.Write([]byte{1, 2, 3})
	n, err = bs.Skip(4)
	assert.NoError(t, err)
	assert.Equal(t, "n", n, int64(3))

	bs.Write([]byte{1, 2, 3})
	assert.Equal(t, "len(bs)", len(bs), 3)
	assert.StringEqual(t, "bs", bs, "[1 2 3]")

	cnt, err := bs.Read(nil)
	assert.NoError(t, err)
	assert.Equal(t, "cnt", cnt, 0)
	assert.Equal(t, "len(bs)", len(bs), 3)
	assert.Equal(t, "bs", bs, Slice([]byte{1, 2, 3}))

	n, err = bs.Skip(0)
	assert.NoError(t, err)
	assert.Equal(t, "n", n, int64(0))
	assert.Equal(t, "len(bs)", len(bs), 3)
	assert.Equal(t, "bs", bs, Slice([]byte{1, 2, 3}))

	p := make([]byte, 2)
	bs.Read(p)
	assert.Equal(t, "len(bs)", len(bs), 1)
	assert.StringEqual(t, "bs", bs, "[3]")
	assert.StringEqual(t, "p", p, "[1 2]")

	bs.Read(make([]byte, 1))
	assert.Equal(t, "len(bs)", len(bs), 0)
	assert.StringEqual(t, "bs", bs, "[]")

	bs.Write([]byte{4, 5})
	assert.Equal(t, "len(bs)", len(bs), 2)
	assert.StringEqual(t, "bs", bs, "[4 5]")

	bs.WriteByte(6)

	c, err = bs.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, "c", c, byte(4))
	assert.StringEqual(t, "bs", bs, "[5 6]")

	bs.WriteRune('A')
	assert.Equal(t, "len(bs)", len(bs), 3)
	assert.StringEqual(t, "bs", bs, "[5 6 65]")
	bs.WriteRune('中')
	assert.Equal(t, "len(bs)", len(bs), 6)
	assert.StringEqual(t, "bs", bs, "[5 6 65 228 184 173]")

	bs.WriteString("世界")
	assert.Equal(t, "len(bs)", len(bs), 12)
	assert.StringEqual(t, "bs", bs, "[5 6 65 228 184 173 228 184 150 231 149 140]")

	bs.Skip(1)
	assert.StringEqual(t, "bs", bs, "[6 65 228 184 173 228 184 150 231 149 140]")

	bs.Close()

	bs.Reset()
	fmt.Fprint(&bs, "ABC")
	assert.StringEqual(t, "bs", bs, "[65 66 67]")

	data := make([]byte, 35*1024)
	io.ReadFull(rand.Reader, data)
	bs = nil
	n, err = bs.ReadFrom(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, "n", n, int64(len(data)))
	assert.Equal(t, "bs == data", bytes.Equal(bs, data), true)

	bs = nil
	n, err = Slice(data).WriteTo(&bs)
	assert.NoError(t, err)
	assert.Equal(t, "n", n, int64(len(data)))
	assert.Equal(t, "bs == data", bytes.Equal(bs, data), true)

	bs = []byte("A中文")
	r, size, err := bs.ReadRune()
	assert.NoError(t, err)
	assert.Equal(t, "size", size, 1)
	assert.Equal(t, "r", r, 'A')
	r, size, err = bs.ReadRune()
	assert.NoError(t, err)
	assert.Equal(t, "size", size, len([]byte("中")))
	assert.Equal(t, "r", r, '中')
	r, size, err = bs.ReadRune()
	assert.NoError(t, err)
	assert.Equal(t, "size", size, len([]byte("文")))
	assert.Equal(t, "r", r, '文')
}

func TestSlice_Bug_Read(t *testing.T) {
	var s Slice
	n, err := s.Read(make([]byte, 1))
	t.Logf("n: %d, err: %v", n, err)
	assert.Equal(t, "n", 0, 0)
	assert.Equal(t, "err", err, io.EOF)
}

type readerReturningEOFLastRead int

func (r *readerReturningEOFLastRead) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n = int(*r)
	if len(p) < n {
		n = len(p)
	}
	for i := 0; i < n; i++ {
		p[i] = 1
	}
	*r -= readerReturningEOFLastRead(n)
	if *r == 0 {
		return n, io.EOF
	}
	return n, nil
}

func TestSlice_ReadFrom(t *testing.T) {
	src := make([]byte, 63*1024)
	for i := range src {
		src[i] = byte(i)
	}
	b := bytes.NewBuffer(src)
	var s Slice
	n, err := s.ReadFrom(b)
	assert.Equal(t, "n", n, int64(len(src)))
	assert.NoError(t, err)
	assert.Equal(t, "s", s, Slice(src))
}

type readerReturningEOFAfterLastRead int

func (r *readerReturningEOFAfterLastRead) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n = int(*r)
	if len(p) < n {
		n = len(p)
	}
	for i := 0; i < n; i++ {
		p[i] = 2
	}
	*r -= readerReturningEOFAfterLastRead(n)
	if n == 0 {
		return n, io.EOF
	}
	return n, nil
}

func TestSlice_ReadFromEOF(t *testing.T) {
	r1 := readerReturningEOFLastRead(1)
	var s Slice
	n, err := s.ReadFrom(&r1)
	assert.Equal(t, "n", n, int64(1))
	assert.NoError(t, err)
	assert.Equal(t, "s", s, Slice([]byte{1}))

	r2 := readerReturningEOFAfterLastRead(1)
	s = nil
	n, err = s.ReadFrom(&r2)
	assert.Equal(t, "n", n, int64(1))
	assert.NoError(t, err)
	assert.Equal(t, "s", s, Slice([]byte{2}))
}

type readerErrUnexpectedEOF int

func (r *readerErrUnexpectedEOF) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n = int(*r)
	if len(p) < n {
		n = len(p)
	}
	for i := 0; i < n; i++ {
		p[i] = 3
	}
	*r -= readerErrUnexpectedEOF(n)
	if *r == 0 {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}

func TestSlice_ReadFromFailed(t *testing.T) {
	r0 := readerErrUnexpectedEOF(0)
	var s Slice
	n, err := s.ReadFrom(&r0)
	assert.Equal(t, "n", n, int64(0))
	assert.Equal(t, "err", err, io.ErrUnexpectedEOF)
	assert.Equal(t, "s", s, Slice(nil))

	r1 := readerErrUnexpectedEOF(1)
	s = nil
	n, err = s.ReadFrom(&r1)
	assert.Equal(t, "n", n, int64(1))
	assert.Equal(t, "err", err, io.ErrUnexpectedEOF)
	assert.Equal(t, "s", s, Slice([]byte{3}))
}

func TestSlice_Bug_ReadRune(t *testing.T) {
	s := Slice{65, 0xff, 66}
	r, sz, err := s.ReadRune()
	assert.Equal(t, "r", r, 'A')
	assert.Equal(t, "sz", sz, 1)
	assert.Equal(t, "err", err, nil)
	r, sz, err = s.ReadRune()
	assert.Equal(t, "r", r, utf8.RuneError)
	assert.Equal(t, "sz", sz, 1)
	assert.Equal(t, "err", err, nil)

	r, sz, err = s.ReadRune()
	assert.Equal(t, "r", r, 'B')
	assert.Equal(t, "sz", sz, 1)
	assert.Equal(t, "err", err, nil)
}

func TestSlice_WriteItoa(t *testing.T) {
	var s Slice
	s.WriteItoa(1234, 10)
	s.WriteItoa(255, 16)

	assert.Equal(t, "s", string(s), "1234ff")
}

func TestSLice_NewPSlice(t *testing.T) {
	s := NewPSlice([]byte{1, 2, 3})
	assert.Equal(t, "*s", []byte(*s), []byte{1, 2, 3})
}

func TestSlice_ReadRune_UnfullRune(t *testing.T) {
	var s Slice = []byte("\xF0\xA4\xAD")
	_, _, err := s.ReadRune()
	assert.Equal(t, "err", err, io.ErrUnexpectedEOF)
}

func TestSlice_WriteRune_InvalidRune(t *testing.T) {
	var s Slice
	_, err := s.WriteRune(utf8.MaxRune + 1)
	assert.Equal(t, "err", err, ErrInvalidRune)
}

func BenchmarkSliceRead1k(b *testing.B) {
	var data [1000]byte
	for i := 0; i < b.N; i++ {
		b := Slice(data[:])
		for {
			if _, err := b.ReadByte(); err != nil {
				break
			}
		}
	}
}

func BenchmarkBytesBufferRead1k(b *testing.B) {
	var data [1000]byte
	for i := 0; i < b.N; i++ {
		b := bytes.NewBuffer(data[:])
		for {
			if _, err := b.ReadByte(); err != nil {
				break
			}
		}
	}
}

func BenchmarkBytesReader1k(b *testing.B) {
	var data [1000]byte
	for i := 0; i < b.N; i++ {
		b := bytes.NewReader(data[:])
		for {
			if _, err := b.ReadByte(); err != nil {
				break
			}
		}
	}
}

func BenchmarkSliceRead10(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		b := Slice(data[:])
		for {
			if _, err := b.ReadByte(); err != nil {
				break
			}
		}
	}
}

func BenchmarkBytesBufferRead10(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		b := bytes.NewBuffer(data[:])
		for {
			if _, err := b.ReadByte(); err != nil {
				break
			}
		}
	}
}

func BenchmarkReader10(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data[:])
		for {
			if _, err := r.ReadByte(); err != nil {
				if err != io.EOF {
					b.Fatalf("r.ReadByte failed: %v", err)
				}
				break
			}
		}
	}
}

func BenchmarkSliceWrite10(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		b := Slice(data[:0])
		for range data {
			b.WriteByte(0)
		}
	}
}

func BenchmarkBytesBufferWrite10_New(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		b := bytes.NewBuffer(data[:0])
		for range data {
			b.WriteByte(0)
		}
	}
}

func BenchmarkBytesBufferWrite10_Def(b *testing.B) {
	var data [10]byte
	for i := 0; i < b.N; i++ {
		var b bytes.Buffer
		for range data {
			b.WriteByte(0)
		}
	}
}

func BenchmarkSliceWrite1k(b *testing.B) {
	var data [1000]byte
	for i := 0; i < b.N; i++ {
		b := Slice(data[:0])
		for range data {
			b.WriteByte(0)
		}
	}
}

func BenchmarkBytesBufferWrite1k(b *testing.B) {
	var data [1000]byte
	for i := 0; i < b.N; i++ {
		b := bytes.NewBuffer(data[:0])
		for range data {
			b.WriteByte(0)
		}
	}
}
