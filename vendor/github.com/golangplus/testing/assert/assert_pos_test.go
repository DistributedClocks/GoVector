package assert

import (
	"fmt"
	"testing"

	"github.com/golangplus/bytes"
	"github.com/golangplus/testing"
)

func TestFilePosition(t *testing.T) {
	var b bytesp.Slice
	bt := &testingp.WriterTB{Writer: &b}

	Equal(bt, "v", 1, 2)
	line := 15 // the line number of the last line
	Equal(t, "log", string(b), fmt.Sprintf("\nassert_pos_test.go:%d: v is expected to be 2, but got 1\n", line))

	b.Reset()
	Panic(bt, "nonpanic", func() {})
	line = 20 // the line number of the last line
	Equal(t, "log", string(b), fmt.Sprintf("\nassert_pos_test.go:%d: nonpanic does not panic as expected.\n", line))

	func(outLine int) {
		b.Reset()
		Equal(bt, "v", 1, 2)
		line := 26 // the line number of the last line
		Equal(t, "log", string(b), fmt.Sprintf("\nassert_pos_test.go:%d: assert_pos_test.go:%d: v is expected to be 2, but got 1\n", outLine, line))
	}(29) // the number in parentheses is the line number of current line

	b.Reset()
	StringEqual(bt, "s", []int{1}, []int{2})
	line = 32 // the line number of the last line
	StringEqual(t, "log", string(b), fmt.Sprintf(`
assert_pos_test.go:%d: Unexpected s: both 1 lines
  Difference(expected ---  actual +++)
    ---   1: "2"
    +++   1: "1"
`, line))
}
