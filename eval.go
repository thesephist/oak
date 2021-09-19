package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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

func MakeString(s string) *StringValue {
	v := StringValue(s)
	return &v
}
func (v *StringValue) String() string {
	return fmt.Sprintf("'%s'", string(*v))
}
func (v *StringValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(*StringValue); ok {
		return bytes.Equal(*v, *w)
	}
	return false
}
func (v *StringValue) stringContent() string {
	return string(*v)
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
const oakTrue = BoolValue(true)
const oakFalse = BoolValue(false)

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

func MakeList(xs ...Value) *ListValue {
	v := ListValue(xs)
	return &v
}
func (v *ListValue) String() string {
	valStrings := make([]string, len(*v))
	for i, val := range *v {
		valStrings[i] = val.String()
	}
	return "[" + strings.Join(valStrings, ", ") + "]"
}
func (v *ListValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(*ListValue); ok {
		if len(*v) != len(*w) {
			return false
		}

		for i, el := range *v {
			if !el.Eq((*w)[i]) {
				return false
			}
		}
		return true
	}

	return false
}

type ObjectValue map[string]Value

// only used for efficient serialization to string
type serializedObjEntry struct {
	key  string
	full string
}

func (v ObjectValue) String() string {
	// TODO: fix how this deals with circular references
	entries := make([]serializedObjEntry, len(v))

	i := 0
	for key, val := range v {
		entries[i] = serializedObjEntry{
			key:  key,
			full: key + ": " + val.String(),
		}
		i++
	}

	// sort entries lexicographically for easier debugging use
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	sb := strings.Builder{}
	sb.WriteString("{")
	for i, entry := range entries {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(entry.full)
	}
	sb.WriteString("}")

	return sb.String()
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
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(FnValue); ok {
		return v.defn == w.defn
	}

	return false
}

type thunkValue struct {
	defn *fnNode
	scope
}

func (v thunkValue) String() string {
	return fmt.Sprintf("thunk of fn %s: %s", v.defn.name, v.defn.body)
}
func (v thunkValue) Eq(u Value) bool {
	panic("Illegal to compare thunk values!")
}
func (c *Context) unwrapThunk(thunk thunkValue) (v Value, err *runtimeError) {
	for isThunk := true; isThunk; thunk, isThunk = v.(thunkValue) {
		v, err = c.evalExprWithOpt(thunk.defn.body, thunk.scope, true)
		if err != nil {
			err.stackTrace = append(err.stackTrace, stackEntry{
				name: thunk.defn.name,
				pos:  thunk.defn.pos(),
			})
			return
		}
	}

	return
}

type scope struct {
	parent *scope
	vars   map[string]Value
}

func (sc *scope) get(name string) (Value, *runtimeError) {
	if v, ok := sc.vars[name]; ok {
		return v, nil
	}
	if sc.parent != nil {
		return sc.parent.get(name)
	}
	return nil, &runtimeError{
		reason: fmt.Sprintf("%s is undefined", name),
	}
}

func (sc *scope) put(name string, v Value) {
	sc.vars[name] = v
}

func (sc *scope) update(name string, v Value) *runtimeError {
	if _, ok := sc.vars[name]; ok {
		sc.vars[name] = v
		return nil
	}
	if sc.parent != nil {
		return sc.parent.update(name, v)
	}
	return &runtimeError{
		reason: fmt.Sprintf("%s is undefined", name),
	}
}

type engine struct {
	// interpreter lock to ensure lack of data races
	sync.Mutex
	// interpreter event loop waitgroup
	sync.WaitGroup
	// for deduplicating imports
	importMap map[string]scope
	// file fd -> Go's File map
	fileMap map[uintptr]*os.File
	fdLock  sync.Mutex
	// log async error streams through this
	reportErr func(error)
}

type Context struct {
	// shared interpreter state
	eng *engine
	// directory containing the root file of this context, used for loading
	// other modules with relative paths / URLs
	rootPath string
	// top level ("global") scope of this context
	scope
}

func NewContext(rootPath string) Context {
	eng := engine{
		importMap: map[string]scope{},
		fileMap:   map[uintptr]*os.File{},
		reportErr: func(err error) {
			fmt.Println(err)
		},
	}
	return Context{
		eng:      &eng,
		rootPath: rootPath,
		scope: scope{
			parent: nil,
			vars:   map[string]Value{},
		},
	}
}

func (c *Context) ChildContext(rootPath string) Context {
	return Context{
		eng:      c.eng,
		rootPath: rootPath,
		scope: scope{
			parent: nil,
			vars:   map[string]Value{},
		},
	}
}

func (c *Context) Lock() {
	c.eng.Lock()
}

func (c *Context) Unlock() {
	c.eng.Unlock()
}

func (c *Context) Wait() {
	c.eng.Wait()
}

type stackEntry struct {
	name string
	pos
}

func (e stackEntry) String() string {
	if e.name != "" {
		return fmt.Sprintf("  in fn %s %s", e.name, e.pos)
	}
	return fmt.Sprintf("  in anonymous fn %s", e.pos)
}

type runtimeError struct {
	reason string
	pos
	stackTrace []stackEntry
}

func (e *runtimeError) Error() string {
	trace := make([]string, len(e.stackTrace))
	for i, entry := range e.stackTrace {
		trace[i] = entry.String()
	}
	return fmt.Sprintf("Runtime error %s: %s\n%s", e.pos, e.reason, strings.Join(trace, "\n"))
}

func (c *Context) Eval(programReader io.Reader) (Value, error) {
	c.Lock()
	defer c.Unlock()

	program, err := io.ReadAll(programReader)
	if err != nil {
		return nil, err
	}

	tokenizer := newTokenizer(string(program))
	tokens := tokenizer.tokenize()

	parser := newParser(tokens)
	nodes, err := parser.parse()
	if err != nil {
		return nil, err
	}

	val, runtimeErr := c.evalNodes(nodes)
	if runtimeErr == nil {
		return val, nil
	}
	return val, runtimeErr

}

func (c *Context) EvalFnValue(maybeFn Value, thunkable bool, args ...Value) (Value, *runtimeError) {
	if fn, ok := maybeFn.(FnValue); ok {
		if len(args) < len(fn.defn.args) {
			// if not enough arguments, fill them with nulls
			difference := len(fn.defn.args) - len(args)
			extraArgs := make([]Value, difference)
			for i := 0; i < difference; i++ {
				extraArgs[i] = null
			}
			args = append(args, extraArgs...)
		}

		fnScope := scope{
			parent: &fn.scope,
			vars:   map[string]Value{},
		}
		for i, argName := range fn.defn.args {
			if argName != "" {
				fnScope.put(argName, args[i])
			}
		}

		if fn.defn.restArg != "" {
			var restList ListValue
			if len(args) > len(fn.defn.args) {
				restList = ListValue(args[len(fn.defn.args):])
			} else {
				restList = ListValue{}
			}

			fnScope.put(fn.defn.restArg, &restList)
		}

		thunk := thunkValue{
			defn:  fn.defn,
			scope: fnScope,
		}
		if thunkable {
			return thunk, nil
		}

		return c.unwrapThunk(thunk)
	} else if fn, ok := maybeFn.(BuiltinFnValue); ok {
		return fn.fn(args)
	} else {
		return nil, &runtimeError{
			reason: fmt.Sprintf("%s is not a function and cannot be called", maybeFn),
		}
	}
}

func (c *Context) evalNodes(nodes []astNode) (Value, *runtimeError) {
	var returnVal Value = null
	var err *runtimeError
	for _, expr := range nodes {
		returnVal, err = c.evalExpr(expr, c.scope)
		if err != nil {
			return nil, err
		}
	}
	return returnVal, nil
}

func intBinaryOp(op tokKind, left, right IntValue) (Value, *runtimeError) {
	switch op {
	case plus:
		return IntValue(left + right), nil
	case minus:
		return IntValue(left - right), nil
	case times:
		return IntValue(left * right), nil
	case divide:
		if right == 0 {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Division by zero"),
			}
		}
		return IntValue(left / right), nil
	case modulus:
		return IntValue(left % right), nil
	case xor:
		return IntValue(left ^ right), nil
	case and:
		return IntValue(left & right), nil
	case or:
		return IntValue(left | right), nil
	case greater:
		return BoolValue(left > right), nil
	case less:
		return BoolValue(left < right), nil
	case geq:
		return BoolValue(left >= right), nil
	case leq:
		return BoolValue(left <= right), nil
	}
	return nil, &runtimeError{
		reason: fmt.Sprintf("Invalid binary operator %s for ints %s, %s", token{kind: op}, left, right),
	}
}

func floatBinaryOp(op tokKind, left, right FloatValue) (Value, *runtimeError) {
	switch op {
	case plus:
		return FloatValue(left + right), nil
	case minus:
		return FloatValue(left - right), nil
	case times:
		return FloatValue(left * right), nil
	case divide:
		if right == 0 {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Division by zero"),
			}
		}
		return FloatValue(left / right), nil
	case modulus:
		return FloatValue(math.Mod(float64(left), float64(right))), nil
	case greater:
		return BoolValue(left > right), nil
	case less:
		return BoolValue(left < right), nil
	case geq:
		return BoolValue(left >= right), nil
	case leq:
		return BoolValue(left <= right), nil
	}
	return nil, &runtimeError{
		reason: fmt.Sprintf("Invalid binary operator %s for floats %s, %s", token{kind: op}, left, right),
	}
}

func (c *Context) evalAsObjKey(node astNode, sc scope) (Value, *runtimeError) {
	if ident, ok := node.(identifierNode); ok {
		return MakeString(ident.payload), nil
	}

	return c.evalExpr(node, sc)
}

func (c *Context) evalExpr(node astNode, sc scope) (Value, *runtimeError) {
	return c.evalExprWithOpt(node, sc, false)
}

func incompatibleError(op tokKind, left, right Value, position pos) *runtimeError {
	return &runtimeError{
		reason: fmt.Sprintf("Cannot %s incompatible values %s, %s",
			token{kind: op}, left, right),
		pos: position,
	}
}

func (c *Context) evalExprWithOpt(node astNode, sc scope, thunkable bool) (Value, *runtimeError) {
	switch n := node.(type) {
	case emptyNode:
		return empty, nil
	case nullNode:
		return null, nil
	case stringNode:
		return MakeString(n.payload), nil
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
		var err *runtimeError
		elems := make([]Value, len(n.elems))
		for i, elNode := range n.elems {
			elems[i], err = c.evalExpr(elNode, sc)
			if err != nil {
				return nil, err
			}
		}
		list := ListValue(elems)
		return &list, nil
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
				case *StringValue:
					keyString = string(*typedKey)
				case IntValue:
					keyString = typedKey.String()
				case FloatValue:
					keyString = typedKey.String()
				default:
					return nil, &runtimeError{reason: fmt.Sprintf("Expected a string or number as object key, got %s", key.String()),
						pos: entry.key.pos(),
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
		val, err := sc.get(n.payload)
		if err != nil {
			err.pos = n.pos()
		}
		return val, err
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
					err.pos = n.pos()
					return nil, err
				}
			}
			return assignedValue, nil
		case listNode:
			assignedList, ok := assignedValue.(*ListValue)
			if !ok {
				return nil, &runtimeError{
					reason: fmt.Sprintf("right side %s of list destructuring is not a list", n.right),
					pos:    n.pos(),
				}
			}

			for i, mustBeIdent := range left.elems {
				ident, ok := mustBeIdent.(identifierNode)
				if !ok {
					if _, ok = mustBeIdent.(emptyNode); ok {
						continue
					}

					return nil, &runtimeError{
						reason: fmt.Sprintf("element %s in destructured list %s is not an identifier", mustBeIdent, left),
						pos:    n.pos(),
					}
				}

				var destructuredEl Value
				if i < len(*assignedList) {
					destructuredEl = (*assignedList)[i]
				} else {
					destructuredEl = null
				}

				if n.isLocal {
					sc.put(ident.payload, destructuredEl)
				} else {
					err := sc.update(ident.payload, destructuredEl)
					if err != nil {
						return nil, err
					}
				}
			}
			return assignedValue, nil
		case objectNode:
			assignedObj, ok := assignedValue.(ObjectValue)
			if !ok {
				return nil, &runtimeError{
					reason: fmt.Sprintf("right side %s of object destructuring is not an object", n.right),
					pos:    n.pos(),
				}
			}

			for _, entryNode := range left.entries {
				key, err := c.evalAsObjKey(entryNode.key, sc)
				if err != nil {
					return nil, err
				}

				mustBeIdent := entryNode.val
				ident, ok := mustBeIdent.(identifierNode)
				if !ok {
					if _, ok = mustBeIdent.(emptyNode); ok {
						continue
					}

					return nil, &runtimeError{
						reason: fmt.Sprintf("value %s in destructured object %s is not an identifier", mustBeIdent, left),
						pos:    n.pos(),
					}
				}

				var keyString string
				if key, ok := key.(*StringValue); ok {
					keyString = string(*key)
				} else {
					keyString = key.String()
				}

				var destructuredEl Value
				if val, ok := assignedObj[keyString]; ok {
					destructuredEl = val
				} else {
					destructuredEl = null
				}

				if n.isLocal {
					sc.put(ident.payload, destructuredEl)
				} else {
					err := sc.update(ident.payload, destructuredEl)
					if err != nil {
						return nil, err
					}
				}
			}
			return assignedValue, nil
		case propertyAccessNode:
			assign := left

			assignLeft, err := c.evalExpr(assign.left, sc)
			if err != nil {
				return nil, err
			}

			assignRight, err := c.evalAsObjKey(assign.right, sc)
			if err != nil {
				return nil, err
			}

			switch target := assignLeft.(type) {
			case *StringValue:
				assignedString, ok := assignedValue.(*StringValue)
				if !ok {
					return nil, &runtimeError{
						reason: fmt.Sprintf("Cannot assign non-string value %s to string in %s", assignedValue, assign),
						pos:    n.pos(),
					}
				}

				byteIndexVal, ok := assignRight.(IntValue)
				if !ok {
					return nil, &runtimeError{
						reason: fmt.Sprintf("Cannot index into string with non-integer index %s", assignRight),
						pos:    n.pos(),
					}
				}
				byteIndex := int(byteIndexVal)

				if byteIndex < 0 || byteIndex > len(*target) {
					return nil, &runtimeError{
						reason: fmt.Sprintf("String assignment index %d out of range in %s", byteIndex, n),
						pos:    n.pos(),
					}
				}

				if byteIndex == len(*target) {
					// append
					*target = append(*target, *assignedString...)
				} else {
					for byteOffset, byteAtOffset := range *assignedString {
						if byteIndex+byteOffset < len(*target) {
							(*target)[byteIndex+byteOffset] = byteAtOffset
						} else {
							*target = append(*target, byteAtOffset)
						}
					}
				}
			case *ListValue:
				listIndexVal, ok := assignRight.(IntValue)
				if !ok {
					return nil, &runtimeError{
						reason: fmt.Sprintf("Cannot index into list with non-integer index %s", assignRight),
						pos:    n.pos(),
					}
				}
				listIndex := int(listIndexVal)

				if listIndex < 0 || listIndex > len(*target) {
					return nil, &runtimeError{
						reason: fmt.Sprintf("List assignment index %d out of range in %s", listIndex, n),
						pos:    n.pos(),
					}
				}

				if listIndex == len(*target) {
					*target = append(*target, assignedValue)
				} else {
					(*target)[listIndex] = assignedValue
				}
			case ObjectValue:
				var objKeyString string
				if objKey, ok := assignRight.(*StringValue); ok {
					objKeyString = string(*objKey)
				} else {
					objKeyString = assignRight.String()
				}

				if _, ok := assignedValue.(EmptyValue); ok {
					delete(target, objKeyString)
				} else {
					target[objKeyString] = assignedValue
				}
			default:
				return nil, &runtimeError{
					reason: fmt.Sprintf("Expected string, list, or object in left-hand side of property assignment, got %s", left.String()),
					pos:    n.pos(),
				}
			}

			return assignLeft, nil
		}
	case propertyAccessNode:
		left, err := c.evalExpr(n.left, sc)
		if err != nil {
			return nil, err
		}

		right, err := c.evalAsObjKey(n.right, sc)
		if err != nil {
			return nil, err
		}

		switch target := left.(type) {
		case *StringValue:
			byteIndex, ok := right.(IntValue)
			if !ok {
				return nil, &runtimeError{
					reason: fmt.Sprintf("Cannot index into string with non-integer index %s", right),
					pos:    n.pos(),
				}
			}

			if byteIndex < 0 || int64(byteIndex) >= int64(len(*target)) {
				return null, nil
			}

			targetByte := StringValue([]byte{(*target)[byteIndex]})
			return &targetByte, nil
		case *ListValue:
			listIndex, ok := right.(IntValue)
			if !ok {
				return nil, &runtimeError{
					reason: fmt.Sprintf("Cannot index into list with non-integer index %s", right),
					pos:    n.pos(),
				}
			}

			if listIndex < 0 || int64(listIndex) >= int64(len(*target)) {
				return null, nil
			}

			return (*target)[listIndex], nil
		case ObjectValue:
			var objKeyString string
			if objKey, ok := right.(*StringValue); ok {
				objKeyString = string(*objKey)
			} else {
				objKeyString = right.String()
			}

			if val, ok := target[objKeyString]; ok {
				return val, nil
			}

			return null, nil
		}

		return nil, &runtimeError{
			reason: fmt.Sprintf("Expected string, list, or object in left-hand side of property access, got %s", left.String()),
			pos:    n.pos(),
		}
	case unaryNode:
		rightComputed, err := c.evalExpr(n.right, sc)
		if err != nil {
			return nil, err
		}

		switch right := rightComputed.(type) {
		case IntValue:
			switch n.op {
			case plus:
				return right, nil
			case minus:
				return -right, nil
			}
		case FloatValue:
			switch n.op {
			case plus:
				return right, nil
			case minus:
				return -right, nil
			}
		case BoolValue:
			switch n.op {
			case exclam:
				return !right, nil
			}
		}
		return nil, &runtimeError{
			reason: fmt.Sprintf("%s is not a valid unary operator for %s", token{kind: n.op}, rightComputed),
			pos:    n.pos(),
		}
	case binaryNode:
		leftComputed, err := c.evalExpr(n.left, sc)
		if err != nil {
			return nil, err
		}

		rightComputed, err := c.evalExpr(n.right, sc)
		if err != nil {
			return nil, err
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
					return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
				}

				leftFloat := FloatValue(float64(int64(left)))
				val, err := floatBinaryOp(n.op, leftFloat, rightFloat)
				if err != nil {
					err.pos = n.pos()
				}
				return val, err
			}

			val, err := intBinaryOp(n.op, left, right)
			if err != nil {
				err.pos = n.pos()
			}
			return val, err
		case FloatValue:
			right, ok := rightComputed.(FloatValue)
			if !ok {
				rightInt, ok := rightComputed.(IntValue)
				if !ok {
					return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
				}

				right = FloatValue(float64(int64(rightInt)))
				val, err := floatBinaryOp(n.op, left, right)
				if err != nil {
					err.pos = n.pos()
				}
				return val, err
			}

			val, err := floatBinaryOp(n.op, left, right)
			if err != nil {
				err.pos = n.pos()
			}
			return val, err
		case *StringValue:
			right, ok := rightComputed.(*StringValue)
			if !ok {
				return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
			}

			switch n.op {
			case plus:
				base := make([]byte, 0, len(*left)+len(*right))
				base = append(base, *left...)
				base = append(base, *right...)
				baseStr := StringValue(base)
				return &baseStr, nil
			case xor:
				max := maxLen(*left, *right)

				ls, rs := zeroExtend(*left, max), zeroExtend(*right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] ^ rs[i]
				}
				resStr := StringValue(res)
				return &resStr, nil
			case and:
				max := maxLen(*left, *right)

				ls, rs := zeroExtend(*left, max), zeroExtend(*right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] & rs[i]
				}
				resStr := StringValue(res)
				return &resStr, nil
			case or:
				max := maxLen(*left, *right)

				ls, rs := zeroExtend(*left, max), zeroExtend(*right, max)
				res := make([]byte, max)
				for i := range res {
					res[i] = ls[i] | rs[i]
				}
				resStr := StringValue(res)
				return &resStr, nil
			case pushArrow:
				*left = append(*left, *right...)
				return left, nil
			case greater:
				return BoolValue(bytes.Compare(*left, *right) > 0), nil
			case less:
				return BoolValue(bytes.Compare(*left, *right) < 0), nil
			case geq:
				return BoolValue(bytes.Compare(*left, *right) >= 0), nil
			case leq:
				return BoolValue(bytes.Compare(*left, *right) <= 0), nil
			}
			return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
		case BoolValue:
			right, ok := rightComputed.(BoolValue)
			if !ok {
				return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
			}

			switch n.op {
			case plus, or:
				return BoolValue(left || right), nil
			case times, and:
				return BoolValue(left && right), nil
			case xor:
				return BoolValue(left != right), nil
			}
		case *ListValue:
			switch n.op {
			case pushArrow:
				*left = append(*left, rightComputed)
				return left, nil
			}
			return nil, incompatibleError(n.op, leftComputed, rightComputed, n.pos())
		}
		return nil, &runtimeError{
			reason: fmt.Sprintf("Binary operator %s is not defined for values %s, %s",
				token{kind: n.op}, leftComputed, rightComputed),
			pos: n.pos(),
		}
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
		if n.restArg != nil {
			rest, err := c.evalExpr(n.restArg, sc)
			if err != nil {
				return nil, err
			}

			restList, ok := rest.(*ListValue)
			if !ok {
				return nil, &runtimeError{
					reason: fmt.Sprintf("Cannot spread a non-list value %s in a function call %s", rest, n),
					pos:    n.pos(),
				}
			}

			args = append(args, *restList...)
		}

		val, err := c.EvalFnValue(maybeFn, thunkable, args...)
		// we only overwrite the error pos if it's nil (i.e. if it was a "nil
		// is not a function" error, where EvalFnValue can't correctly position
		// the error itself)
		if err != nil && err.pos.line == 0 {
			err.pos = n.pos()
		}
		return val, err
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
				return c.evalExprWithOpt(branch.body, sc, thunkable)
			}
		}
		return null, nil
	case blockNode:
		// empty block returns ? (null)
		if len(n.exprs) == 0 {
			return null, nil
		}

		blockScope := scope{
			parent: &sc,
			vars:   map[string]Value{},
		}

		last := len(n.exprs) - 1
		for _, expr := range n.exprs[:last] {
			_, err := c.evalExprWithOpt(expr, blockScope, false)
			if err != nil {
				return nil, err
			}
		}

		return c.evalExprWithOpt(n.exprs[last], blockScope, thunkable)
	}

	panic(fmt.Sprintf("Unexpected astNode type: %s", node))
}
