package syntax

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// Parse builds a tree from the tokens read from the lexer.
func Parse(l *Lexer) (*Node, error) {
	return parse(l)
}

type lexer interface {
	Next() Token
}

func parse(l lexer) (*Node, error) {
	tree := New(KindPattern, nil)
	var err error
	for f, node := parseNode, tree; f != nil; {
		f, node, err = f(node, l)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

type parseFunc func(*Node, lexer) (parseFunc, *Node, error)

func parseNode(node *Node, l lexer) (parseFunc, *Node, error) {
	for {
		switch token := l.Next(); token.Token {
		case TokenEOF:
			return nil, node, nil
		case TokenError:
			return nil, node, errors.New(token.Raw)
		case TokenText:
			node.Insert(New(KindText, Text{token.Raw}))
			return parseNode, node, nil
		case TokenAny:
			node.Insert(New(KindAny, nil))
			return parseNode, node, nil
		case TokenSuper:
			node.Insert(New(KindSuper, nil))
			return parseNode, node, nil
		case TokenSingle:
			node.Insert(New(KindSingle, nil))
			return parseNode, node, nil
		case TokenRangeOpen:
			return parseRange, node, nil
		case TokenTermsOpen:
			n := New(KindAnyOf, nil)
			node.Insert(n)
			p := New(KindPattern, nil)
			n.Insert(p)
			return parseNode, p, nil
		case TokenSeparator:
			n := New(KindPattern, nil)
			node.Parent.Insert(n)
			return parseNode, n, nil
		case TokenTermsClose:
			return parseNode, node.Parent.Parent, nil
		default:
			return nil, node, fmt.Errorf("unexpected token: %s", token)
		}
	}
}

func parseRange(node *Node, l lexer) (parseFunc, *Node, error) {
	var (
		not   bool
		lo    rune
		hi    rune
		chars string
	)
	for {
		token := l.Next()
		switch token.Token {
		case TokenEOF:
			return nil, node, errors.New("unexpected end")
		case TokenError:
			return nil, node, errors.New(token.Raw)
		case TokenNot:
			not = true
		case TokenRangeLo:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, node, fmt.Errorf("unexpected length of lo character")
			}
			lo = r
		case TokenRangeBetween:
			//
		case TokenRangeHi:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, node, fmt.Errorf("unexpected length of lo character")
			}
			hi = r
			if hi < lo {
				return nil, node, fmt.Errorf("hi character '%s' should be greater than lo '%s'", string(hi), string(lo))
			}
		case TokenText:
			chars = token.Raw
		case TokenRangeClose:
			isRange := lo != 0 && hi != 0
			isChars := chars != ""
			if isChars == isRange {
				return nil, node, fmt.Errorf("could not parse range")
			}
			if isRange {
				node.Insert(New(KindRange, Range{
					Lo:  lo,
					Hi:  hi,
					Not: not,
				}))
			} else {
				node.Insert(New(KindList, List{
					Chars: chars,
					Not:   not,
				}))
			}
			return parseNode, node, nil
		}
	}
}
