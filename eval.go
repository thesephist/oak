package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// byte slice helpers from the Ink interpreter source code,
// github.com/thesephist/ink

// zero-extend a slice of bytes to given length
func zeroExtend(s []byte, max int) []byte {
	if max <= len(s) {
		return s
	}

	extended := make([]byte, max)
	copy(extended, s)
	return extended
}

// return the max length of two slices
func maxLen(a, b []byte) int {
	if alen, blen := len(a), len(b); alen < blen {
		return blen
	} else {
		return alen
	}
}

type Value interface {
	String() string
	Eq(Value) bool
}

type EmptyValue byte

// interned "empty" value
const empty EmptyValue = 0

func (v EmptyValue) String() string {
	return "_"
}
func (v EmptyValue) Eq(u Value) bool {
	return true
}

// Null need not contain any data, so we use the most compact data
// representation we can.
type NullValue byte

// interned "null"
const null NullValue = 0

func (v NullValue) String() string {
	return "?"
}
func (v NullValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if _, ok := u.(NullValue); ok {
		return true
	}
	return false
}

type StringValue []byte

var emptyString = StringValue("")

func (v StringValue) String() string {
	return fmt.Sprintf("'%s'", string(v))
}
func (v StringValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(StringValue); ok {
		return bytes.Equal(v, w)
	}
	return false
}

type IntValue int64

func (v IntValue) String() string {
	return strconv.FormatInt(int64(v), 10)
}
func (v IntValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(IntValue); ok {
		return v == w
	}
	return false
}

type FloatValue float64

func (v FloatValue) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}
func (v FloatValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(FloatValue); ok {
		return v == w
	}

	return false
}

type BoolValue bool

// interned booleans
const mgnTrue = BoolValue(true)
const mgnFalse = BoolValue(false)

func (v BoolValue) String() string {
	if v {
		return "true"
	}
	return "false"
}
func (v BoolValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(BoolValue); ok {
		return v == w
	}

	return false
}

type AtomValue string

func (v AtomValue) String() string {
	return ":" + string(v)
}
func (v AtomValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(AtomValue); ok {
		return v == w
	}

	return false
}

type ListValue []Value

func (v ListValue) String() string {
	valStrings := make([]string, len(v))
	for i, val := range v {
		valStrings[i] = val.String()
	}
	return "[" + strings.Join(valStrings, ", ") + "]"
}
func (v ListValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(ListValue); ok {
		if len(v) != len(w) {
			return false
		}

		for i, el := range v {
			if !el.Eq(w[i]) {
				return false
			}
		}
		return true
	}

	return false
}

type ObjectValue map[string]Value

func (v ObjectValue) String() string {
	// TODO: fix how this deals with circular references
	entryStrings := make([]string, len(v))
	i := 0
	for key, val := range v {
		entryStrings[i] = key + ": " + val.String()
		i++
	}
	return "{" + strings.Join(entryStrings, ", ") + "}"
}
func (v ObjectValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(ObjectValue); ok {
		if len(v) != len(w) {
			return false
		}

		for key, val := range v {
			if wVal, ok := w[key]; ok {
				if !val.Eq(wVal) {
					return false
				}
			} else {
				return false
			}
		}

		return true
	}

	return false
}

type FnValue struct {
	defn *fnNode
	scope
}

func (v FnValue) String() string {
	return v.defn.String()
}
func (v FnValue) Eq(u Value) bool {
	if w, ok := u.(FnValue); ok {
		return v.defn == w.defn
	}

	return false
}

type scope struct {
	parent *scope
	vars   map[string]Value
}

func (sc *scope) get(name string) (Value, error) {
	if v, ok := sc.vars[name]; ok {
		return v, nil
	}
	if sc.parent != nil {
		return sc.parent.get(name)
	}
	return nil, runtimeError{
		reason: fmt.Sprintf("%s is undefined", name),
	}
}

func (sc *scope) put(name string, v Value) {
	sc.vars[name] = v
}

func (sc *scope) update(name string, v Value) error {
	if _, ok := sc.vars[name]; ok {
		sc.vars[name] = v
		return nil
	}
	if sc.parent != nil {
		return sc.parent.update(name, v)
	}
	return runtimeError{
		reason: fmt.Sprintf("%s is undefined", name),
	}
}

type Context struct {
	// current working directory of this context, used for loading other
	// modules with relative paths / URLs
	Cwd string
	// path or descriptor of the file being run, used for error reporting
	SourcePath string
	// top level ("global") scope of this context
	scope
}

func NewContext(path, cwd string) Context {
	return Context{
		Cwd:        cwd,
		SourcePath: path,
		scope: scope{
			parent: nil,
			vars:   map[string]Value{},
		},
	}
}

func (c *Context) generateStackTrace() stackEntry {
	// TODO: actually write
	return stackEntry{}
}

type stackEntry struct {
	fnName      string
	parentStack *stackEntry
	pos
}

type vmError struct {
	reason string
}

func (e vmError) Error() string {
	return fmt.Sprintf("VM error: %s", e.reason)
}

type runtimeError struct {
	reason     string
	stackTrace stackEntry
}

func (e runtimeError) Error() string {
	// TODO: display stacktrace
	return fmt.Sprintf("Runtime error: %s", e.reason)
}

func (c *Context) Eval(programReader io.Reader) (Value, error) {
	program, err := io.ReadAll(programReader)
	if err != nil {
		return nil, err
	}

	tokenizer := newTokenizer(string(program))
	tokens := tokenizer.tokenize()
	// fmt.Println(tokens)

	parser := newParser(tokens)
	nodes, err := parser.parse()
	if err != nil {
		return nil, err
	}
	// fmt.Println(nodes)

	return c.evalNodes(nodes)
}

func (c *Context) evalNodes(nodes []astNode) (Value, error) {
	var err error
	var returnVal Value = null
	for _, expr := range nodes {
		returnVal, err = c.evalExpr(expr, c.scope)
		if err != nil {
			return nil, err
		}
	}
	return returnVal, nil
}

func intBinaryOp(op tokKind, left, right IntValue) (Value, error) {
	switch op {
	case plus:
		return IntValue(int64(left) + int64(right)), nil
	case minus:
		return IntValue(int64(left) - int64(right)), nil
	case times:
		return IntValue(int64(left) * int64(right)), nil
	case divide:
		return IntValue(int64(left) / int64(right)), nil
	case modulus:
		return IntValue(int64(left) % int64(right)), nil
	case xor:
		return IntValue(int64(left) ^ int64(right)), nil
	case and:
		return IntValue(int64(left) & int64(right)), nil
	case or:
		return IntValue(int64(left) | int64(right)), nil
	case greater:
		return BoolValue(int64(left) > int64(right)), nil
	case less:
		return BoolValue(int64(left) < int64(right)), nil
	case geq:
		return BoolValue(int64(left) >= int64(right)), nil
	case leq:
		return BoolValue(int64(left) <= int64(right)), nil
	}
	panic(fmt.Sprintf("Invalid binary operator %s", token{kind: op}))
}

func floatBinaryOp(op tokKind, left, right FloatValue) (Value, error) {
	switch op {
	case plus:
		return FloatValue(float64(left) + float64(right)), nil
	case minus:
		return FloatValue(float64(left) - float64(right)), nil
	case times:
		return FloatValue(float64(left) * float64(right)), nil
	case divide:
		return FloatValue(float64(left) / float64(right)), nil
	case modulus:
		return FloatValue(math.Mod(float64(left), float64(right))), nil
	case greater:
		return BoolValue(int64(left) > int64(right)), nil
	case less:
		return BoolValue(int64(left) < int64(right)), nil
	case geq:
		return BoolValue(int64(left) >= int64(right)), nil
	case leq:
		return BoolValue(int64(left) <= int64(right)), nil
	}
	panic(fmt.Sprintf("Invalid binary operator %s", token{kind: op}))
}

func (c *Context) evalExpr(node astNode, sc scope) (Value, error) {
	switch n := node.(type) {
	case emptyNode:
		return empty, nil
	case nullNode:
		return null, nil
	case stringNode:
		return StringValue([]byte(n.payload)), nil
	case numberNode:
		if n.isInteger {
			return IntValue(n.intPayload), nil
		}
		return FloatValue(n.floatPayload), nil
	case booleanNode:
		return BoolValue(n.payload), nil
	case atomNode:
		return AtomValue(n.payload), nil
	case listNode:
		var err error
		elems := make([]Value, len(n.elems))
		for i, elNode := range n.elems {
			elems[i], err = c.evalExpr(elNode, sc)
			if err != nil {
				return nil, err
			}
		}
		return ListValue(elems), nil
	case objectNode:
		obj := ObjectValue{}
		for _, entry := range n.entries {
			var keyString string

			if identKey, ok := entry.key.(identifierNode); ok {
				keyString = identKey.payload
			} else {
				key, err := c.evalExpr(entry.key, sc)
				if err != nil {
					return nil, err
				}
				switch typedKey := key.(type) {
				case StringValue:
					keyString = string(typedKey)
				case IntValue:
					keyString = typedKey.String()
				case FloatValue:
					keyString = typedKey.String()
				default:
					return nil, runtimeError{
						reason: fmt.Sprintf("Expected a string or number as object key, got %s", key.String()),
					}
				}
			}

			val, err := c.evalExpr(entry.val, sc)
			if err != nil {
				return nil, err
			}

			obj[keyString] = val
		}
		return obj, nil
	case fnNode:
		fn := FnValue{
			defn:  &n,
			scope: sc,
		}
		if fn.defn.name != "" {
			sc.put(fn.defn.name, fn)
		}
		return fn, nil
	case identifierNode:
		return sc.get(n.payload)
	case assignmentNode:
		assignedValue, err := c.evalExpr(n.right, sc)
		if err != nil {
			return nil, err
		}
		switch left := n.left.(type) {
		case identifierNode:
			if n.isLocal {
				sc.put(left.payload, assignedValue)
			} else {
				err := sc.update(left.payload, assignedValue)
				if err != nil {
					return nil, err
				}
			}
			return assignedValue, nil
		case listNode:
			// TODO: implement list destructuring assignment
			panic("list destructuring not implemented!")
		case objectNode:
			// TODO: implement object destructuring assignment
			panic("object destructuring not implemented!")
		case propertyAccessNode:
			// TODO: implement object property assignment
			panic("assign to property not implemented!")
		}
		panic(fmt.Sprintf("Illegal left-hand side of assignment in %s", n))
	case propertyAccessNode:
		left, err := c.evalExpr(n.left, sc)
		if err != nil {
			return nil, err
		}

		var right Value
		if rightIdent, ok := n.right.(identifierNode); ok {
			right = StringValue([]byte(rightIdent.payload))
		} else {
			var err error
			right, err = c.evalExpr(n.right, sc)
			if err != nil {
				return nil, err
			}
		}

		switch target := left.(type) {
		case StringValue:
			byteIndex, ok := right.(IntValue)
			if !ok {
				return nil, runtimeError{
					reason: fmt.Sprintf("Cannot index into string with non-integer index %s", right),
				}
			}

			if byteIndex < 0 || int64(byteIndex) > int64(len(target)) {
				return null, nil
			}

			return StringValue([]byte{target[byteIndex]}), nil
		case ListValue:
			listIndex, ok := right.(IntValue)
			if !ok {
				return nil, runtimeError{
					reason: fmt.Sprintf("Cannot index into list with non-integer index %s", right),
				}
			}

			if listIndex < 0 || int64(listIndex) > int64(len(target)) {
				return null, nil
			}

			return target[listIndex], nil
		case ObjectValue:
			var objKeyString string
			if objKey, ok := right.(StringValue); ok {
				objKeyString = string(objKey)
			} else {
				objKeyString = right.String()
			}

			if val, ok := target[objKeyString]; ok {
				return val, nil
			}

			return null, nil
		}

		return nil, runtimeError{
			reason: fmt.Sprintf("Expected string, list, or object in left-hand side of property access, got %s", left.String()),
		}
	case unaryNode:
		// TODO: implement
		panic("unaryNode not implemented!")
	case binaryNode:
		leftComputed, err := c.evalExpr(n.left, sc)
		if err != nil {
			return nil, err
		}

		rightComputed, err := c.evalExpr(n.right, sc)
		if err != nil {
			return nil, err
		}

		incompatibleError := runtimeError{
			reason: fmt.Sprintf("Cannot %s incompatible values %s, %s",
				token{kind: n.op}, leftComputed, rightComputed),
		}

		if n.op == eq {
			return BoolValue(leftComputed.Eq(rightComputed)), nil
		} else if n.op == neq {
			return BoolValue(!leftComputed.Eq(rightComputed)), nil
		}

		switch left := leftComputed.(type) {
		case IntValue:
			right, ok := rightComputed.(IntValue)
			if !ok {
				rightFloat, ok := rightComputed.(FloatValue)
				if !ok {
					return nil, incompatibleError
				}

				leftFloat := FloatValue(float64(int64(left)))
				return floatBinaryOp(n.op, leftFloat, rightFloat)
			}

			return intBinaryOp(n.op, left, right)
		case FloatValue:
			right, ok := rightComputed.(FloatValue)
			if !ok {
				rightInt, ok := rightComputed.(IntValue)
				if !ok {
					return nil, incompatibleError
				}

				leftInt := IntValue(math.Trunc(float64(left)))
				return intBinaryOp(n.op, leftInt, rightInt)
			}

			return floatBinaryOp(n.op, left, right)
		case StringValue:
			right, ok := rightComputed.(StringValue)
			if !ok {
				return nil, incompatibleError
			}

			switch n.op {
			case plus:
				base := make([]byte, 0, len(left)+len(right))
				base = append(base, left...)
				return StringValue(append(base, right...)), nil
			case xor:
				max := maxLen(left, right)

				ls, rs := zeroExtend(left, max), zeroExtend(right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] ^ rs[i]
				}
				return StringValue(res), nil
			case and:
				max := maxLen(left, right)

				ls, rs := zeroExtend(left, max), zeroExtend(right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] & rs[i]
				}
				return StringValue(res), nil
			case or:
				max := maxLen(left, right)

				ls, rs := zeroExtend(left, max), zeroExtend(right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] | rs[i]
				}
				return StringValue(res), nil
			case greater:
				return BoolValue(bytes.Compare(left, right) > 0), nil
			case less:
				return BoolValue(bytes.Compare(left, right) < 0), nil
			case geq:
				return BoolValue(bytes.Compare(left, right) >= 0), nil
			case leq:
				return BoolValue(bytes.Compare(left, right) <= 0), nil
			}
			panic(fmt.Sprintf("Invalid binary operator %s", token{kind: n.op}))
		case BoolValue:
			right, ok := rightComputed.(BoolValue)
			if !ok {
				return nil, incompatibleError
			}

			switch n.op {
			case plus, or:
				return BoolValue(left || right), nil
			case times, and:
				return BoolValue(left && right), nil
			case xor:
				return BoolValue(left != right), nil
			}
		}
		panic(fmt.Sprintf("Binary operator %s is not defined for values %s (%t), %s (%t)",
			token{kind: n.op}, leftComputed, leftComputed, rightComputed, rightComputed))
	case fnCallNode:
		maybeFn, err := c.evalExpr(n.fn, sc)
		if err != nil {
			return nil, err
		}

		args := make([]Value, len(n.args))
		for i, argNode := range n.args {
			args[i], err = c.evalExpr(argNode, sc)
			if err != nil {
				return nil, err
			}
		}

		if fn, ok := maybeFn.(FnValue); ok {
			// TODO: implement restArgs
			if len(args) < len(fn.defn.args) {
				// if not enough arguments, fill them with nulls
				difference := len(fn.defn.args) - len(args)
				extraArgs := make([]Value, difference)
				for i := 0; i < difference; i++ {
					extraArgs[i] = null
				}
				args = append(args, extraArgs...)
			} else {
				// if too many arguments, just slice to the right size
				args = args[:len(fn.defn.args)]
			}
			fnScope := scope{
				parent: &fn.scope,
				vars:   map[string]Value{},
			}
			for i, argName := range fn.defn.args {
				fnScope.put(argName, args[i])
			}
			return c.evalExpr(fn.defn.body, fnScope)
		} else if fn, ok := maybeFn.(BuiltinFnValue); ok {
			return fn.fn(args)
		} else {
			return nil, runtimeError{
				reason: fmt.Sprintf("%s (from %s) is not a function and cannot be called", maybeFn, n.fn),
			}
		}
	case ifExprNode:
		cond, err := c.evalExpr(n.cond, sc)
		if err != nil {
			return nil, err
		}

		for _, branch := range n.branches {
			target, err := c.evalExpr(branch.target, sc)
			if err != nil {
				return nil, err
			}

			if cond.Eq(target) {
				return c.evalExpr(branch.body, sc)
			}
		}
		return null, nil
	case blockNode:
		var err error
		blockScope := scope{
			parent: &sc,
			vars:   map[string]Value{},
		}

		// empty block returns ? (null)
		var returnVal Value = null
		for _, expr := range n.exprs {
			returnVal, err = c.evalExpr(expr, blockScope)
			if err != nil {
				return nil, err
			}
		}
		return returnVal, nil
	}
	return null, nil
}
