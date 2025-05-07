package syntax

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func runesHead(s string, r int) string {
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

func runesTail(s string, r int) string {
	var i, n int
	for i = len(s); i >= 0; {
		var ok bool
		for j := 1; j <= 4 && i-j >= 0; j++ {
			v, _ := utf8.DecodeRuneInString(s[i-j:])
			if v != utf8.RuneError {
				i -= j
				n++
				ok = true
				break
			}
		}
		if !ok || n == r {
			return s[i:]
		}
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

func runesIndexAnyRune(s string, rs []rune) int {
	for _, r := range rs {
		if i := strings.IndexRune(s, r); i != -1 {
			return i
		}
	}
	return -1
}

func runesLastIndexAnyRune(s string, rs []rune) int {
	for _, r := range rs {
		i := -1
		if 0 <= r && r < utf8.RuneSelf {
			i = strings.LastIndexByte(s, byte(r))
		} else {
			sub := s
			for len(sub) > 0 {
				j := strings.IndexRune(s, r)
				if j == -1 {
					break
				}
				i = j
				sub = sub[i+1:]
			}
		}
		if i != -1 {
			return i
		}
	}
	return -1
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
