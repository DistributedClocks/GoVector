// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package testingp is a plus to standard "testing" package.
*/
package testingp

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/golangplus/fmt"
)

var (
	// error thrown when failed
	FailedErr = errors.New("FailedErr")
	// error thrown when skipped
	SkippedErr = errors.New("SkippedErr")
)

// *WriterTB implements the testing.TB interface.
// An io.Writer can be specified as the destination of logging.
// Skipped and failed status are also stored.
// FailedErr is thrown when FailNow is called and SkippedErr for SkipNow.
//
// This type is especially useful for writing testcases of tools for testing.
type WriterTB struct {
	// this embeding implements private methods of testing.TB
	testing.TB

	// The destination of the logs
	io.Writer
	// The suffix for each log
	Suffix string

	failed  bool
	skipped bool
}

var _ testing.TB = (*WriterTB)(nil)

func (wtb *WriterTB) Error(args ...interface{}) {
	wtb.Log(args...)
	wtb.Fail()
}

func (wtb *WriterTB) Errorf(format string, args ...interface{}) {
	wtb.Logf(format, args...)
	wtb.Fail()
}

func (wtb *WriterTB) Fail() {
	wtb.failed = true
}

func (wtb *WriterTB) FailNow() {
	wtb.Fail()
	panic(FailedErr)
}

func (wtb *WriterTB) Failed() bool {
	return wtb.failed
}

func (wtb *WriterTB) Fatal(args ...interface{}) {
	wtb.Log(args...)
	wtb.FailNow()
}

func (wtb *WriterTB) Fatalf(format string, args ...interface{}) {
	wtb.Logf(format, args...)
	wtb.FailNow()
}

func (wtb *WriterTB) Log(args ...interface{}) {
	if wtb.Suffix != "" {
		io.WriteString(wtb, wtb.Suffix)
		wtb.Write([]byte(": "))
	}
	fmt.Fprintln(wtb.Writer, args...)
}

func (wtb *WriterTB) Logf(format string, args ...interface{}) {
	if wtb.Suffix != "" {
		io.WriteString(wtb, wtb.Suffix)
		wtb.Write([]byte(": "))
	}
	fmtp.Fprintfln(wtb.Writer, format, args...)
}

func (wtb *WriterTB) Skip(args ...interface{}) {
	wtb.Log(args...)
	wtb.SkipNow()
}

func (wtb *WriterTB) SkipNow() {
	wtb.skipped = true
	panic(SkippedErr)
}

func (wtb *WriterTB) Skipf(format string, args ...interface{}) {
	wtb.Logf(format, args...)
	wtb.SkipNow()
}

func (wtb *WriterTB) Skipped() bool {
	return wtb.skipped
}
