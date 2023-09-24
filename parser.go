package gokdl

import (
	"bytes"
	"fmt"
	pkg "github.com/lunjon/gokdl/internal"
	"log"
	"strconv"
	"strings"
)

type parseContext struct {
	logger *log.Logger
}

type parser struct {
	sc     *pkg.Scanner
	logger *log.Logger
}

func newParser(logger *log.Logger, bs []byte) *parser {
	r := bytes.NewReader(bs)
	return &parser{
		sc:     pkg.NewScanner(r),
		logger: logger,
	}
}

func (p *parser) parse() (Doc, error) {
	cx := &parseContext{
		logger: p.logger,
	}

	nodes, err := parseScope(cx, p.sc, false)

	return Doc{
		nodes: nodes,
	}, err
}

func parseScope(cx *parseContext, sc *pkg.Scanner, isChild bool) ([]Node, error) {
	nodes := []Node{}
	done := false

	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			break
		}

		switch token {
		case pkg.WS:
			continue
		case pkg.SEMICOLON:
			continue
		case pkg.CBRACK_CLOSE:
			if isChild {
				done = true
			} else {
				return nil, fmt.Errorf("unexpected token: %s (isChild=%v)", lit, isChild)
			}
		case pkg.COMMENT_LINE:
			sc.ScanLine()
		case pkg.COMMENT_MUL_OPEN:
			if err := scanMultilineComment(cx, sc); err != nil {
				return nil, err
			}
		case pkg.COMMENT_SD:
			panic("todo")
		case pkg.QUOTE:
			// Identifier in quotes => parse as string
			lit, err := scanString(cx, sc)
			if err != nil {
				return nil, err
			}
			node, err := scanNode(cx, sc, lit)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		default:
			if pkg.IsInitialIdentToken(token) {
				fmt.Println("A:", lit)
				sc.Unread()
				text := sc.ScanBareIdent()
				fmt.Println("B:", text)

				node, err := scanNode(cx, sc, text)
				if err != nil {
					cx.logger.Println("Error scanning node:", text, err)
					return nil, err
				}
				nodes = append(nodes, node)
			} else {
				return nil, fmt.Errorf("unexpected token: %s", lit)
			}
		}
	}

	return nodes, nil
}

func scanMultilineComment(cx *parseContext, sc *pkg.Scanner) error {
	cx.logger.Println("Scanning multiline comment")

	for {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			break
		}

		if token == pkg.COMMENT_MUL_CLOSE {
			return nil
		}
		cx.logger.Printf("Literal: %s", lit)
	}

	return fmt.Errorf("no closing of multiline comment")
}

func scanNode(cx *parseContext, sc *pkg.Scanner, name string) (Node, error) {
	// This function gets called immediately after an
	// idenfitier was read. So just check that the following
	// token is valid.
	next, nextlit := sc.Scan()
	if !pkg.IsAnyOf(next, pkg.EOF, pkg.WS, pkg.SEMICOLON, pkg.CBRACK_CLOSE) {
		return Node{}, fmt.Errorf("unexpected token in identifier: %s", nextlit)
	}

	fmt.Println("Next:", nextlit)

	sc.Unread()

	children := []Node{}
	args := []Arg{}
	props := []Prop{}

	done := false
	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			break
		}

		cx.logger.Println(token, lit)

		switch token {
		case pkg.BACKSLASH:
			sc.ScanWhitespace()
		case pkg.SEMICOLON:
			done = true
		case pkg.WS:
			// Newline (or semicolon) ends a node
			if strings.HasSuffix(lit, "\n") || strings.HasPrefix(lit, "\n") {
				done = true
			}
		case pkg.COMMENT_LINE:
			sc.ScanLine()
			done = true
		case pkg.COMMENT_MUL_OPEN:
			if err := scanMultilineComment(cx, sc); err != nil {
				return Node{}, err
			}
		case pkg.COMMENT_SD:
			panic("todo")
		case pkg.NUM:
			arg := newArg(lit)
			args = append(args, arg)
		case pkg.QUOTE:
			s, err := scanString(cx, sc)
			if err != nil {
				return Node{}, err
			}

			nextToken, _ := sc.Scan()
			if nextToken == pkg.EQUAL {
				sc.Unread()
				prop, err := scanProp(cx, sc, s)
				if err != nil {
					return Node{}, err
				}
				props = append(props, prop)
			} else {
				sc.Unread()
				arg := newArg(s)
				args = append(args, arg)
			}
		case pkg.CBRACK_OPEN:
			cx.logger.Println("Beginning to parse child scope")
			ns, err := parseScope(cx, sc, true)
			if err != nil {
				return Node{}, err
			}
			children = append(children, ns...)
		default:
			cx.logger.Printf("%s: %s", name, lit)
			if pkg.IsInitialIdentToken(token) {
				sc.Unread()
				id := sc.ScanBareIdent()
				prop, err := scanProp(cx, sc, id)
				if err != nil {
					return Node{}, err
				}
				props = append(props, prop)
			} else {
				return Node{}, fmt.Errorf("unexpected token: %s", lit)
			}
		}
	}

	cx.logger.Printf("Succesfully scanned node %s", name)

	return Node{
		Name:     name,
		Children: children,
		Props:    props,
		Args:     args,
	}, nil
}

func scanString(cx *parseContext, sc *pkg.Scanner) (string, error) {
	cx.logger.Println("Scanning string literal")

	buf := strings.Builder{}
	done := false
	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			return "", fmt.Errorf("error reading string literal: reached EOF")
		}

		switch token {
		case pkg.QUOTE:
			done = true
		default:
			buf.WriteString(lit)
		}
	}

	return buf.String(), nil
}

func scanProp(cx *parseContext, sc *pkg.Scanner, name string) (Prop, error) {
	cx.logger.Println("Scanning node property:", name)

	tok, _ := sc.Scan()
	if tok != pkg.EQUAL {
		return Prop{}, fmt.Errorf("invalid node property: %s: expected '=' after identifier", name)
	}

	done := false
	var value any
	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			return Prop{}, fmt.Errorf("invalid node property: reached EOF")
		}

		switch token {
		case pkg.INVALID:
			return Prop{}, fmt.Errorf("invalid property value")
		case pkg.NUM:
			value = lit
			done = true
		case pkg.QUOTE:
			s, err := scanString(cx, sc)
			if err != nil {
				return Prop{}, err
			}
			value = s
			done = true
		default:
			// Not a number or string => try parse bool
			sc.Unread()
			t, letters := sc.ScanLetters()
			if t != pkg.EOF {
				b, err := strconv.ParseBool(letters)
				if err != nil {
					return Prop{}, err
				}
				value = b
				done = true
			} else {
				return Prop{}, fmt.Errorf("invalid property value")

			}
		}
	}

	cx.logger.Printf("Succesfully scanned property: %s=%v", name, value)

	return Prop{
		Name:  name,
		Value: value,
	}, nil
}
