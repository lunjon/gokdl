package gokdl

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	pkg "github.com/lunjon/gokdl/internal"
)

type parseContext struct {
	logger *log.Logger
}

// The type responsible for parsing the documents.
// The parser relies on the Scanner (internal) for
// parsing.
//
// Specification: https://github.com/kdl-org/kdl/blob/main/SPEC.md
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

// Parses a root or child scope (inside a node).
func parseScope(cx *parseContext, sc *pkg.Scanner, isChild bool) ([]Node, error) {
	nodes := []Node{} // The nodes accumulated in this scope
	done := false     // When true, parsing of the scope (root or children) is done

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
			// Parse the following content as node and ignore the result
			nextToken, _ := sc.Scan()
			if pkg.IsInitialIdentToken(nextToken) {
				text := sc.ScanBareIdent()
				if _, err := scanNode(cx, sc, text); err != nil {
					return nil, fmt.Errorf("expected a node after slash-dash comment: %s", err)
				}
			} else {
				return nil, fmt.Errorf("expected a node after slash-dash comment")
			}
		case pkg.QUOTE:
			// Identifier in quotes => parse as string
			lit, err := scanString(cx, sc, "")
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
				text := sc.ScanBareIdent()
				node, err := scanNode(cx, sc, lit+text)
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

	sc.Unread()

	children := []Node{}
	args := []Arg{}
	props := []Prop{}

	done := false
	skip := false // Used with slash-dash comments

	typeAnnotation := ""
	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			break
		}

		if typeAnnotation != "" && pkg.IsAnyOf(token, pkg.BACKSLASH, pkg.SEMICOLON, pkg.CBRACK_OPEN) {
			return Node{}, fmt.Errorf("unexpected type annotation")
		}

		switch token {
		case pkg.BACKSLASH:
			sc.ScanWhitespace()
		case pkg.SEMICOLON:
			done = true
		case pkg.WS:
			// FIXME: handle other types of newline control characters (see the spec)
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
			// We need to continue to parse and ignore the next result.
			skip = true
			// typeAnnotation = ""
		case pkg.NUM_INT:
			if skip {
				skip = false
				typeAnnotation = ""
				continue
			}

			arg, err := newIntArg(lit, typeAnnotation)
			if err != nil {
				return Node{}, err
			}
			args = append(args, arg)
			typeAnnotation = ""
		case pkg.NUM_FLOAT, pkg.NUM_SCI:
			if skip {
				skip = false
				typeAnnotation = ""
				continue
			}

			arg, err := newFloatArg(lit, typeAnnotation)
			if err != nil {
				return Node{}, err
			}

			args = append(args, arg)
			typeAnnotation = ""
		case pkg.QUOTE:
			s, err := scanString(cx, sc, typeAnnotation)
			if err != nil {
				return Node{}, err
			}

			nextToken, _ := sc.Scan()
			if nextToken == pkg.EQUAL {
				prop, err := scanProp(cx, sc, s, typeAnnotation)
				if err != nil {
					return Node{}, err
				}

				if !skip {
					props = append(props, prop)
				}
				skip = false
			} else {
				if !skip {
					sc.Unread()
					arg := newArg(s, TypeAnnotation(typeAnnotation))
					args = append(args, arg)
				}

				skip = false
			}

			typeAnnotation = ""
		case pkg.CBRACK_OPEN:
			ns, err := parseScope(cx, sc, true)
			if err != nil {
				return Node{}, err
			}

			if !skip {
				children = append(children, ns...)
			}

			skip = false
		case pkg.CBRACK_CLOSE:
			done = true
			// sc.Unread()
		case pkg.PAREN_OPEN:
			annot, err := scanTypeAnnotation(cx, sc)
			if err != nil {
				return Node{}, err
			}
			typeAnnotation = annot
		default:
			// At this point there are multiple cases that can happen:
			// - The following value is a literal: null, true, false
			//   - These should be treated as such
			// - It is the start of a property name
			//
			// All the literals have valid initial identifier tokens.
			// That is, n(ull), t(rue) and f(alse) can be the start
			// of an identifier and NOT the literals.
			//
			// Thus we need to check the following tokens in order
			// to decide what it is.

			{ // Check literals
				var value any
				var ok bool

				_, next := sc.ScanLetters()
				next = lit + next

				switch next {
				case "null":
					value = nil // Default for any...
					ok = true
				case "true":
					value = true
					ok = true
				case "false":
					value = false
					ok = true
				}

				if ok {
					if typeAnnotation != "" {
						return Node{}, fmt.Errorf("unexpected type annotation")
					}

					if !skip {
						args = append(args, newArg(value, ""))
						skip = false
					}
					continue
				} else {
					sc.Unread()
				}
			}

			if pkg.IsInitialIdentToken(token) {
				id := sc.ScanBareIdent()
				next, _ := sc.Scan()
				if next != pkg.EQUAL {
					return Node{}, fmt.Errorf("unexpected identifier")
				}

				prop, err := scanProp(cx, sc, lit+id, typeAnnotation)
				if err != nil {
					return Node{}, err
				}

				if !skip {
					props = append(props, prop)
				}
				skip = false
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

func scanString(cx *parseContext, sc *pkg.Scanner, typeAnnot string) (string, error) {
	cx.logger.Println("Scanning string literal")

	buf := strings.Builder{}
	done := false
	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			return "", fmt.Errorf("error reading string literal: reached EOF")
		}

		switch token {
		case pkg.BACKSLASH:
			next, nextLit := sc.Scan()
			if next == pkg.QUOTE {
				buf.WriteString(`\"`)
			} else {
				buf.WriteString(lit)
				buf.WriteString(nextLit)
			}
		case pkg.QUOTE:
			done = true
		default:
			buf.WriteString(lit)
		}
	}

	fmt.Println("UNQUOTING:", buf.String())
	sss, err := strconv.Unquote("\"" + buf.String() + "\"")
	if err != nil {
		return "", err
	}

	return parseStringValue(sss, typeAnnot)
}

func scanProp(cx *parseContext, sc *pkg.Scanner, name, typeAnnotation string) (Prop, error) {
	cx.logger.Println("Scanning node property:", name)
	_, _ = sc.ScanWhitespace()

	done := false
	var value any
	var valueTypeAnnot string

	for !done {
		token, lit := sc.Scan()
		if token == pkg.EOF {
			return Prop{}, fmt.Errorf("invalid node property: reached EOF")
		}

		switch token {
		case pkg.INVALID:
			return Prop{}, fmt.Errorf("invalid property value")
		case pkg.NUM_INT:
			n, err := parseIntValue(lit, valueTypeAnnot)
			if err != nil {
				return Prop{}, err
			}
			value = n
			done = true
		case pkg.NUM_FLOAT, pkg.NUM_SCI:
			n, err := parseFloatValue(lit, valueTypeAnnot)
			if err != nil {
				return Prop{}, err
			}
			value = n
			done = true
		case pkg.QUOTE:
			s, err := scanString(cx, sc, valueTypeAnnot)
			if err != nil {
				return Prop{}, err
			}
			value = s
			done = true
		case pkg.PAREN_OPEN:
			t, err := scanTypeAnnotation(cx, sc)
			if err != nil {
				return Prop{}, err
			}

			valueTypeAnnot = t
		default:
			// Not a number or string => try parse bool or null
			sc.Unread()
			t, letters := sc.ScanLetters()
			if t != pkg.EOF {
				switch letters {
				case "null":
					value = nil
				case "true":
					value = true
				case "false":
					value = false
				default:
					return Prop{}, fmt.Errorf("invalid property value")
				}

				if valueTypeAnnot != "" {
					return Prop{}, fmt.Errorf("unexpected type annotation")
				}

				done = true
			} else {
				return Prop{}, fmt.Errorf("invalid property value")
			}
		}
	}

	cx.logger.Printf("Succesfully scanned property: %s=%v", name, value)

	return Prop{
		Name:           name,
		TypeAnnot:      TypeAnnotation(typeAnnotation),
		Value:          value,
		ValueTypeAnnot: TypeAnnotation(valueTypeAnnot),
	}, nil
}

func scanTypeAnnotation(cx *parseContext, sc *pkg.Scanner) (string, error) {
	annot := sc.ScanWhile(func(r rune) bool {
		return unicode.In(r, unicode.Digit, unicode.Letter)
	})

	next, _ := sc.Scan()
	if next != pkg.PAREN_CLOSE {
		return "", fmt.Errorf("unclosed type annotation")
	}

	annot = strings.TrimSpace(annot)
	if annot == "" {
		return "", fmt.Errorf("invalid type annotation: empty")
	}

	return annot, nil
}
