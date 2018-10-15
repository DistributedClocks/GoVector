package assert

import (
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {
	test := func(a, b string, d int, matA []int, matB []int) {
		ra, rb := []rune(a), []rune(b)
		actD, actMatA, actMatB := match(len(ra), len(rb),
			func(iA, iB int) int {
				if ra[iA] == rb[iB] {
					return 0
				}

				return 100
			},
			func(iA int) int {
				return 100 + iA
			},
			func(iB int) int {
				return 110 + iB
			})
		Equal(t, fmt.Sprintf("Edit-distance between %s and %s", a, b), actD, d)
		StringEqual(t, fmt.Sprintf("matA for matchting between %s and %s", a, b), actMatA, matA)
		StringEqual(t, fmt.Sprintf("matB for matchting between %s and %s", a, b), actMatB, matB)
	}

	test("abcd", "bcde", 213, []int{-1, 0, 1, 2}, []int{1, 2, 3, -1})
	test("abcde", "", 510, []int{-1, -1, -1, -1, -1}, []int{})
	test("", "abcde", 560, []int{}, []int{-1, -1, -1, -1, -1})
	test("", "", 0, []int{}, []int{})
	test("abcde", "abcde", 0, []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4})
	test("abcde", "dabce", 213, []int{1, 2, 3, -1, 4}, []int{-1, 0, 1, 2, 4})
	test("abcde", "abfde", 100, []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4})
}
