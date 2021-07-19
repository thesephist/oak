package main

import (
	"fmt"
	"strconv"
	"strings"
)

type astNode interface {
	String() string
}

type emptyNode struct{}

func (n emptyNode) String() string {
	return "_"
}

type nullNode struct{}

func (n nullNode) String() string {
	return "?"
}

type stringNode struct {
	payload string
}

func (n stringNode) String() string {
	return fmt.Sprintf("%s", strconv.Quote(n.payload))
}

type numberNode struct {
	isInteger    bool
	intPayload   int64
	floatPayload float64
}

func (n numberNode) String() string {
	if n.isInteger {
		return strconv.FormatInt(n.intPayload, 10)
	}
	return strconv.FormatFloat(n.floatPayload, 'g', -1, 64)
}

type booleanNode struct {
	payload bool
}

func (n booleanNode) String() string {
	if n.payload {
		return "true"
	}
	return "false"
}

type atomNode struct {
	payload string
}

func (n atomNode) String() string {
	return ":" + n.payload
}

type listNode struct {
	elems []astNode
}

func (n listNode) String() string {
	elemStrings := make([]string, len(n.elems))
	for i, el := range n.elems {
		elemStrings[i] = el.String()
	}
	return "[" + strings.Join(elemStrings, ", ") + "]"
}

type objectEntryNode struct {
	key astNode
	val astNode
}

func (n objectEntryNode) String() string {
	return n.key.String() + ": " + n.val.String()
}

type objectNode struct {
	entries []objectEntryNode
}

func (n objectNode) String() string {
	entryStrings := make([]string, len(n.entries))
	for i, ent := range n.entries {
		entryStrings[i] = ent.String()
	}
	return "{ " + strings.Join(entryStrings, ", ") + " }"
}

type fnNode struct {
	name    string // "" for anonymous fns
	args    []string
	restArg string
	body    astNode
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

type identifierNode struct {
	payload string
}

func (n identifierNode) String() string {
	return n.payload
}

type assignmentNode struct {
	isLocal bool
	left    astNode
	right   astNode
}

func (n assignmentNode) String() string {
	if n.isLocal {
		return n.left.String() + " := " + n.right.String()
	}
	return n.left.String() + " <- " + n.right.String()
}

type propertyAccessNode struct {
	left  astNode
	right astNode
}

func (n propertyAccessNode) String() string {
	return "(" + n.left.String() + "." + n.right.String() + ")"
}

type unaryNode struct {
	op    tokKind
	right astNode
}

func (n unaryNode) String() string {
	opTok := token{kind: n.op}
	return opTok.String() + n.right.String()
}

type binaryNode struct {
	op    tokKind
	left  astNode
	right astNode
}

func (n binaryNode) String() string {
	opTok := token{kind: n.op}
	return "(" + n.left.String() + " " + opTok.String() + " " + n.right.String() + ")"
}

type fnCallNode struct {
	fn   astNode
	args []astNode
}

func (n fnCallNode) String() string {
	argStrings := make([]string, len(n.args))
	for i, arg := range n.args {
		argStrings[i] = arg.String()
	}
	return fmt.Sprintf("call[%s](%s)", n.fn, strings.Join(argStrings, ", "))
}

type ifBranchNode struct {
	target astNode
	body   astNode
}

func (n ifBranchNode) String() string {
	return n.target.String() + " -> " + n.body.String()
}

type ifExprNode struct {
	cond     astNode
	branches []ifBranchNode
}

func (n ifExprNode) String() string {
	branchStrings := make([]string, len(n.branches))
	for i, br := range n.branches {
		branchStrings[i] = br.String()
	}
	return "if " + n.cond.String() + " {" + strings.Join(branchStrings, ", ") + "}"
}

type blockNode struct {
	exprs []astNode
}

func (n blockNode) String() string {
	exprStrings := make([]string, len(n.exprs))
	for i, ex := range n.exprs {
		exprStrings[i] = ex.String()
	}
	return "{ " + strings.Join(exprStrings, ", ") + " }"
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
		}
	}

	next := p.next()
	if next.kind != kind {
		return token{kind: unknown}, parseError{
			reason: fmt.Sprintf("Unexpected token %s, expected %s", next, tok),
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

	node := assignmentNode{
		isLocal: p.next().kind == assign,
		left:    left,
	}

	right, err := p.nextNode()
	if err != nil {
		return nil, err
	}
	node.right = right

	return node, nil
}

// parseUnit is responsible for parsing the smallest complete syntactic "units"
// of Magnolia's syntax, like literals including function literals, grouped
// expressions in blocks, and if/with expressions.
func (p *parser) parseUnit() (astNode, error) {
	tok := p.next()
	switch tok.kind {
	case qmark:
		return nullNode{}, nil
	case stringLiteral:
		return stringNode{payload: tok.payload}, nil
	case numberLiteral:
		if strings.ContainsRune(tok.payload, '.') {
			f, err := strconv.ParseFloat(tok.payload, 64)
			if err != nil {
				return nil, parseError{reason: err.Error()}
			}
			return numberNode{
				isInteger:    false,
				floatPayload: f,
			}, nil
		}
		n, err := strconv.ParseInt(tok.payload, 10, 64)
		if err != nil {
			return nil, parseError{reason: err.Error()}
		}
		return numberNode{
			isInteger:  true,
			intPayload: n,
		}, nil
	case trueLiteral:
		return booleanNode{payload: true}, nil
	case falseLiteral:
		return booleanNode{payload: false}, nil
	case colon:
		if p.peek().kind == identifier {
			return atomNode{payload: p.next().payload}, nil
		}
		return nil, parseError{
			reason: fmt.Sprintf("Expected identifier after ':', got %s", p.peek()),
		}
	case leftBracket:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		itemNodes := []astNode{}
		for !p.isEOF() && p.peek().kind != rightBracket {
			node, err := p.nextNode()
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

		return listNode{elems: itemNodes}, nil
	case leftBrace:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		// empty {} is always considered an object -- an empty block is illegal
		if p.peek().kind == rightBrace {
			p.next() // eat the rightBrace
			return objectNode{entries: []objectEntryNode{}}, nil
		}

		firstExpr, err := p.nextNode()
		if err != nil {
			return nil, err
		}
		if p.isEOF() {
			return nil, parseError{
				reason: fmt.Sprintf("Unexpected end of input inside block or object"),
			}
		}

		if p.peek().kind == colon {
			// it's an object
			p.next() // eat the colon
			valExpr, err := p.nextNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			entries := []objectEntryNode{
				{key: firstExpr, val: valExpr},
			}

			for !p.isEOF() && p.peek().kind != rightBrace {
				key, err := p.nextNode()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(colon); err != nil {
					return nil, err
				}

				val, err := p.nextNode()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(comma); err != nil {
					return nil, err
				}

				entries = append(entries, objectEntryNode{
					key: key,
					val: val,
				})
			}
			if _, err := p.expect(rightBrace); err != nil {
				return nil, err
			}

			return objectNode{entries: entries}, nil
		}

		// it's a block
		exprs := []astNode{firstExpr}
		if _, err := p.expect(comma); err != nil {
			return nil, err
		}

		for !p.isEOF() && p.peek().kind != rightBrace {
			expr, err := p.nextNode()
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

		return blockNode{exprs: exprs}, nil
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
					return nil, err
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
			p.next()
			p.next()
			body = blockNode{exprs: []astNode{}}
		} else {
			body, err = p.nextNode()
			if err != nil {
				return nil, err
			}
		}

		return fnNode{
			name:    name,
			args:    args,
			restArg: restArg,
			body:    body,
		}, nil
	case underscore:
		return emptyNode{}, nil
	case identifier:
		return identifierNode{payload: tok.payload}, nil
	case minus, exclam:
		right, err := p.parseUnit()
		if err != nil {
			return nil, err
		}
		return unaryNode{
			op:    tok.kind,
			right: right,
		}, nil
	case ifKeyword:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		condNode, err := p.nextNode()
		if err != nil {
			return nil, err
		}

		if _, err = p.expect(leftBrace); err != nil {
			return nil, err
		}

		branches := []ifBranchNode{}
		for !p.isEOF() && p.peek().kind != rightBrace {
			target, err := p.nextNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(branchArrow); err != nil {
				return nil, err
			}

			body, err := p.nextNode()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(comma); err != nil {
				return nil, err
			}

			branches = append(branches, ifBranchNode{
				target: target,
				body:   body,
			})
		}
		if _, err := p.expect(rightBrace); err != nil {
			return nil, err
		}

		return ifExprNode{
			cond:     condNode,
			branches: branches,
		}, nil
	case withKeyword:
		p.pushMinPrec(0)
		defer p.popMinPrec()

		withExprBase, err := p.nextNode()
		if err != nil {
			return nil, err
		}

		withExprBaseCall, ok := withExprBase.(fnCallNode)
		if !ok {
			return nil, parseError{
				reason: fmt.Sprintf("with keyword should be followed by a function call, found %s", withExprBase),
			}
		}

		withExprLastArg, err := p.nextNode()
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
			expr, err := p.nextNode()
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
		return blockNode{exprs: exprs}, nil
	}
	return nil, parseError{
		reason: fmt.Sprintf("Unexpected token %s", tok),
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
	default:
		return -1
	}
}

func (p *parser) nextNode() (astNode, error) {
	return p.nextNodeInPipe(false)
}

func (p *parser) nextNodeInPipe(inPipe bool) (astNode, error) {
	node, err := p.parseUnit()
	if err != nil {
		return nil, err
	}

	for !p.isEOF() && p.peek().kind != comma {
		switch p.peek().kind {
		case dot:
			p.next() // eat the dot
			right, err := p.parseUnit()
			if err != nil {
				return nil, err
			}

			node = propertyAccessNode{
				left:  node,
				right: right,
			}
		case leftParen:
			p.next() // eat the leftParen

			args := []astNode{}
			for !p.isEOF() && p.peek().kind != rightParen {
				arg, err := p.nextNode()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)

				if _, err = p.expect(comma); err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(rightParen); err != nil {
				return nil, err
			}

			node = fnCallNode{
				fn:   node,
				args: args,
			}
		case assign, nonlocalAssign:
			// whatever follows an assignment expr cannot bind to the
			// assignment expression itself by syntax rule, so we simply return
			return p.parseAssignment(node)
		case plus, minus, times, divide, modulus,
			xor, and, or,
			greater, less, eq, geq, leq, neq:
			// this case implements a mini Pratt parser threaded through the
			// larger Magnolia syntax parser, using the parser struct itself to
			// keep track of the power / precedence stack since other forms may
			// be parsed in between, as in 1 + f(g(x := y)) + 2
			minPrec := p.lastMinPrec()

			for {
				if p.isEOF() {
					return nil, parseError{
						reason: "Incomplete binary expression",
					}
				}

				op := p.peek().kind
				prec := infixOpPrecedence(op)
				if prec <= minPrec {
					break
				}
				p.next() // eat the operator

				if p.isEOF() {
					return nil, parseError{
						reason: fmt.Sprintf("Incomplete binary expression with %s", token{kind: op}),
					}
				}

				p.pushMinPrec(prec)
				right, err := p.nextNode()
				if err != nil {
					return nil, err
				}
				p.popMinPrec()

				node = binaryNode{
					op:    op,
					left:  node,
					right: right,
				}
			}

			// whatever follows a binary expr cannot bind to the binary
			// expression by syntax rule, so we simply return
			return node, nil
		case pipeArrow:
			if inPipe {
				return node, nil
			}

			p.next() // eat the pipe

			pipeRight, err := p.nextNodeInPipe(true)
			if err != nil {
				return nil, err
			}
			// above is guaranteed to return a fnCallNode
			pipedFnCall, _ := pipeRight.(fnCallNode)

			pipedFnCall.args = append([]astNode{node}, pipedFnCall.args...)
			node = pipedFnCall
		default:
			return node, nil
		}
	}
	// the trailing comma is handled as necessary in callers of nextNode

	return node, nil
}

func (p *parser) parse() ([]astNode, error) {
	nodes := []astNode{}

	for !p.isEOF() {
		node, err := p.nextNode()
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
