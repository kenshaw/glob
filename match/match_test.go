package match

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

var bench_separators = []rune{'.'}

const bench_pattern = "abcdefghijklmnopqrstuvwxyz0123456789"

func TestAppendMerge(t *testing.T) {
	for id, test := range []struct {
		segments [2][]int
		exp      []int
	}{
		{
			[2][]int{
				{0, 6, 7},
				{0, 1, 3},
			},
			[]int{0, 1, 3, 6, 7},
		},
		{
			[2][]int{
				{0, 1, 3, 6, 7},
				{0, 1, 10},
			},
			[]int{0, 1, 3, 6, 7, 10},
		},
	} {
		act := appendMerge(test.segments[0], test.segments[1])
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d merge sort segments unexpected:\nact: %v\nexp:%v", id, act, test.exp)
			continue
		}
	}
}

func getTable() []int {
	table := make([]int, utf8.MaxRune+1)
	for i := 0; i <= utf8.MaxRune; i++ {
		table[i] = utf8.RuneLen(rune(i))
	}
	return table
}

var table = getTable()

const runeToLen = 'q'
