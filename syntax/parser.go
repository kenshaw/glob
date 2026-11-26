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
	tree := New(Pattern, nil)
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
	switch token := l.Next(); token.Token {
	case TokenEOF:
		if node.Parent != nil {
			return nil, node, errors.New("unexpected end of input")
		}
		return nil, node, nil
	case TokenError:
		return nil, node, errors.New(token.Raw)
	case TokenText:
		node.Insert(New(Text, TextData{token.Raw}))
		return parseNode, node, nil
	case TokenAny:
		node.Insert(New(Any, nil))
		return parseNode, node, nil
	case TokenSuper:
		node.Insert(New(Super, nil))
		return parseNode, node, nil
	case TokenSingle:
		node.Insert(New(Single, nil))
		return parseNode, node, nil
	case TokenRangeOpen:
		return parseRange, node, nil
	case TokenSequence:
		n, err := expandSequence(token.Raw)
		if err != nil {
			return nil, node, err
		}
		node.Insert(n)
		return parseNode, node, nil
	case TokenTermsOpen:
		n := New(AnyOf, nil)
		node.Insert(n)
		p := New(Pattern, nil)
		n.Insert(p)
		return parseNode, p, nil
	case TokenExtOpen:
		op, _ := utf8.DecodeRuneInString(token.Raw)
		n := New(Ext, ExtData{Op: op})
		node.Insert(n)
		p := New(Pattern, nil)
		n.Insert(p)
		return parseNode, p, nil
	case TokenSeparator, TokenExtSeparator:
		if node.Parent == nil {
			return nil, node, fmt.Errorf("unexpected token: %s", token)
		}
		n := New(Pattern, nil)
		node.Parent.Insert(n)
		return parseNode, n, nil
	case TokenTermsClose, TokenExtClose:
		if node.Parent == nil || node.Parent.Parent == nil {
			return nil, node, fmt.Errorf("unexpected token: %s", token)
		}
		return parseNode, node.Parent.Parent, nil
	default:
		return nil, node, fmt.Errorf("unexpected token: %s", token)
	}
}

func parseRange(node *Node, l lexer) (parseFunc, *Node, error) {
	var (
		data   ClassData
		lo, hi rune
	)
	for {
		token := l.Next()
		switch token.Token {
		case TokenEOF:
			return nil, node, errors.New("unexpected end of input")
		case TokenError:
			return nil, node, errors.New(token.Raw)
		case TokenNot:
			data.Not = true
		case TokenRangeLo:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, node, errors.New("unexpected length of lo character")
			}
			lo = r
		case TokenRangeBetween:
		case TokenRangeHi:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, node, errors.New("unexpected length of hi character")
			}
			hi = r
			if hi < lo {
				return nil, node, fmt.Errorf("hi character '%s' should be greater than lo '%s'", string(hi), string(lo))
			}
			data.Ranges = append(data.Ranges, RangeData{Lo: lo, Hi: hi})
		case TokenText:
			data.Chars += token.Raw
		case TokenClass:
			data.Classes = append(data.Classes, token.Raw)
		case TokenRangeClose:
			switch {
			case len(data.Ranges) == 0 && len(data.Classes) == 0 && data.Chars != "":
				node.Insert(New(List, ListData{
					Chars: data.Chars,
					Not:   data.Not,
				}))
			case len(data.Ranges) == 1 && len(data.Classes) == 0 && data.Chars == "":
				node.Insert(New(Range, RangeData{
					Lo:  data.Ranges[0].Lo,
					Hi:  data.Ranges[0].Hi,
					Not: data.Not,
				}))
			case len(data.Ranges) == 0 && len(data.Classes) == 0:
				return nil, node, errors.New("could not parse range")
			default:
				node.Insert(New(Class, data))
			}
			return parseNode, node, nil
		default:
			return nil, node, fmt.Errorf("unexpected token: %s", token)
		}
	}
}
