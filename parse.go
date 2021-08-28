package main

import (
	"fmt"
	"strconv"
	"strings"
)

type astNode interface {
	String() string
	pos() pos
}

type emptyNode struct {
	tok *token
}

func (n emptyNode) String() string {
	return "_"
}
func (n emptyNode) pos() pos {
	return n.tok.pos
}

type nullNode struct {
	tok *token
}

func (n nullNode) String() string {
	return "?"
}
func (n nullNode) pos() pos {
	return n.tok.pos
}

type stringNode struct {
	payload string
	tok     *token
}

func (n stringNode) String() string {
	return fmt.Sprintf("%s", strconv.Quote(n.payload))
}
func (n stringNode) pos() pos {
	return n.tok.pos
}

type numberNode struct {
	isInteger    bool
	intPayload   int64
	floatPayload float64
	tok          *token
}

func (n numberNode) String() string {
	if n.isInteger {
		return strconv.FormatInt(n.intPayload, 10)
	}
	return strconv.FormatFloat(n.floatPayload, 'g', -1, 64)
}
func (n numberNode) pos() pos {
	return n.tok.pos
}

type booleanNode struct {
	payload bool
	tok     *token
}

func (n booleanNode) String() string {
	if n.payload {
		return "true"
	}
	return "false"
}
func (n booleanNode) pos() pos {
	return n.tok.pos
}

type atomNode struct {
	payload string
	tok     *token
}

func (n atomNode) String() string {
	return ":" + n.payload
}
func (n atomNode) pos() pos {
	return n.tok.pos
}

type listNode struct {
	elems []astNode
	tok   *token
}

func (n listNode) String() string {
	elemStrings := make([]string, len(n.elems))
	for i, el := range n.elems {
		elemStrings[i] = el.String()
	}
	return "[" + strings.Join(elemStrings, ", ") + "]"
}
func (n listNode) pos() pos {
	return n.tok.pos
}

type objectEntry struct {
	key astNode
	val astNode
}

func (n objectEntry) String() string {
	return n.key.String() + ": " + n.val.String()
}

type objectNode struct {
	entries []objectEntry
	tok     *token
}

func (n objectNode) String() string {
	entryStrings := make([]string, len(n.entries))
	for i, ent := range n.entries {
		entryStrings[i] = ent.String()
	}
	return "{ " + strings.Join(entryStrings, ", ") + " }"
}
func (n objectNode) pos() pos {
	return n.tok.pos
}

type fnNode struct {
	name    string // "" for anonymous fns
	args    []string
	restArg string
	body    astNode
	tok     *token
}

func (n fnNode) String() string {
	var head string
	if n.name == "" {
		head = "fn"
	} else {
		head = "fn " + n.name
	}

	if n.restArg == "" {
		head += "(" + strings.Join(n.args, ", ") + ")"
	} else {
		head += "(" + strings.Join(n.args, ", ") + ", " + n.restArg + "...)"
	}

	return head + " " + n.body.String()
}
func (n fnNode) pos() pos {
	return n.tok.pos
}

type identifierNode struct {
	payload string
	tok     *token
}

func (n identifierNode) String() string {
	return n.payload
}
func (n identifierNode) pos() pos {
	return n.tok.pos
}

type assignmentNode struct {
	isLocal bool
	left    astNode
	right   astNode
	tok     *token
}

func (n assignmentNode) String() string {
	if n.isLocal {
		return n.left.String() + " := " + n.right.String()
	}
	return n.left.String() + " <- " + n.right.String()
}
func (n assignmentNode) pos() pos {
	return n.tok.pos
}

type propertyAccessNode struct {
	left  astNode
	right astNode
	tok   *token
}

func (n propertyAccessNode) String() string {
	return "(" + n.left.String() + "." + n.right.String() + ")"
}
func (n propertyAccessNode) pos() pos {
	return n.tok.pos
}

type unaryNode struct {
	op    tokKind
	right astNode
	tok   *token
}

func (n unaryNode) String() string {
	opTok := token{kind: n.op}
	return opTok.String() + n.right.String()
}
func (n unaryNode) pos() pos {
	return n.tok.pos
}

type binaryNode struct {
	op    tokKind
	left  astNode
	right astNode
	tok   *token
}

func (n binaryNode) String() string {
	opTok := token{kind: n.op}
	return "(" + n.left.String() + " " + opTok.String() + " " + n.right.String() + ")"
}
func (n binaryNode) pos() pos {
	return n.tok.pos
}

type fnCallNode struct {
	fn      astNode
	args    []astNode
	restArg astNode
	tok     *token
}

func (n fnCallNode) String() string {
	// TODO: incorporate restArg
	argStrings := make([]string, len(n.args))
	for i, arg := range n.args {
		argStrings[i] = arg.String()
	}
	return fmt.Sprintf("call[%s](%s)", n.fn, strings.Join(argStrings, ", "))
}
func (n fnCallNode) pos() pos {
	return n.tok.pos
}

type ifBranch struct {
	target astNode
	body   astNode
}

func (n ifBranch) String() string {
	return n.target.String() + " -> " + n.body.String()
}

type ifExprNode struct {
	cond     astNode
	branches []ifBranch
	tok      *token
}

func (n ifExprNode) String() string {
	branchStrings := make([]string, len(n.branches))
	for i, br := range n.branches {
		branchStrings[i] = br.String()
	}
	return "if " + n.cond.String() + " {" + strings.Join(branchStrings, ", ") + "}"
}
func (n ifExprNode) pos() pos {
	return n.tok.pos
}

type blockNode struct {
	exprs []astNode
	tok   *token
}

func (n blockNode) String() string {
	exprStrings := make([]string, len(n.exprs))
	for i, ex := range n.exprs {
		exprStrings[i] = ex.String()
	}
	return "{ " + strings.Join(exprStrings, ", ") + " }"
}

func (n blockNode) pos() pos {
	return n.tok.pos
}

type parser struct {
	tokens        []token
	index         int
	minBinaryPrec []int
}

func newParser(tokens []token) parser {
	return parser{
		tokens:        tokens,
		index:         0,
		minBinaryPrec: []int{0},
	}
}

func (p *parser) lastMinPrec() int {
	return p.minBinaryPrec[len(p.minBinaryPrec)-1]
}

func (p *parser) pushMinPrec(prec int) {
	p.minBinaryPrec = append(p.minBinaryPrec, prec)
}

func (p *parser) popMinPrec() {
	p.minBinaryPrec = p.minBinaryPrec[:len(p.minBinaryPrec)-1]
}

func (p *parser) isEOF() bool {
	return p.index == len(p.tokens)
}

func (p *parser) peek() token {
	return p.tokens[p.index]
}

func (p *parser) peekAhead(n int) token {
	if p.index+n > len(p.tokens) {
		// Use comma as "nothing is here" value
		return token{kind: comma}
	}
	return p.tokens[p.index+n]
}

func (p *parser) next() token {
	tok := p.tokens[p.index]

	if p.index < len(p.tokens) {
		p.index++
	}

	return tok
}

func (p *parser) back() {
	if p.index > 0 {
		p.index--
	}
}

func (p *parser) expect(kind tokKind) (token, error) {
	tok := token{kind: kind}

	if p.isEOF() {
		return token{kind: unknown}, parseError{
			reason: fmt.Sprintf("Unexpected end of input, expected %s", tok),
			pos:    tok.pos,
		}
	}

	next := p.next()
	if next.kind != kind {
		return token{kind: unknown}, parseError{
			reason: fmt.Sprintf("Unexpected token %s, expected %s", next, tok),
			pos:    tok.pos,
		}
	}

	return next, nil
}

func (p *parser) readUntilTokenKind(kind tokKind) []token {
	tokens := []token{}
	for !p.isEOF() && p.peek().kind != kind {
		tokens = append(tokens, p.next())
	}
	return tokens
}

// concrete astNode parse functions

type parseError struct {
	reason string
	pos
}

func (e parseError) Error() string {
	return fmt.Sprintf("Parse error at %s: %s", e.pos.String(), e.reason)
}

func (p *parser) parseAssignment(left astNode) (astNode, error) {
	if p.peek().kind != assign &&
		p.peek().kind != nonlocalAssign {
		return left, nil
	}

	next := p.next()
	node := assignmentNode{
		isLocal: next.kind == assign,
		left:    left,
		tok:     &next,
	}

	right, err := p.parseNode()
	if err != nil {
		return nil, err
	}
	node.right = right

	return node, nil
}

// parseUnit is responsible for parsing the smallest complete syntactic "units"
// of Oak's syntax, like literals including function literals, grouped
// expressions in blocks, and if/with expressions.
func (p *parser) parseUnit() (astNode, error) {
	tok := p.next()
	switch tok.kind {
	case qmark:
		return nullNode{tok: &tok}, nil
	case stringLiteral:
		return stringNode{payload: tok.payload, tok: &tok}, nil
	case numberLiteral:
		if strings.ContainsRune(tok.payload, '.') {
			f, err := strconv.ParseFloat(tok.payload, 64)
			if err != nil {
				return nil, parseError{reason: err.Error(), pos: tok.pos}
			}
			return numberNode{
				isInteger:    false,
				floatPayload: f,
				tok:          &tok,
			}, nil
		}
		n, err := strconv.ParseInt(tok.payload, 10, 64)
		if err != nil {
			return nil, parseError{reason: err.Error(), pos: tok.pos}
		}
		return numberNode{
			isInteger:  true,
			intPayload: n,
			tok:        &tok,
		}, nil
	case trueLiteral:
		return booleanNode{payload: true, tok: &tok}, nil
	case falseLiteral:
		return booleanNode{payload: false, tok: &tok}, nil
	case colon:
		if p.peek().kind == identifier {
			return atomNode{payload: p.next().payload, tok: &tok}, nil
		}
		// TODO: let keywords be valid atoms
		return nil, parseError{
			reason: fmt.Sprintf("Expected identifier after ':', got %s", p.peek()),
			pos:    tok.pos,
		}
	case leftBracket:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		itemNodes := []astNode{}
		for !p.isEOF() && p.peek().kind != rightBracket {
			node, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			itemNodes = append(itemNodes, node)
		}
		if _, err := p.expect(rightBracket); err != nil {
			return nil, err
		}

		return listNode{elems: itemNodes, tok: &tok}, nil
	case leftBrace:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		// empty {} is always considered an object -- an empty block is illegal
		if p.peek().kind == rightBrace {
			p.next() // eat the rightBrace
			return objectNode{entries: []objectEntry{}, tok: &tok}, nil
		}

		firstExpr, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if p.isEOF() {
			return nil, parseError{
				reason: fmt.Sprintf("Unexpected end of input inside block or object"),
				pos:    tok.pos,
			}
		}

		if p.peek().kind == colon {
			// it's an object
			p.next() // eat the colon
			valExpr, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			entries := []objectEntry{
				{key: firstExpr, val: valExpr},
			}

			for !p.isEOF() && p.peek().kind != rightBrace {
				key, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(colon); err != nil {
					return nil, err
				}

				val, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(comma); err != nil {
					return nil, err
				}

				entries = append(entries, objectEntry{
					key: key,
					val: val,
				})
			}
			if _, err := p.expect(rightBrace); err != nil {
				return nil, err
			}

			return objectNode{entries: entries, tok: &tok}, nil
		}

		// it's a block
		exprs := []astNode{firstExpr}
		if _, err := p.expect(comma); err != nil {
			return nil, err
		}

		for !p.isEOF() && p.peek().kind != rightBrace {
			expr, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			exprs = append(exprs, expr)
		}
		if _, err := p.expect(rightBrace); err != nil {
			return nil, err
		}

		return blockNode{exprs: exprs, tok: &tok}, nil
	case fnKeyword:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		name := ""
		if p.peek().kind == identifier {
			// optional named fn
			name = p.next().payload
		}

		args := []string{}
		var restArg string
		if p.peek().kind == leftParen {
			// optional argument list
			p.next() // eat the leftParen
			for !p.isEOF() && p.peek().kind != rightParen {
				arg, err := p.expect(identifier)
				if err != nil {
					p.back() // try again

					_, err := p.expect(underscore)
					if err != nil {
						return nil, err
					}

					args = append(args, "")

					if _, err := p.expect(comma); err != nil {
						return nil, err
					}

					continue
				}

				// maybe this is a rest arg
				if p.peek().kind == ellipsis {
					restArg = arg.payload
					p.next() // eat the ellipsis

					_, err = p.expect(comma)
					if err != nil {
						return nil, err
					}
					break
				}

				args = append(args, arg.payload)

				if _, err := p.expect(comma); err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(rightParen); err != nil {
				return nil, err
			}
		}

		// Exception to the "{} is empty object" rule is that `fn {}` parses as
		// a function with an empty block as a body
		var body astNode
		var err error
		if p.peek().kind == leftBrace && p.peekAhead(1).kind == rightBrace {
			blockStartTok := p.next()
			p.next()
			body = blockNode{exprs: []astNode{}, tok: &blockStartTok}
		} else {
			body, err = p.parseNode()
			if err != nil {
				return nil, err
			}
		}

		return fnNode{
			name:    name,
			args:    args,
			restArg: restArg,
			body:    body,
			tok:     &tok,
		}, nil
	case underscore:
		return emptyNode{tok: &tok}, nil
	case identifier:
		return identifierNode{payload: tok.payload, tok: &tok}, nil
	case minus, exclam:
		right, err := p.parseSubNode()
		if err != nil {
			return nil, err
		}

		return unaryNode{
			op:    tok.kind,
			right: right,
			tok:   &tok,
		}, nil
	case ifKeyword:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		condNode, err := p.parseNode()
		if err != nil {
			return nil, err
		}

		if _, err = p.expect(leftBrace); err != nil {
			return nil, err
		}

		branches := []ifBranch{}
		for !p.isEOF() && p.peek().kind != rightBrace {
			targets := []astNode{}
			for !p.isEOF() && p.peek().kind != branchArrow {
				target, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if p.peek().kind != branchArrow {
					if _, err := p.expect(comma); err != nil {
						return nil, err
					}
				}

				targets = append(targets, target)
			}
			if _, err := p.expect(branchArrow); err != nil {
				return nil, err
			}

			body, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			// We want to support multi-target branches, but don't want to
			// incur the performance overhead in the interpreter/evaluator of
			// keeping every single target as a Go slice, when the vast
			// majority of targets will be single-value, which requires just a
			// pointer to an astNode.
			//
			// So instead of doing that, we penalize the multi-value case by
			// essentially considering it syntax sugar and splitting such
			// branches into multiple AST branches, each with one target value.
			for _, target := range targets {
				branches = append(branches, ifBranch{
					target: target,
					body:   body,
				})
			}
		}
		if _, err := p.expect(rightBrace); err != nil {
			return nil, err
		}

		return ifExprNode{
			cond:     condNode,
			branches: branches,
			tok:      &tok,
		}, nil
	case withKeyword:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		withExprBase, err := p.parseNode()
		if err != nil {
			return nil, err
		}

		withExprBaseCall, ok := withExprBase.(fnCallNode)
		if !ok {
			return nil, parseError{
				reason: fmt.Sprintf("with keyword should be followed by a function call, found %s", withExprBase),
				pos:    tok.pos,
			}
		}

		withExprLastArg, err := p.parseNode()
		if err != nil {
			return nil, err
		}

		withExprBaseCall.args = append(withExprBaseCall.args, withExprLastArg)
		return withExprBaseCall, nil
	case leftParen:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		exprs := []astNode{}
		for !p.isEOF() && p.peek().kind != rightParen {
			expr, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			exprs = append(exprs, expr)
		}
		if _, err := p.expect(rightParen); err != nil {
			return nil, err
		}
		return blockNode{exprs: exprs, tok: &tok}, nil
	}
	return nil, parseError{
		reason: fmt.Sprintf("Unexpected token %s at start of unit", tok),
		pos:    tok.pos,
	}
}

func infixOpPrecedence(op tokKind) int {
	switch op {
	case plus, minus:
		return 40
	case times, divide:
		return 50
	case modulus:
		return 80
	case eq, greater, less, geq, leq, neq:
		return 30
	case and:
		return 20
	case xor:
		return 15
	case or:
		return 10
	case pushArrow:
		// assignment-like semantics
		return 1
	default:
		return -1
	}
}

// parseSubNode is responsible for parsing independent "terms" in the Oak
// syntax, like terms in unary and binary expressions and in pipelines. It is
// in between parseUnit and parseNode.
func (p *parser) parseSubNode() (astNode, error) {
	p.pushMinPrec(0)
	defer p.popMinPrec()

	node, err := p.parseUnit()
	if err != nil {
		return nil, err
	}

	for !p.isEOF() {
		switch p.peek().kind {
		case dot:
			next := p.next() // eat the dot
			right, err := p.parseUnit()
			if err != nil {
				return nil, err
			}

			node = propertyAccessNode{
				left:  node,
				right: right,
				tok:   &next,
			}
		case leftParen:
			next := p.next() // eat the leftParen

			args := []astNode{}
			var restArg astNode = nil
			for !p.isEOF() && p.peek().kind != rightParen {
				arg, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if p.peek().kind == ellipsis {
					p.next() // eat the ellipsis

					if _, err = p.expect(comma); err != nil {
						return nil, err
					}

					restArg = arg

					break
				} else {
					args = append(args, arg)
				}

				if _, err = p.expect(comma); err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(rightParen); err != nil {
				return nil, err
			}

			node = fnCallNode{
				fn:      node,
				args:    args,
				restArg: restArg,
				tok:     &next,
			}
		default:
			return node, nil
		}
	}

	return node, nil
}

// parseNode returns the next top-level astNode from the parser
func (p *parser) parseNode() (astNode, error) {
	node, err := p.parseSubNode()
	if err != nil {
		return nil, err
	}

	for !p.isEOF() && p.peek().kind != comma {
		switch p.peek().kind {
		case assign, nonlocalAssign:
			// whatever follows an assignment expr cannot bind to the
			// assignment expression itself by syntax rule, so we simply return
			return p.parseAssignment(node)
		case plus, minus, times, divide, modulus,
			xor, and, or, pushArrow,
			greater, less, eq, geq, leq, neq:
			// this case implements a mini Pratt parser threaded through the
			// larger Oak syntax parser, using the parser struct itself to keep
			// track of the power / precedence stack since other forms may be
			// parsed in between, as in 1 + f(g(x := y)) + 2
			minPrec := p.lastMinPrec()

			for {
				if p.isEOF() {
					return nil, parseError{
						reason: "Incomplete binary expression",
						pos:    p.peek().pos,
					}
				}

				peeked := p.peek()
				op := peeked.kind
				prec := infixOpPrecedence(op)
				if prec <= minPrec {
					break
				}
				p.next() // eat the operator

				if p.isEOF() {
					return nil, parseError{
						reason: fmt.Sprintf("Incomplete binary expression with %s", token{kind: op}),
						pos:    p.peek().pos,
					}
				}

				p.pushMinPrec(prec)
				right, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				p.popMinPrec()

				node = binaryNode{
					op:    op,
					left:  node,
					right: right,
					tok:   &peeked,
				}
			}

			// whatever follows a binary expr cannot bind to the binary
			// expression by syntax rule, so we simply return
			return node, nil
		case pipeArrow:
			pipe := p.next() // eat the pipe

			pipeRight, err := p.parseSubNode()
			if err != nil {
				return nil, err
			}
			pipedFnCall, ok := pipeRight.(fnCallNode)
			if !ok {
				return nil, parseError{
					reason: fmt.Sprintf("Expected function call after |>, got %s", pipeRight),
					pos:    pipe.pos,
				}
			}

			pipedFnCall.args = append([]astNode{node}, pipedFnCall.args...)
			node = pipedFnCall
		default:
			return node, nil
		}
	}
	// the trailing comma is handled as necessary in callers of parseNode

	return node, nil
}

func (p *parser) parse() ([]astNode, error) {
	nodes := []astNode{}

	for !p.isEOF() {
		node, err := p.parseNode()
		if err != nil {
			return nodes, err
		}

		if _, err = p.expect(comma); err != nil {
			return nodes, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
