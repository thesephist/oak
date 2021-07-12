package main

import (
	"fmt"
	"unicode"
)

type tokenizer struct {
	source []rune
	index  int

	fileName string
	line     int
	col      int
}

type pos struct {
	fileName string
	line     int
	col      int
}

func (p pos) String() string {
	return fmt.Sprintf("[%d:%d]", p.line, p.col)
}

type tokKind int

const (
	// sentinel
	unknown tokKind = iota

	// language tokens
	comma
	dot
	leftParen
	rightParen
	leftBracket
	rightBracket
	leftBrace
	rightBrace
	assign
	nonlocalAssign
	branchArrow
	colon
	ellipsis
	qmark
	exclam

	// binary operators
	plus
	minus
	times
	divide
	modulus
	xor
	and
	or
	greater
	less
	eq
	geq
	leq

	// keywords
	ifKeyword
	fnKeyword
	withKeyword

	// identifiers and literals
	empty
	identifier
	trueLiteral
	falseLiteral
	stringLiteral
	numberLiteral
)

type token struct {
	kind tokKind
	pos
	payload string
}

func (t token) String() string {
	switch t.kind {
	case comma:
		return ","
	case dot:
		return "."
	case leftParen:
		return "("
	case rightParen:
		return ")"
	case leftBracket:
		return "["
	case rightBracket:
		return "]"
	case leftBrace:
		return "{"
	case rightBrace:
		return "}"
	case assign:
		return ":="
	case nonlocalAssign:
		return "<-"
	case branchArrow:
		return "->"
	case colon:
		return ":"
	case ellipsis:
		return "..."
	case qmark:
		return "?"
	case exclam:
		return "!"
	case plus:
		return "+"
	case minus:
		return "-"
	case times:
		return "*"
	case divide:
		return "/"
	case modulus:
		return "%"
	case xor:
		return "^"
	case and:
		return "&"
	case or:
		return "|"
	case greater:
		return ">"
	case less:
		return "<"
	case eq:
		return "="
	case geq:
		return ">="
	case leq:
		return "<="
	case ifKeyword:
		return "if"
	case fnKeyword:
		return "fn"
	case withKeyword:
		return "with"
	case empty:
		return "_"
	case identifier:
		return fmt.Sprintf("var(%s)", t.payload)
	case trueLiteral:
		return "true"
	case falseLiteral:
		return "false"
	case stringLiteral:
		return fmt.Sprintf("string(%s)", t.payload)
	case numberLiteral:
		return fmt.Sprintf("number(%s)", t.payload)
	default:
		return "(unknown token)"
	}
}

// TODO: refactor with io.RuneReader
func newTokenizer(sourceString string) tokenizer {
	return tokenizer{
		source:   []rune(sourceString),
		index:    0,
		fileName: "(input)",
		line:     0,
		col:      0,
	}
}

// TODO: correctly update positions with every next() call
func (t *tokenizer) currentPos() pos {
	return pos{
		fileName: t.fileName,
		line:     t.line,
		col:      t.col,
	}
}

func (t *tokenizer) isEOF() bool {
	return t.index == len(t.source)
}

func (t *tokenizer) peek() rune {
	return t.source[t.index]
}

func (t *tokenizer) peekAhead(n int) rune {
	if t.index+n > len(t.source) {
		// In Magnolia, whitespace is insingificant, so we return it as the
		// "nothing is here" value.
		return ' '
	}
	return t.source[t.index+n]
}

func (t *tokenizer) next() rune {
	char := t.source[t.index]

	if t.index < len(t.source) {
		t.index++
	}

	return char
}

func (t *tokenizer) back() {
	if t.index > 0 {
		t.index--
	}
}

func (t *tokenizer) readUntilRune(c rune) string {
	accumulator := []rune{}
	for !t.isEOF() && t.peek() != c {
		accumulator = append(accumulator, t.next())
	}
	return string(accumulator)
}

func (t *tokenizer) readValidIdentifier() string {
	accumulator := []rune{}
	for {
		if t.isEOF() {
			break
		}

		c := t.next()
		if unicode.IsLetter(c) || c == '_' || c == '?' || c == '!' {
			accumulator = append(accumulator, c)
		} else {
			t.back()
			break
		}
	}
	return string(accumulator)
}

func (t *tokenizer) nextToken() token {
	c := t.next()

	// TODO: tokenize comments
	switch c {
	case ',':
		return token{kind: comma, pos: t.currentPos()}
	case '.':
		if t.peek() == '.' && t.peekAhead(1) == '.' {
			pos := t.currentPos()
			t.next()
			t.next()
			return token{kind: ellipsis, pos: pos}
		}
		return token{kind: dot, pos: t.currentPos()}
	case '(':
		return token{kind: leftParen, pos: t.currentPos()}
	case ')':
		return token{kind: rightParen, pos: t.currentPos()}
	case '[':
		return token{kind: leftBracket, pos: t.currentPos()}
	case ']':
		return token{kind: rightBracket, pos: t.currentPos()}
	case '{':
		return token{kind: leftBrace, pos: t.currentPos()}
	case '}':
		return token{kind: rightBrace, pos: t.currentPos()}
	case ':':
		if unicode.IsDigit(t.peekAhead(1)) {
			// TODO: finish dot-leading decimals
			return token{kind: comma, pos: t.currentPos()}
		} else if t.peekAhead(1) == '=' {
			pos := t.currentPos()
			t.next()
			return token{kind: assign, pos: pos}
		}
		return token{kind: colon, pos: t.currentPos()}
	case '<':
		switch t.peekAhead(1) {
		case '-':
			t.next()
			return token{kind: nonlocalAssign, pos: t.currentPos()}
		case '=':
			t.next()
			return token{kind: leq, pos: t.currentPos()}
		}
		return token{kind: less, pos: t.currentPos()}
	case '?':
		return token{kind: qmark, pos: t.currentPos()}
	case '!':
		return token{kind: exclam, pos: t.currentPos()}
	case '+':
		return token{kind: plus, pos: t.currentPos()}
	case '-':
		switch t.peekAhead(1) {
		case '>':
			t.next()
			return token{kind: branchArrow, pos: t.currentPos()}
		}
		return token{kind: minus, pos: t.currentPos()}
	case '*':
		return token{kind: times, pos: t.currentPos()}
	case '/':
		return token{kind: divide, pos: t.currentPos()}
	case '%':
		return token{kind: modulus, pos: t.currentPos()}
	case '^':
		return token{kind: xor, pos: t.currentPos()}
	case '&':
		return token{kind: and, pos: t.currentPos()}
	case '|':
		return token{kind: or, pos: t.currentPos()}
	case '>':
		if t.peekAhead(1) == '=' {
			pos := t.currentPos()
			t.next()
			return token{kind: geq, pos: pos}
		}
		return token{kind: greater, pos: t.currentPos()}
	case '=':
		return token{kind: eq, pos: t.currentPos()}
	case '\'':
		// TODO: support escape sequences
		// TODO: support literal newlines, extra tabs collapsed to newlines
		// TODO: support escaped quote
		pos := t.currentPos()
		payload := t.readUntilRune('\'')
		t.next() // read ending quote
		return token{
			kind:    stringLiteral,
			pos:     pos,
			payload: payload,
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// TODO: implement
		fallthrough
	default:
		pos := t.currentPos()
		payload := string(c) + t.readValidIdentifier()
		switch payload {
		case "_":
			return token{kind: empty, pos: pos}
		case "if":
			return token{kind: ifKeyword, pos: pos}
		case "fn":
			return token{kind: fnKeyword, pos: pos}
		case "with":
			return token{kind: withKeyword, pos: pos}
		case "true":
			return token{kind: trueLiteral, pos: pos}
		case "false":
			return token{kind: falseLiteral, pos: pos}
		default:
			return token{kind: identifier, pos: pos, payload: payload}
		}
	}
}

func (t *tokenizer) tokenize() []token {
	tokens := []token{}

	// snip whitespace before
	for !t.isEOF() && unicode.IsSpace(t.peek()) {
		t.next()
	}

	last := token{kind: comma}
	for !t.isEOF() {
		next := t.nextToken()

		if (last.kind != leftParen && last.kind != leftBracket &&
			last.kind != leftBrace && last.kind != comma) &&
			(next.kind == rightParen || next.kind == rightBracket ||
				next.kind == rightBrace) {
			tokens = append(tokens, token{
				kind: comma,
				pos:  t.currentPos(),
			})
		}
		tokens = append(tokens, next)

		// snip whitespace after
		for !t.isEOF() && unicode.IsSpace(t.peek()) {
			if t.peek() == '\n' {
				switch next.kind {
				case comma, leftParen, leftBracket, leftBrace, plus, minus,
					times, divide, modulus, exclam, greater, less, eq, geq, leq,
					assign, nonlocalAssign, dot, colon, fnKeyword, ifKeyword,
					withKeyword, branchArrow:
					// do nothing
				default:
					next = token{
						kind: comma,
						pos:  t.currentPos(),
					}
					tokens = append(tokens, next)
				}
			}
			t.next()
		}

		last = next
	}

	if last.kind != comma {
		tokens = append(tokens, token{
			kind: comma,
			pos:  t.currentPos(),
		})
	}

	return tokens
}
