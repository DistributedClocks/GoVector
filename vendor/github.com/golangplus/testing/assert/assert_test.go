// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package assert

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/golangplus/bytes"
	"github.com/golangplus/testing"
)

func TestSuccess(t *testing.T) {
	True(t, "return value", Equal(t, "v", 1, 1))
	True(t, "return true", Equal(t, "slice", []int{1}, []int{1}))
	True(t, "return true", Equal(t, "map", map[int]int{2: 1}, map[int]int{2: 1}))
	True(t, "return value", ValueShould(t, "s", "abc", func(s string) bool {
		return s == "abc"
	}, "is not abc"))
	True(t, "return value", ValueShould(t, "s", "abc", true, "is not abc"))
	True(t, "return value", NotEqual(t, "v", 1, 4))
	True(t, "return value", True(t, "bool", true))
	True(t, "return value", Should(t, true, "failed"))
	True(t, "return value", False(t, "bool", false))
	True(t, "return value", StringEqual(t, "string", 1, "1"))
	True(t, "return value", NoError(t, nil))
	True(t, "return value", Error(t, errors.New("fail")))
}

func ExampleEqual() {
	// The following two lines are for test/example of assert package itself. Use
	// *testing.T as t in normal testing instead.
	IncludeFilePosition = false
	defer func() { IncludeFilePosition = true }()
	t := &testingp.WriterTB{Writer: os.Stdout}

	Equal(t, "v", 1, 2)
	Equal(t, "v", 1, "1")
	Equal(t, "m", map[string]int{"Extra": 2, "Modified": 4}, map[string]int{"Missing": 1, "Modified": 5})

	// OUTPUT:
	// v is expected to be 2, but got 1
	// v is expected to be "1"(type=string), but got 1(type=int)
	// m is unexpected:
	//   m["Modified"] is expected to be 5, but got 4
	//   extra "Extra" -> 2
	//   missing "Missing" -> 1
}

func ExampleValueShould() {
	// The following two lines are for test/example of assert package itself. Use
	// *testing.T as t in normal testing instead.
	IncludeFilePosition = false
	defer func() { IncludeFilePosition = true }()
	t := &testingp.WriterTB{Writer: os.Stdout}

	ValueShould(t, "s", "\xff\xfe\xfd", utf8.ValidString, "is not valid UTF8")
	ValueShould(t, "s", "abcd", len("abcd") <= 3, "has more than 3 bytes")

	// OUTPUT:
	// s is not valid UTF8: "\xff\xfe\xfd"(type string)
	// s has more than 3 bytes: "abcd"(type string)
}

func ExampleStringEqual() {
	// The following two lines are for test/example of assert package itself. Use
	// *testing.T as t in normal testing instead.
	IncludeFilePosition = false
	defer func() { IncludeFilePosition = true }()
	t := &testingp.WriterTB{Writer: os.Stdout}

	StringEqual(t, "s", []int{2, 3}, []string{"1", "2"})
	StringEqual(t, "s", `
Extra
Modified act`, `
Modified exp
Missing`)

	// OUTPUT:
	// Unexpected s: both 2 lines
	//   Difference(expected ---  actual +++)
	//     ---   1: "1"
	//     +++   2: "3"
	// Unexpected s: both 3 lines
	//   Difference(expected ---  actual +++)
	//     ---   2: "Modified exp"
	//     ---   3: "Missing"
	//     +++   2: "Extra"
	//     +++   3: "Modified act"
}

func TestFailures(t *testing.T) {
	IncludeFilePosition = false
	defer func() { IncludeFilePosition = true }()
	var b bytesp.Slice
	bt := &testingp.WriterTB{Writer: &b}

	Equal(bt, "v", 1, "2")
	NotEqual(bt, "v", 1, 1)
	True(bt, "v", false)
	Should(bt, false, "Should failed")
	Panic(t, "ShouldOrDie", func() {
		ShouldOrDie(bt, false, "ShouldOrDie failed")
	})
	StringEqual(bt, "s", 1, "2")
	False(bt, "v", true)
	Panic(bt, "nonpanic", func() {})
	Error(bt, nil)
	NoError(bt, errors.New("failed"))
	Panic(t, "NoErrorOrDie", func() {
		NoErrorOrDie(bt, errors.New("failed"))
	})

	StringEqual(t, "output", "\n"+string(b), `
v is expected to be "2"(type=string), but got 1(type=int)
v is not expected to be "1"
v unexpectedly got false
Should failed
ShouldOrDie failed
s is expected to be "2", but got "1"
v unexpectedly got true
nonpanic does not panic as expected.
Expecting error but nil got!
failed
failed
`)
}

func TestPanic(t *testing.T) {
	True(t, "return value", Panic(t, "panic", func() {
		panic("error")
	}))
}

func TestDeepValueDiff_Equal(t *testing.T) {
	type S struct {
		V int
	}
	type F func()

	shouldEqual := func(a, b interface{}) {
		m, eq := deepValueDiff("v", reflect.ValueOf(a), reflect.ValueOf(b))
		True(t, "eq", eq)
		ValueShould(t, "m", m, m == "", "is not empty")
	}
	selfEqual := func(a interface{}) {
		shouldEqual(a, a)
	}
	shouldEqual(nil, nil)
	shouldEqual(1, 1)
	shouldEqual([...]int{1}, [...]int{1})
	shouldEqual([]int{1}, []int{1})
	shouldEqual(map[string]int{"A": 1}, map[string]int{"A": 1})
	shouldEqual(F(nil), F(nil))
	shouldEqual(S{V: 1}, S{V: 1})
	var a, b interface{}
	shouldEqual(&a, &b)
	a, b = 123, 123
	shouldEqual(&a, &b)

	selfEqual(&S{V: 1})
	selfEqual(S{V: 1})
	selfEqual([]int{1})
	selfEqual(map[string]int{"A": 1})
	var c interface{}
	selfEqual(&c)
	selfEqual(time.Now())
}

func TestDeepValueDiff_NotEqual(t *testing.T) {
	type S struct {
		V int
	}
	shouldNotEqual := func(a, b interface{}) {
		m, eq := deepValueDiff("variablename", reflect.ValueOf(a), reflect.ValueOf(b))
		False(t, "eq", eq)
		ValueShould(t, "m", m, strings.Contains(m, "variablename"), "does not contain variablename")
	}
	shouldNotEqual(nil, "word")
	shouldNotEqual([...]int{1}, [...]int{1, 2})
	shouldNotEqual([...]int{1}, [...]int{2})
	shouldNotEqual(int32(1), int64(1))
	shouldNotEqual([]int{1}, []int{1, 2})
	shouldNotEqual([]int{1}, []int{2})
	shouldNotEqual(S{V: 1}, S{V: 2})
	shouldNotEqual(&S{V: 1}, &S{V: 2})
	shouldNotEqual(map[string]int{"A": 1}, map[string]int{})
	shouldNotEqual(map[string]int{"A": 1}, map[string]int{"A": 2})
	shouldNotEqual(map[string]int{}, map[string]int{"B": 2})
	shouldNotEqual(func() {}, func() {})
	shouldNotEqual(time.Now(), time.Now().Add(time.Hour))
	var a interface{}
	var b interface{} = 123
	var c interface{} = 456
	shouldNotEqual(&a, &b)
	shouldNotEqual(&b, &c)
}

func TestDeepValueDiff_Message(t *testing.T) {
	type S struct {
		V int
		A []string
	}
	shouldNotEqualWithMessage := func(act, exp interface{}, msg string) {
		m, eq := deepValueDiff("v", reflect.ValueOf(act), reflect.ValueOf(exp))
		False(t, "eq", eq)
		StringEqual(t, "m", m, msg)
	}
	shouldNotEqualWithMessage([...]int{1, 3}, [...]int{1, 2}, `some elements of v are not expected:
  v[1] is expected to be 2, but got 3`)
	shouldNotEqualWithMessage([]int{1, 3}, []int{1, 2}, `some elements of v are not expected:
  v[1] is expected to be 2, but got 3`)
	shouldNotEqualWithMessage(S{V: 1}, S{V: 2}, `some fields of v are not expected:
  v.V is expected to be 2, but got 1`)
	shouldNotEqualWithMessage(S{A: []string{"1"}}, S{A: []string{"2"}}, `some fields of v are not expected:
  some elements of v.A are not expected:
    v.A[0] is expected to be "2", but got "1"`)
	shouldNotEqualWithMessage(map[string]S{"A": S{A: []string{"1"}}}, map[string]S{"A": S{A: []string{"2"}}}, `v is unexpected:
  some fields of v["A"] are not expected:
    some elements of v["A"].A are not expected:
      v["A"].A[0] is expected to be "2", but got "1"`)
	shouldNotEqualWithMessage(map[string]S{"A": S{A: []string{"1"}}}, map[string]S{"B": S{A: []string{"2"}}}, `v is unexpected:
  extra "A" -> {V:0 A:[1]}
  missing "B" -> {V:0 A:[2]}`)
	shouldNotEqualWithMessage((*S)(nil), &S{}, `v is expected to be &{V:0 A:[]}, but got <nil>`)
}
