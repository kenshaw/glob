package syntax

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func runesHead(s string, r int) string {
	if r <= 0 {
		return ""
	}
	var i, m int
	for i < len(s) {
		_, n := utf8.DecodeRuneInString(s[i:])
		i += n
		m += 1
		if m == r {
			break
		}
	}
	return s[:i]
}

// runesTail returns the last r runes of s. An invalid byte is one rune, as a
// range loop over s counts it.
func runesTail(s string, r int) string {
	i := len(s)
	for n := 0; n < r && i > 0; n++ {
		_, w := utf8.DecodeLastRuneInString(s[:i])
		i -= w
	}
	return s[i:]
}

func runesExactlyRunesCount(s string, n int) bool {
	var m int
	for range s {
		m++
		if m > n {
			return false
		}
	}
	return m == n
}

// runesIndexAnyRune returns the index of the first rune of s that is in rs,
// or -1.
func runesIndexAnyRune(s string, rs []rune) int {
	switch len(rs) {
	case 0:
		return -1
	case 1:
		return strings.IndexRune(s, rs[0])
	}
	return strings.IndexFunc(s, func(r rune) bool { return slices.Contains(rs, r) })
}

// runesLastIndexAnyRune returns the index of the last rune of s that is in
// rs, or -1.
func runesLastIndexAnyRune(s string, rs []rune) int {
	switch {
	case len(rs) == 0:
		return -1
	case len(rs) == 1 && 0 <= rs[0] && rs[0] < utf8.RuneSelf:
		return strings.LastIndexByte(s, byte(rs[0]))
	}
	return strings.LastIndexFunc(s, func(r rune) bool { return slices.Contains(rs, r) })
}

func runesIndex(s, needle []rune) int {
	ls, ln := len(s), len(needle)
	switch {
	case ln == 0:
		return 0
	case ln == 1:
		return runesIndexRune(s, needle[0])
	case ln == ls:
		if runesEqual(s, needle) {
			return 0
		}
		return -1
	case ln > ls:
		return -1
	}
head:
	for i := 0; i < ls && ls-i >= ln; i++ {
		for y := range ln {
			if s[i+y] != needle[y] {
				continue head
			}
		}
		return i
	}
	return -1
}

func runesLastIndex(s, needle []rune) int {
	ls, ln := len(s), len(needle)
	switch {
	case ln == 0:
		if ls == 0 {
			return 0
		}
		return ls
	case ln == 1:
		return runesIndexLastRune(s, needle[0])
	case ln == ls:
		if runesEqual(s, needle) {
			return 0
		}
		return -1
	case ln > ls:
		return -1
	}
head:
	for i := ls - 1; i >= 0 && i >= ln; i-- {
		for y := ln - 1; y >= 0; y-- {
			if s[i-(ln-y-1)] != needle[y] {
				continue head
			}
		}
		return i - ln + 1
	}
	return -1
}

// IndexAny returns the index of the first instance of any Unicode code point
// from chars in s, or -1 if no Unicode code point from chars is present in s.
func runesIndexAny(s, chars []rune) int {
	if len(chars) > 0 {
		for i, c := range s {
			if slices.Contains(chars, c) {
				return i
			}
		}
	}
	return -1
}

func runesIndexRune(s []rune, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}

func runesIndexLastRune(s []rune, r rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == r {
			return i
		}
	}
	return -1
}

func runesEqual(a, b []rune) bool {
	// TODO use bytes.Equal with unsafe.
	if len(a) == len(b) {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

// runesHasRunes reports whether s holds at least n runes.
func runesHasRunes(s string, n int) bool {
	if n <= 0 {
		return true
	}
	var m int
	for range s {
		m++
		if m >= n {
			return true
		}
	}
	return false
}

// runeWidth returns the width in bytes of the rune at s[i:]. An invalid byte
// has the width 1, as a range loop over s steps over it, and not the width of
// [utf8.RuneError].
func runeWidth(s string, i int) int {
	if s[i] < utf8.RuneSelf {
		return 1
	}
	_, w := utf8.DecodeRuneInString(s[i:])
	return w
}
