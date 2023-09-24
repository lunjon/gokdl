package gokdl

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type ParseContext struct {
	logger *log.Logger
}

type Parser struct {
	sc     *scanner
	logger *log.Logger
}

func newParser(bs []byte) *Parser {
	r := bytes.NewReader(bs)
	return &Parser{
		sc:     newScanner(r),
		logger: log.New(os.Stderr, "", 0),
	}
}

func (p *Parser) parse() (Doc, error) {
	cx := &ParseContext{
		logger: p.logger,
	}

	nodes, err := parseScope(cx, p.sc, false)

	return Doc{
		nodes: nodes,
	}, err
}

func parseScope(cx *ParseContext, sc *scanner, isChild bool) ([]Node, error) {
	nodes := []Node{}
	done := false

	for !done {
		token, lit := sc.scan()
		if token == EOF {
			break
		}

		switch token {
		case WS:
			continue
		case SEMICOLON:
			continue
		case CBRACK_CLOSE:
			if isChild {
				done = true
			} else {
				return nil, fmt.Errorf("unexpected token: %s", lit)
			}
		case COMMENT_LINE:
			sc.scanLine()
		case COMMENT_MUL_OPEN:
			if err := scanMultilineComment(cx, sc); err != nil {
				return nil, err
			}
		case COMMENT_SD:
			panic("todo")
		case QUOTE:
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
			if IsInitialIdentToken(token) {
				sc.unread()
				lit := sc.scanBareIdent()

				node, err := scanNode(cx, sc, lit)
				if err != nil {
					cx.logger.Println("Error scanning node:", lit, err)
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

func scanMultilineComment(cx *ParseContext, sc *scanner) error {
	cx.logger.Println("Scanning multiline comment")

	for {
		token, lit := sc.scan()
		if token == EOF {
			break
		}

		if token == COMMENT_MUL_CLOSE {
			return nil
		}
		cx.logger.Printf("Literal: %s", lit)
	}

	return fmt.Errorf("no closing of multiline comment")
}

func scanNode(cx *ParseContext, sc *scanner, name string) (Node, error) {
	children := []Node{}
	args := []Arg{}
	props := []Prop{}

	done := false
	for !done {
		token, lit := sc.scan()
		if token == EOF {
			break
		}

		cx.logger.Println(token, lit)

		switch token {
		case BACKSLASH:
			sc.scanWhitespace()
		case SEMICOLON:
			done = true
		case WS:
			// Newline (or semicolon) ends a node
			if strings.HasSuffix(lit, "\n") || strings.HasPrefix(lit, "\n") {
				done = true
			}
		case COMMENT_LINE:
			sc.scanLine()
			done = true
		case COMMENT_MUL_OPEN:
			if err := scanMultilineComment(cx, sc); err != nil {
				return Node{}, err
			}
		case COMMENT_SD:
			panic("todo")
		case NUM:
			arg := newArg(lit)
			args = append(args, arg)
		case QUOTE:
			s, err := scanString(cx, sc)
			if err != nil {
				return Node{}, err
			}

			nextToken, _ := sc.scan()
			if nextToken == EQUAL {
				sc.unread()
				prop, err := scanProp(cx, sc, s)
				if err != nil {
					return Node{}, err
				}
				props = append(props, prop)
			} else {
				sc.unread()
				arg := newArg(s)
				args = append(args, arg)
			}
		case CBRACK_OPEN:
			cx.logger.Println("Beginning to parse child scope")
			ns, err := parseScope(cx, sc, true)
			if err != nil {
				return Node{}, err
			}
			children = append(children, ns...)
		default:
			if IsInitialIdentToken(token) {
				sc.unread()
				id := sc.scanBareIdent()
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

func scanString(cx *ParseContext, sc *scanner) (string, error) {
	cx.logger.Println("Scanning string literal")

	buf := strings.Builder{}
	done := false
	for !done {
		token, lit := sc.scan()
		if token == EOF {
			return "", fmt.Errorf("error reading string literal: reached EOF")
		}

		switch token {
		case QUOTE:
			done = true
			continue
		default:
			buf.WriteString(lit)
		}
	}

	return buf.String(), nil
}

func scanProp(cx *ParseContext, sc *scanner, name string) (Prop, error) {
	cx.logger.Println("Scanning node property:", name)

	tok, _ := sc.scan()
	if tok != EQUAL {
		return Prop{}, fmt.Errorf("invalid node property: %s: expected '=' after identifier", name)
	}

	done := false
	var value any
	for !done {
		token, lit := sc.scan()
		if token == EOF {
			return Prop{}, fmt.Errorf("invalid node property: reached EOF")
		}

		switch token {
		case INVALID:
			return Prop{}, fmt.Errorf("invalid property value")
		case NUM:
			value = lit
			done = true
		case QUOTE:
			s, err := scanString(cx, sc)
			if err != nil {
				return Prop{}, err
			}
			value = s
			done = true
		default:
			// Not a number or string => try parse bool
			sc.unread()
			t, letters := sc.scanLetters()
			if t != EOF {
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
