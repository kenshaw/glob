package syntax

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"
)

type TokenType int

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
)

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
	}
	return "undef"
}

type Token struct {
	Token TokenType
	Raw   string
}

func (token Token) String() string {
	return fmt.Sprintf("%s<%q>", token.Token, token.Raw)
}

type tokens []Token

func (v *tokens) shift() (ret Token) {
	ret = (*v)[0]
	copy(*v, (*v)[1:])
	*v = (*v)[:len(*v)-1]
	return
}

func (v *tokens) push(token Token) {
	*v = append(*v, token)
}

func (v *tokens) empty() bool {
	return len(*v) == 0
}

type Lexer struct {
	src          string
	pos          int
	err          error
	tokens       tokens
	termsLevel   int
	lastRune     rune
	lastRuneSize int
	hasRune      bool
}

func NewLexer(src string) *Lexer {
	l := &Lexer{
		src:    src,
		tokens: make(tokens, 0, 4),
	}
	return l
}

func (l *Lexer) Next() Token {
	if l.err != nil {
		return Token{TokenError, l.err.Error()}
	}
	if !l.tokens.empty() {
		return l.tokens.shift()
	}
	l.fetchItem()
	return l.Next()
}

func (l *Lexer) peek() (r rune, w int) {
	if l.pos == len(l.src) {
		return 0, 0
	}
	r, w = utf8.DecodeRuneInString(l.src[l.pos:])
	if r == utf8.RuneError {
		l.errorf("could not read rune")
		r, w = 0, 0
	}
	return
}

func (l *Lexer) read() rune {
	if l.hasRune {
		l.hasRune = false
		l.seek(l.lastRuneSize)
		return l.lastRune
	}
	r, s := l.peek()
	l.seek(s)
	l.lastRune = r
	l.lastRuneSize = s
	return r
}

func (l *Lexer) seek(w int) {
	l.pos += w
}

func (l *Lexer) unread() {
	if l.hasRune {
		l.errorf("could not unread rune")
		return
	}
	l.seek(-l.lastRuneSize)
	l.hasRune = true
}

func (l *Lexer) errorf(s string) {
	l.err = errors.New(s)
}

func (l *Lexer) inTerms() bool {
	return l.termsLevel > 0
}

func (l *Lexer) termsEnter() {
	l.termsLevel++
}

func (l *Lexer) termsLeave() {
	l.termsLevel--
}

func (l *Lexer) fetchItem() {
	r := l.read()
	switch {
	case r == 0:
		l.tokens.push(Token{TokenEOF, ""})
	case r == char_terms_open:
		l.termsEnter()
		l.tokens.push(Token{TokenTermsOpen, string(r)})
	case r == char_comma && l.inTerms():
		l.tokens.push(Token{TokenSeparator, string(r)})
	case r == char_terms_close && l.inTerms():
		l.tokens.push(Token{TokenTermsClose, string(r)})
		l.termsLeave()
	case r == char_range_open:
		l.tokens.push(Token{TokenRangeOpen, string(r)})
		l.fetchRange()
	case r == char_single:
		l.tokens.push(Token{TokenSingle, string(r)})
	case r == charAny:
		if l.read() == charAny {
			l.tokens.push(Token{TokenSuper, string(r) + string(r)})
		} else {
			l.unread()
			l.tokens.push(Token{TokenAny, string(r)})
		}
	default:
		l.unread()
		var breakers []rune
		if l.inTerms() {
			breakers = inTermsBreakers
		} else {
			breakers = inTextBreakers
		}
		l.fetchText(breakers)
	}
}

func (l *Lexer) fetchRange() {
	var wantHi bool
	var wantClose bool
	var seenNot bool
	for {
		r := l.read()
		if r == 0 {
			l.errorf("unexpected end of input")
			return
		}
		if wantClose {
			if r != char_range_close {
				l.errorf("expected close range character")
			} else {
				l.tokens.push(Token{TokenRangeClose, string(r)})
			}
			return
		}
		if wantHi {
			l.tokens.push(Token{TokenRangeHi, string(r)})
			wantClose = true
			continue
		}
		if !seenNot && r == char_range_not {
			l.tokens.push(Token{TokenNot, string(r)})
			seenNot = true
			continue
		}
		if n, w := l.peek(); n == char_range_between {
			l.seek(w)
			l.tokens.push(Token{TokenRangeLo, string(r)})
			l.tokens.push(Token{TokenRangeBetween, string(n)})
			wantHi = true
			continue
		}
		l.unread() // unread first peek and fetch as text
		l.fetchText([]rune{char_range_close})
		wantClose = true
	}
}

func (l *Lexer) fetchText(breakers []rune) {
	var data []rune
	var escaped bool
loop:
	for {
		r := l.read()
		if r == 0 {
			break
		}
		if !escaped {
			if r == char_escape {
				escaped = true
				continue
			}
			if runesIndexRune(breakers, r) != -1 {
				l.unread()
				break loop
			}
		}
		escaped = false
		data = append(data, r)
	}
	if len(data) > 0 {
		l.tokens.push(Token{TokenText, string(data)})
	}
}

func IsSpecial(c byte) bool {
	return bytes.IndexByte(specials, c) != -1
}

const (
	charAny            = '*'
	char_comma         = ','
	char_single        = '?'
	char_escape        = '\\'
	char_range_open    = '['
	char_range_close   = ']'
	char_terms_open    = '{'
	char_terms_close   = '}'
	char_range_not     = '!'
	char_range_between = '-'
)

var specials = []byte{
	charAny,
	char_single,
	char_escape,
	char_range_open,
	char_range_close,
	char_terms_open,
	char_terms_close,
}

var (
	inTextBreakers  = []rune{char_single, charAny, char_range_open, char_terms_open}
	inTermsBreakers = append(inTextBreakers, char_terms_close, char_comma)
)
