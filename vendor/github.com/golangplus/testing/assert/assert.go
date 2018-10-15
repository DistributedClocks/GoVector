// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package assert provides some assertion functions for testing.

Return values: true if the assert holds, false otherwise.
*/
package assert

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Set this to false to avoid include file position in logs.
var IncludeFilePosition = true

func isTestFuncName(name string) bool {
	p := strings.LastIndex(name, ".")
	if p < 0 {
		return false
	}

	name = name[p+1:]
	return strings.HasPrefix(name, "Test")
}

// Default: skip == 0
func assertPos(skip int) string {
	if !IncludeFilePosition {
		return ""
	}
	res := ""
	for i := 0; i < 5; i++ {
		pc, file, line, ok := runtime.Caller(skip + 2)
		if !ok {
			return ""
		}
		res = fmt.Sprintf("%s:%d: ", path.Base(file), line) + res

		if isTestFuncName(runtime.FuncForPC(pc).Name()) {
			break
		}
		skip++
	}
	return "\n" + res
}

func valueMessage(v reflect.Value, incLen bool) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	var m string
	switch v.Kind() {
	case reflect.String:
		m = fmt.Sprintf("%q", v.Interface())
	default:
		m = fmt.Sprintf("%+v", v.Interface())
	}
	if incLen {
		m = fmt.Sprintf("(len=%d)%s", v.Len(), m)
	}
	return m
}

func needLen(act, exp reflect.Value) bool {
	if !act.IsValid() || !exp.IsValid() {
		return false
	}
	if act.Type() != exp.Type() {
		return false
	}
	switch act.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.String:
		return act.Len() != exp.Len()
	}
	return false
}

func needType(act, exp reflect.Value) bool {
	if !act.IsValid() || !exp.IsValid() {
		return false
	}
	return act.Type() != exp.Type()
}

func diffMessage(name string, act, exp reflect.Value) string {
	incLen := needLen(act, exp)
	actMsg := valueMessage(act, incLen)
	expMsg := valueMessage(exp, incLen)
	if needType(act, exp) {
		actMsg = fmt.Sprintf("%s(type=%v)", actMsg, act.Type())
		expMsg = fmt.Sprintf("%s(type=%v)", expMsg, exp.Type())
	}
	msg := fmt.Sprintf("%s is expected to be %s, but got %s", name, expMsg, actMsg)
	if len(msg) >= 80 {
		msg = fmt.Sprintf("%s is expected to be\n  %s\nbut got\n  %s", name, expMsg, actMsg)
	}
	return msg
}

func deepValueDiff(name string, act, exp reflect.Value) (message string, equal bool) {
	if !act.IsValid() || !exp.IsValid() {
		if act.IsValid() == exp.IsValid() {
			return "", true
		}
		return diffMessage(name, act, exp), false
	}
	if act.Type() != exp.Type() {
		return diffMessage(name, act, exp), false
	}
	switch act.Kind() {
	case reflect.Array:
		m, eq := []string(nil), true
		for i := 0; i < act.Len(); i++ {
			if mi, e := deepValueDiff(fmt.Sprintf("%s[%d]", name, i), act.Index(i), exp.Index(i)); !e {
				m = append(m, strings.Split(mi, "\n")...)
				eq = false
			}
		}
		if eq {
			return "", true
		}
		return fmt.Sprintf("some elements of %v are not expected:\n  %s", name, strings.Join(m, "\n  ")), false
	case reflect.Slice:
		if act.Len() != exp.Len() {
			return diffMessage(name, act, exp), false
		}
		if act.Pointer() == exp.Pointer() {
			return "", true
		}
		m, eq := []string(nil), true
		for i := 0; i < act.Len(); i++ {
			if mi, e := deepValueDiff(fmt.Sprintf("%s[%d]", name, i), act.Index(i), exp.Index(i)); !e {
				m = append(m, strings.Split(mi, "\n")...)
				eq = false
			}
		}
		if eq {
			return "", true
		}
		return fmt.Sprintf("some elements of %v are not expected:\n  %s", name, strings.Join(m, "\n  ")), false
	case reflect.Interface:
		if act.IsNil() || exp.IsNil() {
			if act.IsNil() == exp.IsNil() {
				return "", true
			}
			return diffMessage(name, act, exp), false
		}
		return deepValueDiff(name, act.Elem(), exp.Elem())
	case reflect.Ptr:
		if act.Pointer() == exp.Pointer() {
			return "", true
		}
		if act.IsNil() != exp.IsNil() {
			return diffMessage(name, act, exp), false
		}
		return deepValueDiff(name, act.Elem(), exp.Elem())
	case reflect.Struct:
		m, eq := []string(nil), true
		for i, n := 0, act.NumField(); i < n; i++ {
			if act.Type().Field(i).PkgPath != "" {
				// Contains unexported fields, use reflect.DeepEqual
				if reflect.DeepEqual(act.Interface(), exp.Interface()) {
					return "", true
				}
				eq = false
				// Try export difference of exported fields
				break
			}
		}
		for i, n := 0, act.NumField(); i < n; i++ {
			if act.Type().Field(i).PkgPath != "" {
				continue
			}
			if mi, e := deepValueDiff(fmt.Sprintf("%s.%s", name, act.Type().Field(i).Name), act.Field(i), exp.Field(i)); !e {
				m = append(m, strings.Split(mi, "\n")...)
				eq = false
			}
		}
		if eq {
			return "", true
		}
		return fmt.Sprintf("some fields of %v are not expected:\n  %s", name, strings.Join(m, "\n  ")), false
	case reflect.Map:
		if act.Pointer() == exp.Pointer() {
			return "", true
		}
		m, eq := []string(nil), true
		for _, k := range act.MapKeys() {
			actV := act.MapIndex(k)
			expV := exp.MapIndex(k)
			if !expV.IsValid() {
				continue
			}
			if mk, e := deepValueDiff(fmt.Sprintf("%s[%v]", name, valueMessage(k, false)), actV, expV); !e {
				eq = false
				m = append(m, strings.Split(mk, "\n")...)
			}
		}
		for _, k := range act.MapKeys() {
			actV := act.MapIndex(k)
			expV := exp.MapIndex(k)
			if expV.IsValid() {
				continue
			}
			eq = false
			m = append(m, fmt.Sprintf("extra %s -> %s", valueMessage(k, false), valueMessage(actV, false)))
		}
		for _, k := range exp.MapKeys() {
			actV := act.MapIndex(k)
			if actV.IsValid() {
				// Have checked in previous loop.
				continue
			}
			eq = false
			m = append(m, fmt.Sprintf("missing %s -> %s", valueMessage(k, false), valueMessage(exp.MapIndex(k), false)))
		}
		if eq {
			return "", true
		}
		return fmt.Sprintf("%v is unexpected:\n  %v", name, strings.Join(m, "\n  ")), false
	case reflect.Func:
		if act.IsNil() && exp.IsNil() {
			return "", true
		}
		// Can't do better than this:
		return diffMessage(name, act, exp), false
	default:
		if act.Interface() == exp.Interface() {
			return "", true
		}
		return diffMessage(name, act, exp), false
	}
}

func Equal(t testing.TB, name string, act, exp interface{}) bool {
	m, eq := deepValueDiff(name, reflect.ValueOf(act), reflect.ValueOf(exp))
	if eq {
		return true
	}
	t.Errorf("%s%s", assertPos(0), m)
	return false
}

// @param expToFunc could be a func with a single input value and a bool return, or a bool value directly.
func ValueShould(t testing.TB, name string, act interface{}, expToFunc interface{}, descIfFailed string) bool {
	expFunc := reflect.ValueOf(expToFunc)
	actValue := reflect.ValueOf(act)
	var succ bool
	if expFunc.Kind() == reflect.Bool {
		succ = expFunc.Bool()
	} else if expFunc.Kind() == reflect.Func {
		if expFunc.Type().NumIn() != 1 {
			t.Errorf("%sassert: expToFunc must have one parameter", assertPos(0))
			return false
		}
		if expFunc.Type().NumOut() != 1 {
			t.Errorf("%sassert: expToFunc must have one return value", assertPos(0))
			return false
		}
		if expFunc.Type().Out(0).Kind() != reflect.Bool {
			t.Errorf("%sassert: expToFunc must return a bool", assertPos(0))
			return false
		}
		succ = expFunc.Call([]reflect.Value{actValue})[0].Bool()
	} else {
		t.Errorf("%sassert: expToFunc must be a func or a bool", assertPos(0))
		return false
	}
	if !succ {
		t.Errorf("%s%s %s: %q(type %v)", assertPos(0), name, descIfFailed,
			fmt.Sprint(act), actValue.Type())
	}
	return succ
}

func NotEqual(t testing.TB, name string, act, exp interface{}) bool {
	if act == exp {
		t.Errorf("%s%s is not expected to be %q", assertPos(0), name, fmt.Sprint(exp))
		return false
	}
	return true
}

func False(t testing.TB, name string, act bool) bool {
	if act {
		t.Errorf("%s%s unexpectedly got true", assertPos(0), name)
	}
	return !act
}

func True(t testing.TB, name string, act bool) bool {
	if !act {
		t.Errorf("%s%s unexpectedly got false", assertPos(0), name)
	}
	return act
}

func Should(t testing.TB, vl bool, showIfFailed string) bool {
	if !vl {
		t.Errorf("%s%s", assertPos(0), showIfFailed)
	}
	return vl
}

func ShouldOrDie(t testing.TB, vl bool, showIfFailed string) {
	if !vl {
		t.Fatalf("%s%s", assertPos(0), showIfFailed)
	}
}

func sliceToStrings(a reflect.Value) []string {
	l := make([]string, a.Len())
	for i := 0; i < a.Len(); i++ {
		l[i] = fmt.Sprintf("%+v", a.Index(i).Interface())
	}
	return l
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func linesEqual(skip int, t testing.TB, name string, act, exp reflect.Value) bool {
	actS, expS := sliceToStrings(act), sliceToStrings(exp)
	if stringSliceEqual(actS, expS) {
		return true
	}

	title := fmt.Sprintf("%sUnexpected %s: ", assertPos(skip), name)
	if len(expS) == len(actS) {
		title = fmt.Sprintf("%sboth %d lines", title, len(expS))
	} else {
		title = fmt.Sprintf("%sexp %d, act %d lines", title, len(expS), len(actS))
	}
	t.Error(title)
	t.Log("  Difference(expected ---  actual +++)")

	_, expMat, actMat := match(len(expS), len(actS), func(expI, actI int) int {
		if expS[expI] == actS[actI] {
			return 0
		}
		return 2
	}, func(int) int {
		return 1
	}, func(int) int {
		return 1
	})
	for i, j := 0, 0; i < len(expS) || j < len(actS); {
		switch {
		case j >= len(actS) || i < len(expS) && expMat[i] < 0:
			t.Logf("    --- %3d: %q", i+1, expS[i])
			i++
		case i >= len(expS) || j < len(actS) && actMat[j] < 0:
			t.Logf("    +++ %3d: %q", j+1, actS[j])
			j++
		default:
			if expS[i] != actS[j] {
				t.Logf("    --- %3d: %q", i+1, expS[i])
				t.Logf("    +++ %3d: %q", j+1, actS[j])
			} // else
			i++
			j++
		}
	}
	return false
}

// StringEqual compares the string representation of the values.
// If act and exp are both slices, they were matched by elements and the results are
// presented in a diff style (if not totally equal).
func StringEqual(t testing.TB, name string, act, exp interface{}) bool {
	actV, expV := reflect.ValueOf(act), reflect.ValueOf(exp)
	if actV.Kind() == reflect.Slice && expV.Kind() == reflect.Slice {
		return linesEqual(1, t, name, actV, expV)
	}
	actS, expS := fmt.Sprintf("%+v", act), fmt.Sprintf("%+v", exp)
	if actS == expS {
		return true
	}
	if strings.ContainsRune(actS, '\n') || strings.ContainsRune(expS, '\n') {
		return linesEqual(1, t, name,
			reflect.ValueOf(strings.Split(actS, "\n")),
			reflect.ValueOf(strings.Split(expS, "\n")))
	}
	msg := fmt.Sprintf("%s%s is expected to be %q, but got %q", assertPos(0), name,
		fmt.Sprint(exp), fmt.Sprint(act))
	if len(msg) >= 80 {
		msg = fmt.Sprintf("%s%s is expected to be\n  %q\nbut got\n  %q", assertPos(0), name,
			fmt.Sprint(exp), fmt.Sprint(act))
	}
	t.Error(msg)
	return false
}

func NoError(t testing.TB, err error) bool {
	if err != nil {
		t.Errorf("%s%v", assertPos(0), err)
		return false
	}
	return true
}

func NoErrorOrDie(t testing.TB, err error) {
	if err != nil {
		t.Fatalf("%s%v", assertPos(0), err)
	}
}

func Error(t testing.TB, err error) bool {
	if err == nil {
		t.Error("Expecting error but nil got!")
		return false
	}
	return true
}

func Panic(t testing.TB, name string, f func()) bool {
	if !func() (res bool) {
		defer func() {
			res = recover() != nil
		}()

		f()
		return
	}() {
		t.Errorf("%s%s does not panic as expected.", assertPos(0), name)
		return false
	}
	return true
}
