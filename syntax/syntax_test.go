package syntax

import (
	"strconv"
	"testing"
)

func TestCommonChildren(t *testing.T) {
	for i, test := range []struct {
		nodes []*Node
		left  []*Node
		right []*Node
	}{
		{
			nodes: []*Node{
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"z"}),
					New(KindText, Text{"c"}),
				),
			},
		},
		{
			nodes: []*Node{
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"z"}),
					New(KindText, Text{"c"}),
				),
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"c"}),
				),
			},
			left: []*Node{
				New(KindText, Text{"a"}),
			},
			right: []*Node{
				New(KindText, Text{"c"}),
			},
		},
		{
			nodes: []*Node{
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"c"}),
					New(KindText, Text{"d"}),
				),
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"c"}),
					New(KindText, Text{"c"}),
					New(KindText, Text{"d"}),
				),
			},
			left: []*Node{
				New(KindText, Text{"a"}),
				New(KindText, Text{"b"}),
			},
			right: []*Node{
				New(KindText, Text{"c"}),
				New(KindText, Text{"d"}),
			},
		},
		{
			nodes: []*Node{
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"c"}),
				),
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"b"}),
					New(KindText, Text{"c"}),
				),
			},
			left: []*Node{
				New(KindText, Text{"a"}),
				New(KindText, Text{"b"}),
			},
			right: []*Node{
				New(KindText, Text{"c"}),
			},
		},
		{
			nodes: []*Node{
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"d"}),
				),
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"d"}),
				),
				New(KindNothing, nil,
					New(KindText, Text{"a"}),
					New(KindText, Text{"e"}),
				),
			},
			left: []*Node{
				New(KindText, Text{"a"}),
			},
			right: []*Node{},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			left, right := CommonChildren(test.nodes)
			if !Equal(left, test.left) {
				t.Errorf(
					"left, right := commonChildren(); left = %v; want %v",
					left, test.left,
				)
			}
			if !Equal(right, test.right) {
				t.Errorf(
					"left, right := commonChildren(); right = %v; want %v",
					right, test.right,
				)
			}
		})
	}
}
