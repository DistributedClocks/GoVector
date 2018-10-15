// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testingp

import (
	"bytes"
	"testing"
)

func TestWriterTB(t *testing.T) {
	var b bytes.Buffer
	wtb := &WriterTB{
		Writer: &b,
		Suffix: "T",
	}

	wtb.Log("hello", "world")
	if wtb.Failed() {
		t.Error("wtb.Failed() should be false")
	}
	wtb.Logf("Hello: %s", "World")
	if wtb.Failed() {
		t.Error("wtb.Failed() should be false")
	}

	wtb.Error("error", "log")
	if !wtb.Failed() {
		t.Error("wtb.Failed() should be true")
	}
	// reset the flag
	wtb.failed = false

	wtb.Errorf("Error: %s", "Log")
	if !wtb.Failed() {
		t.Error("wtb.Failed() should be true")
	}
	// reset the flag
	wtb.failed = false

	wtb.Fail()
	if !wtb.Failed() {
		t.Error("wtb.Failed() should be true")
	}
	// reset the flag
	wtb.failed = false

	err := func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.FailNow()
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != FailedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	err = func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.Fatal("fatal", "msg")
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != FailedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	err = func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.Fatalf("Fatal: %s", "msg")
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != FailedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	err = func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.SkipNow()
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != SkippedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	err = func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.Skip("skip", "this")
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != SkippedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	err = func() (err interface{}) {
		defer func() {
			err = recover()
		}()
		wtb.Skipf("Skip: %s", "Now")
		return
	}()
	if err == nil {
		t.Error("Should panic")
	} else if err != SkippedErr {
		t.Errorf("Expected FailedErr but got: %v", err)
	}

	actLogs := b.String()
	expLogs := `T: hello world
T: Hello: World
T: error log
T: Error: Log
T: fatal msg
T: Fatal: msg
T: skip this
T: Skip: Now
`
	if actLogs != expLogs {
		t.Errorf("Expected logs:\n %v, but got: \n%v", expLogs, actLogs)
	}
}
