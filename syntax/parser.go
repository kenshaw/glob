package syntax

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

type Lexer interface {
	Next() Token
}

type parseFn func(*Node, Lexer) (parseFn, *Node, error)

func Parse(lexer Lexer) (*Node, error) {
	var parser parseFn
	root := NewNode(KindPattern, nil)
	var (
		tree *Node
		err  error
	)
	for parser, tree = parserMain, root; parser != nil; {
		parser, tree, err = parser(tree, lexer)
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

func parserMain(tree *Node, lex Lexer) (parseFn, *Node, error) {
	for {
		token := lex.Next()
		switch token.Type {
		case TokenEOF:
			return nil, tree, nil
		case TokenError:
			return nil, tree, errors.New(token.Raw)
		case TokenText:
			Insert(tree, NewNode(KindText, Text{token.Raw}))
			return parserMain, tree, nil
		case TokenAny:
			Insert(tree, NewNode(KindAny, nil))
			return parserMain, tree, nil
		case TokenSuper:
			Insert(tree, NewNode(KindSuper, nil))
			return parserMain, tree, nil
		case TokenSingle:
			Insert(tree, NewNode(KindSingle, nil))
			return parserMain, tree, nil
		case TokenRangeOpen:
			return parserRange, tree, nil
		case TokenTermsOpen:
			a := NewNode(KindAnyOf, nil)
			Insert(tree, a)
			p := NewNode(KindPattern, nil)
			Insert(a, p)
			return parserMain, p, nil
		case TokenSeparator:
			p := NewNode(KindPattern, nil)
			Insert(tree.Parent, p)
			return parserMain, p, nil
		case TokenTermsClose:
			return parserMain, tree.Parent.Parent, nil
		default:
			return nil, tree, fmt.Errorf("unexpected token: %s", token)
		}
	}
	return nil, tree, fmt.Errorf("unknown error")
}

func parserRange(tree *Node, lex Lexer) (parseFn, *Node, error) {
	var (
		not   bool
		lo    rune
		hi    rune
		chars string
	)
	for {
		token := lex.Next()
		switch token.Type {
		case TokenEOF:
			return nil, tree, errors.New("unexpected end")
		case TokenError:
			return nil, tree, errors.New(token.Raw)
		case TokenNot:
			not = true
		case TokenRangeLo:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, fmt.Errorf("unexpected length of lo character")
			}
			lo = r
		case TokenRangeBetween:
			//
		case TokenRangeHi:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, fmt.Errorf("unexpected length of lo character")
			}
			hi = r
			if hi < lo {
				return nil, tree, fmt.Errorf("hi character '%s' should be greater than lo '%s'", string(hi), string(lo))
			}
		case TokenText:
			chars = token.Raw
		case TokenRangeClose:
			isRange := lo != 0 && hi != 0
			isChars := chars != ""
			if isChars == isRange {
				return nil, tree, fmt.Errorf("could not parse range")
			}
			if isRange {
				Insert(tree, NewNode(KindRange, Range{
					Lo:  lo,
					Hi:  hi,
					Not: not,
				}))
			} else {
				Insert(tree, NewNode(KindList, List{
					Chars: chars,
					Not:   not,
				}))
			}
			return parserMain, tree, nil
		}
	}
}
