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
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"z"}),
					New(Text, TextData{"c"}),
				),
			},
		},
		{
			nodes: []*Node{
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"z"}),
					New(Text, TextData{"c"}),
				),
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"c"}),
				),
			},
			left: []*Node{
				New(Text, TextData{"a"}),
			},
			right: []*Node{
				New(Text, TextData{"c"}),
			},
		},
		{
			nodes: []*Node{
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"c"}),
					New(Text, TextData{"d"}),
				),
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"c"}),
					New(Text, TextData{"c"}),
					New(Text, TextData{"d"}),
				),
			},
			left: []*Node{
				New(Text, TextData{"a"}),
				New(Text, TextData{"b"}),
			},
			right: []*Node{
				New(Text, TextData{"c"}),
				New(Text, TextData{"d"}),
			},
		},
		{
			nodes: []*Node{
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"c"}),
				),
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"b"}),
					New(Text, TextData{"c"}),
				),
			},
			left: []*Node{
				New(Text, TextData{"a"}),
				New(Text, TextData{"b"}),
			},
			right: []*Node{
				New(Text, TextData{"c"}),
			},
		},
		{
			nodes: []*Node{
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"d"}),
				),
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"d"}),
				),
				New(Nothing, nil,
					New(Text, TextData{"a"}),
					New(Text, TextData{"e"}),
				),
			},
			left: []*Node{
				New(Text, TextData{"a"}),
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
