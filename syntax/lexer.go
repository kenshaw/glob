package syntax

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TokenType is the type of a token.
type TokenType int

// Token types.
const (
	TokenEOF TokenType = iota
	TokenError
	TokenText
	TokenChar
	TokenAny
	TokenSuper
	TokenSingle
	TokenNot
	TokenSeparator
	TokenRangeOpen
	TokenRangeClose
	TokenRangeLo
	TokenRangeHi
	TokenRangeBetween
	TokenTermsOpen
	TokenTermsClose
	TokenClass
	TokenExtOpen
	TokenExtSeparator
	TokenExtClose
	TokenSequence
)

// String satisfies the [fmt.Stringer] interface.
func (typ TokenType) String() string {
	switch typ {
	case TokenEOF:
		return "eof"
	case TokenError:
		return "error"
	case TokenText:
		return "text"
	case TokenChar:
		return "char"
	case TokenAny:
		return "any"
	case TokenSuper:
		return "super"
	case TokenSingle:
		return "single"
	case TokenNot:
		return "not"
	case TokenSeparator:
		return "separator"
	case TokenRangeOpen:
		return "range_open"
	case TokenRangeClose:
		return "range_close"
	case TokenRangeLo:
		return "range_lo"
	case TokenRangeHi:
		return "range_hi"
	case TokenRangeBetween:
		return "range_between"
	case TokenTermsOpen:
		return "terms_open"
	case TokenTermsClose:
		return "terms_close"
	case TokenClass:
		return "class"
	case TokenExtOpen:
		return "ext_open"
	case TokenExtSeparator:
		return "ext_separator"
	case TokenExtClose:
		return "ext_close"
	case TokenSequence:
		return "sequence"
	}
	return "undef"
}

// Token is a token of a pattern.
type Token struct {
	Token TokenType
	Raw   string
}

// String satisfies the [fmt.Stringer] interface.
func (token Token) String() string {
	return fmt.Sprintf("%s<%q>", token.Token, token.Raw)
}

// Lexer splits a pattern into tokens.
type Lexer struct {
	src    string
	pos    int
	err    error
	tokens []Token
	// stack holds the open groups, innermost last: charTermsOpen for a brace
	// group and charExtOpen for an extended group.
	stack []rune
}

// NewLexer creates a lexer for the pattern.
func NewLexer(src string) *Lexer {
	l := &Lexer{
		src: src,
	}
	l.run()
	return l
}

// Next returns the next token. After the last token, it returns a token of
// type [TokenEOF]. If the pattern is not valid, it returns a token of type
// [TokenError].
func (l *Lexer) Next() Token {
	if l.err != nil {
		return Token{TokenError, l.err.Error()}
	}
	if len(l.tokens) == 0 {
		return Token{TokenEOF, ""}
	}
	token := l.tokens[0]
	l.tokens = l.tokens[1:]
	return token
}

// run reads the whole pattern.
func (l *Lexer) run() {
	for l.err == nil {
		if !l.fetchItem() {
			break
		}
	}
	switch {
	case l.err != nil:
		l.tokens = nil
	case len(l.stack) != 0:
		l.errorf("unexpected end of input: unclosed %q", string(l.stack[len(l.stack)-1]))
	default:
		l.push(TokenEOF, "")
	}
}

// peek returns the rune at the offset n runes from the current position, or
// eof.
func (l *Lexer) peek(n int) rune {
	pos := l.pos
	for {
		if pos >= len(l.src) {
			return eof
		}
		r, w := utf8.DecodeRuneInString(l.src[pos:])
		if n == 0 {
			return r
		}
		pos += w
		n--
	}
}

// read returns the rune at the current position and moves past it, or
// returns eof.
func (l *Lexer) read() rune {
	if l.pos >= len(l.src) {
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += w
	return r
}

// invalid reports whether the input at the current position is not valid
// UTF-8. The rune U+FFFD itself is valid.
func (l *Lexer) invalid() bool {
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	return r == utf8.RuneError && w == 1
}

// push adds a token.
func (l *Lexer) push(typ TokenType, raw string) {
	l.tokens = append(l.tokens, Token{typ, raw})
}

// errorf sets the error of the lexer.
func (l *Lexer) errorf(format string, v ...any) {
	if l.err == nil {
		l.err = fmt.Errorf(format, v...)
	}
}

// in reports whether the innermost open group is of the kind r.
func (l *Lexer) in(r rune) bool {
	return len(l.stack) != 0 && l.stack[len(l.stack)-1] == r
}

// fetchItem reads the next item. It returns false at the end of the input.
func (l *Lexer) fetchItem() bool {
	r := l.peek(0)
	switch {
	case r == eof:
		return false
	case r == charTermsOpen:
		l.read()
		if seq := l.sequence(); seq != "" {
			l.push(TokenSequence, seq)
			return true
		}
		l.stack = append(l.stack, charTermsOpen)
		l.push(TokenTermsOpen, string(r))
	case r == charComma && l.in(charTermsOpen):
		l.read()
		l.push(TokenSeparator, string(r))
	case r == charTermsClose && l.in(charTermsOpen):
		l.read()
		l.stack = l.stack[:len(l.stack)-1]
		l.push(TokenTermsClose, string(r))
	case r == charExtSeparator && l.in(charExtOpen):
		l.read()
		l.push(TokenExtSeparator, string(r))
	case r == charExtClose && l.in(charExtOpen):
		l.read()
		l.stack = l.stack[:len(l.stack)-1]
		l.push(TokenExtClose, string(r))
	case r == charRangeOpen:
		l.read()
		l.push(TokenRangeOpen, string(r))
		l.fetchRange()
	case r == charAny && l.peek(1) == charAny && l.peek(2) == charExtOpen:
		// **( is a star and the group *(, as in bash.
		l.read()
		l.push(TokenAny, string(r))
	case r == charAny && l.peek(1) == charAny:
		l.read()
		l.read()
		l.push(TokenSuper, string(r)+string(r))
	case isExtOp(r) && l.peek(1) == charExtOpen:
		l.read()
		l.read()
		l.stack = append(l.stack, charExtOpen)
		l.push(TokenExtOpen, string(r)+string(charExtOpen))
	case r == charSingle:
		l.read()
		l.push(TokenSingle, string(r))
	case r == charAny:
		l.read()
		l.push(TokenAny, string(r))
	default:
		l.fetchText()
	}
	return true
}

// sequenceRE matches the body of a brace sequence, such as {1..10},
// {01..10..2} or {a..e}, after the opening brace. A number has at most 18
// digits, so that no sum of two numbers overflows an int.
var sequenceRE = regexp.MustCompile(`^(?:[+-]?[0-9]{1,18}\.\.[+-]?[0-9]{1,18}|[a-zA-Z]\.\.[a-zA-Z])(?:\.\.[+-]?[0-9]{1,18})?\}`)

// sequence reads a brace sequence after an opening brace. It returns the
// body of the sequence without its braces, or an empty string when the input
// does not hold one. The opening brace is already read.
func (l *Lexer) sequence() string {
	m := sequenceRE.FindString(l.src[l.pos:])
	if m == "" {
		return ""
	}
	l.pos += len(m)
	return m[:len(m)-1]
}

// fetchRange reads a bracket expression after its opening bracket. A
// bracket expression holds characters, ranges such as a-z, and named
// classes such as [:alpha:].
func (l *Lexer) fetchRange() {
	if r := l.peek(0); r == charRangeNot || r == charRangeCaret {
		l.read()
		l.push(TokenNot, string(r))
	}
	var chars []rune
	flush := func() {
		if len(chars) != 0 {
			l.push(TokenText, string(chars))
			chars = nil
		}
	}
	for first := true; ; first = false {
		r := l.peek(0)
		switch {
		case r == eof:
			l.errorf("unexpected end of input: unclosed %q", string(charRangeOpen))
			return
		case r == charRangeClose && !first:
			l.read()
			flush()
			l.push(TokenRangeClose, string(r))
			return
		case r == charRangeOpen && l.peek(1) == charColon:
			name, ok := l.bracketed(charColon)
			if !ok {
				break
			}
			if !isClassName(name) {
				l.errorf("unknown character class %q", name)
				return
			}
			flush()
			l.push(TokenClass, name)
			continue
		}
		lo, ok := l.rangeChar()
		if !ok {
			return
		}
		// A hyphen that is followed by the closing bracket is a character.
		if l.peek(0) == charRangeBetween && l.peek(1) != charRangeClose && l.peek(1) != eof {
			l.read()
			hi, ok := l.rangeChar()
			if !ok {
				return
			}
			flush()
			l.push(TokenRangeLo, string(lo))
			l.push(TokenRangeBetween, string(charRangeBetween))
			l.push(TokenRangeHi, string(hi))
			continue
		}
		chars = append(chars, lo)
	}
}

// rangeChar reads one character of a bracket expression: a character, an
// escaped character, an equivalence class such as [=a=], or a collating
// symbol such as [.-.].
func (l *Lexer) rangeChar() (rune, bool) {
	r := l.peek(0)
	if r == charRangeOpen && (l.peek(1) == charEquals || l.peek(1) == charPeriod) {
		delim := l.peek(1)
		if s, ok := l.bracketed(delim); ok {
			if utf8.RuneCountInString(s) != 1 {
				l.errorf("unsupported %q%s%q in character class", string(charRangeOpen)+string(delim), s, string(delim)+string(charRangeClose))
				return 0, false
			}
			c, _ := utf8.DecodeRuneInString(s)
			return c, true
		}
	}
	if l.invalid() {
		l.errorf("invalid utf-8 in pattern")
		return 0, false
	}
	l.read()
	switch r {
	case eof:
		l.errorf("unexpected end of input: unclosed %q", string(charRangeOpen))
		return 0, false
	case charEscape:
		c := l.read()
		if c == eof {
			l.errorf("unexpected end of input after %q", string(charEscape))
			return 0, false
		}
		return c, true
	}
	return r, true
}

// bracketed reads a bracketed name such as [:alpha:], where delim is the
// colon. It returns the name and true. If the input does not hold the closing
// delimiter and bracket, it reads nothing and returns false.
func (l *Lexer) bracketed(delim rune) (string, bool) {
	start := l.pos + 2 // after the bracket and the delimiter
	end := strings.Index(l.src[start:], string(delim)+string(charRangeClose))
	if end == -1 {
		return "", false
	}
	l.pos = start + end + 2
	return l.src[start : start+end], true
}

// fetchText reads text up to the next special character.
func (l *Lexer) fetchText() {
	var data []rune
	for {
		r := l.peek(0)
		switch {
		case r == eof:
		case l.invalid():
			l.errorf("invalid utf-8 in pattern")
			return
		case r == charEscape:
			l.read()
			c := l.read()
			if c == eof {
				l.errorf("unexpected end of input after %q", string(charEscape))
				return
			}
			data = append(data, c)
			continue
		case l.breaks(r):
		default:
			l.read()
			data = append(data, r)
			continue
		}
		break
	}
	if len(data) != 0 {
		l.push(TokenText, string(data))
	}
}

// breaks reports whether r, at the current position, ends a text.
func (l *Lexer) breaks(r rune) bool {
	switch r {
	case charSingle, charAny, charRangeOpen, charTermsOpen:
		return true
	case charComma, charTermsClose:
		return l.in(charTermsOpen)
	case charExtSeparator, charExtClose:
		return l.in(charExtOpen)
	}
	return isExtOp(r) && l.peek(1) == charExtOpen
}

// IsSpecial returns true when c is a special character of the syntax.
func IsSpecial(c byte) bool {
	return strings.IndexByte(specials, c) != -1
}

// isExtOp reports whether r opens an extended group when a parenthesis
// follows it.
func isExtOp(r rune) bool {
	switch r {
	case charSingle, charAny, charPlus, charAt, charRangeNot:
		return true
	}
	return false
}

// eof is the rune that the lexer reads at the end of the input.
const eof rune = -1

// Characters of the syntax.
const (
	charAny          = '*'
	charComma        = ','
	charSingle       = '?'
	charEscape       = '\\'
	charRangeOpen    = '['
	charRangeClose   = ']'
	charTermsOpen    = '{'
	charTermsClose   = '}'
	charRangeNot     = '!'
	charRangeCaret   = '^'
	charRangeBetween = '-'
	charColon        = ':'
	charEquals       = '='
	charPeriod       = '.'
	charPlus         = '+'
	charAt           = '@'
	charExtOpen      = '('
	charExtClose     = ')'
	charExtSeparator = '|'
)

// specials holds the characters that [glob.Quote] escapes.
const specials = `*?\[]{},()|`
