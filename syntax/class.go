package syntax

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ClassData is the value of a node of type [Class]: a bracket expression that
// holds more than one list or range, or a named class.
type ClassData struct {
	Not     bool
	Fold    bool
	Chars   string
	Ranges  []RangeData
	Classes []string
}

// Contains reports whether the class matches the rune r. When Fold is set,
// the class matches r when it holds any rune of the case folding orbit of r.
func (d ClassData) Contains(r rune) bool {
	in := d.contains(r)
	if d.Fold && !in {
		for c := unicode.SimpleFold(r); c != r && !in; c = unicode.SimpleFold(c) {
			in = d.contains(c)
		}
	}
	return in != d.Not
}

// contains reports whether r is a member of the class, without the
// negation.
func (d ClassData) contains(r rune) bool {
	if strings.ContainsRune(d.Chars, r) {
		return true
	}
	for _, rng := range d.Ranges {
		if rng.Lo <= r && r <= rng.Hi {
			return true
		}
	}
	for _, name := range d.Classes {
		if classes[name](r) {
			return true
		}
	}
	return false
}

// String satisfies the [fmt.Stringer] interface.
func (d ClassData) String() string {
	var b strings.Builder
	b.WriteByte('[')
	if d.Not {
		b.WriteByte('!')
	}
	b.WriteString(d.Chars)
	for _, rng := range d.Ranges {
		b.WriteRune(rng.Lo)
		b.WriteByte('-')
		b.WriteRune(rng.Hi)
	}
	for _, name := range d.Classes {
		b.WriteString("[:" + name + ":]")
	}
	b.WriteByte(']')
	return b.String()
}

// classes holds the named classes of a bracket expression, such as
// [:alpha:]. A class that POSIX defines for ASCII only, such as digit and
// xdigit, matches ASCII only. The other classes follow the Unicode categories,
// as a shell does in a UTF-8 locale.
var classes = map[string]func(rune) bool{
	"alnum":  isAlnum,
	"alpha":  unicode.IsLetter,
	"ascii":  func(r rune) bool { return 0 <= r && r < utf8.RuneSelf },
	"blank":  func(r rune) bool { return r == ' ' || r == '\t' || unicode.Is(unicode.Zs, r) },
	"cntrl":  unicode.IsControl,
	"digit":  isDigit,
	"graph":  isGraph,
	"lower":  unicode.IsLower,
	"print":  unicode.IsPrint,
	"punct":  isPunct,
	"space":  unicode.IsSpace,
	"upper":  unicode.IsUpper,
	"word":   func(r rune) bool { return r == '_' || isAlnum(r) },
	"xdigit": isXdigit,
}

// isClassName reports whether name is a named class.
func isClassName(name string) bool {
	_, ok := classes[name]
	return ok
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isXdigit(r rune) bool {
	return isDigit(r) || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F'
}

func isAlnum(r rune) bool {
	return isDigit(r) || unicode.IsLetter(r)
}

func isGraph(r rune) bool {
	return unicode.IsGraphic(r) && !unicode.IsSpace(r)
}

func isPunct(r rune) bool {
	if r < utf8.RuneSelf {
		return isGraph(r) && !isAlnum(r)
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// ClassMatcher matches one rune of a [ClassData].
type ClassMatcher struct {
	data ClassData
}

// NewClass creates a class matcher.
func NewClass(data ClassData) ClassMatcher {
	return ClassMatcher{data}
}

// Match satisfies the [Matcher] interface.
func (m ClassMatcher) Match(s string) bool {
	r, w := utf8.DecodeRuneInString(s)
	if w == 0 || len(s) > w {
		return false
	}
	return m.data.Contains(r)
}

// Len satisfies the [Matcher] interface.
func (ClassMatcher) Len() int {
	return 1
}

// Size satisfies the [Sizer] interface.
func (ClassMatcher) Size() int {
	return 1
}

// Index satisfies the [Indexer] interface.
func (m ClassMatcher) Index(s string) (int, []int) {
	for i, r := range s {
		if m.data.Contains(r) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m ClassMatcher) String() string {
	return fmt.Sprintf("<class:%s>", m.data)
}

// ExtData is the value of a node of type [Ext]: an extended group such as
// @(a|b). Op is the rune before the parenthesis.
type ExtData struct {
	Op rune
}

// String satisfies the [fmt.Stringer] interface.
func (d ExtData) String() string {
	return string(d.Op)
}

// MaxSequence is the largest number of items that a brace sequence, such as
// {1..10}, can expand to.
const MaxSequence = 10000

// expandSequence expands the body of a brace sequence, such as 1..10,
// 01..10..2 or a..e, to an [AnyOf] node. It follows bash: a number with a
// leading zero pads every item to the same width, and the step is the
// absolute value of the third part.
func expandSequence(raw string) (*Node, error) {
	parts := strings.Split(raw, "..")
	step := 1
	if len(parts) == 3 {
		n, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("parsing step of brace sequence %q: %w", raw, err)
		}
		step = max(abs(n), 1)
	}
	var items []string
	switch lo, hi := parts[0], parts[1]; {
	case !isDigit(rune(lo[len(lo)-1])):
		// The two ends are ASCII letters, so the sequence has at most 58
		// items.
		for _, r := range sequence(rune(lo[0]), rune(hi[0]), rune(min(step, utf8.RuneSelf))) {
			items = append(items, string(r))
		}
	default:
		a, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("parsing brace sequence %q: %w", raw, err)
		}
		b, err := strconv.Atoi(hi)
		if err != nil {
			return nil, fmt.Errorf("parsing brace sequence %q: %w", raw, err)
		}
		if n := abs(b-a)/step + 1; n > MaxSequence || n < 0 {
			return nil, fmt.Errorf("brace sequence %q has more than %d items", raw, MaxSequence)
		}
		width := 0
		if padded(lo) || padded(hi) {
			width = max(len(lo), len(hi))
		}
		for _, n := range sequence(a, b, step) {
			items = append(items, pad(n, width))
		}
	}
	node := New(AnyOf, nil)
	for _, item := range items {
		node.Insert(New(Pattern, nil, New(Text, TextData{item})))
	}
	return node, nil
}

// sequence returns the integers from a to b by step, in either direction.
func sequence[T int | rune](a, b, step T) []T {
	var v []T
	if a <= b {
		for n := a; n <= b; n += step {
			v = append(v, n)
		}
	} else {
		for n := a; n >= b; n -= step {
			v = append(v, n)
		}
	}
	return v
}

// padded reports whether the number s has a leading zero.
func padded(s string) bool {
	s = strings.TrimLeft(s, "+-")
	return len(s) > 1 && s[0] == '0'
}

// pad formats n with zeros to the width, which counts the sign.
func pad(n, width int) string {
	s := strconv.Itoa(abs(n))
	sign := ""
	if n < 0 {
		sign = "-"
	}
	if w := width - len(sign); len(s) < w {
		s = strings.Repeat("0", w-len(s)) + s
	}
	return sign + s
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// FoldRune returns the smallest rune of the case folding orbit of r. Two runes
// fold to the same rune when [strings.EqualFold] treats them as equal.
func FoldRune(r rune) rune {
	m := r
	for c := unicode.SimpleFold(r); c != r; c = unicode.SimpleFold(c) {
		m = min(m, c)
	}
	return m
}

// FoldString maps each rune of s with [FoldRune].
func FoldString(s string) string {
	return strings.Map(FoldRune, s)
}

// Fold returns a copy of the tree for matching without regard to case. A
// caller folds the input with [FoldString] before matching it.
func Fold(node *Node) *Node {
	n := New(node.Type, node.Value)
	switch node.Type {
	case Text:
		n.Value = TextData{FoldString(node.Value.(TextData).Text)}
	case List:
		l := node.Value.(ListData)
		n = New(Class, ClassData{Not: l.Not, Fold: true, Chars: l.Chars})
	case Range:
		r := node.Value.(RangeData)
		n = New(Class, ClassData{Not: r.Not, Fold: true, Ranges: []RangeData{{Lo: r.Lo, Hi: r.Hi}}})
	case Class:
		c := node.Value.(ClassData)
		c.Fold = true
		c.Ranges = slices.Clone(c.Ranges)
		c.Classes = slices.Clone(c.Classes)
		n.Value = c
	}
	for _, c := range node.Children {
		n.Insert(Fold(c))
	}
	return n
}

// Globstar returns a copy of the tree in which a super star that is a whole
// segment and is followed by a separator also matches no segment at all. In
// the pattern a/**/b, the part **/ becomes {**/,}, so the pattern matches a/b
// as well as a/x/b. This is the globstar of bash and of other shells.
func Globstar(node *Node, sep []rune) *Node {
	return globstar(node, sep, true)
}

// globstar rewrites the tree for [Globstar]. The flag start reports whether
// the node starts a segment.
func globstar(node *Node, sep []rune, start bool) *Node {
	n := New(node.Type, node.Value)
	children := node.Children
	for i := 0; i < len(children); i++ {
		c := children[i]
		first := start
		if node.Type == Pattern {
			first = segmentStart(children, i, sep, start)
		}
		if node.Type == Pattern && first && (c.Type == Super || c.Type == Any) {
			// A run of stars that holds a super star is a super star.
			j, super := i, false
			for ; j < len(children) && (children[j].Type == Super || children[j].Type == Any); j++ {
				super = super || children[j].Type == Super
			}
			if super && j < len(children) && children[j].Type == Text {
				text := children[j].Value.(TextData).Text
				if r, w := utf8.DecodeRuneInString(text); slices.Contains(sep, r) {
					n.Insert(New(AnyOf, nil,
						New(Pattern, nil, New(Super, nil), New(Text, TextData{text[:w]})),
						New(Pattern, nil),
					))
					if rest := text[w:]; rest != "" {
						n.Insert(New(Text, TextData{rest}))
					}
					i = j
					continue
				}
			}
		}
		n.Insert(globstar(c, sep, first))
	}
	return n
}

// segmentStart reports whether the child i of a pattern starts a segment:
// it is the first child of a pattern that starts one, or the child before it
// ends a segment.
func segmentStart(children []*Node, i int, sep []rune, start bool) bool {
	if i == 0 {
		return start
	}
	return segmentEnd(children[i-1], sep, segmentStart(children, i-1, sep, start))
}

// segmentEnd reports whether every match of the node ends with a separator,
// or is empty after a segment start. The flag start reports whether the node
// starts a segment.
func segmentEnd(node *Node, sep []rune, start bool) bool {
	switch node.Type {
	case Text:
		r, _ := utf8.DecodeLastRuneInString(node.Value.(TextData).Text)
		return slices.Contains(sep, r)
	case Pattern:
		if len(node.Children) == 0 {
			return start
		}
		last := len(node.Children) - 1
		return segmentEnd(node.Children[last], sep, segmentStart(node.Children, last, sep, start))
	case AnyOf:
		for _, c := range node.Children {
			if !segmentEnd(c, sep, start) {
				return false
			}
		}
		return len(node.Children) != 0
	}
	return false
}
