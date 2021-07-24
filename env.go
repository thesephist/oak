package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func (c *Context) requireArgLen(fnName string, args []Value, count int) error {
	if len(args) < count {
		return runtimeError{
			reason:     fmt.Sprintf("%s requires %d arguments, got %d", fnName, count, len(args)),
			stackTrace: c.generateStackTrace(),
		}
	}

	return nil
}

type BuiltinFnValue struct {
	name string
	fn   func([]Value) (Value, error)
}

func (v BuiltinFnValue) String() string {
	return fmt.Sprintf("fn %s { <native fn> }", v.name)
}
func (v BuiltinFnValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(BuiltinFnValue); ok {
		return v.name == w.name
	}
	return false
}

func (c *Context) LoadFunc(name string, fn func([]Value) (Value, error)) {
	c.scope.put(name, BuiltinFnValue{
		name: name,
		fn:   fn,
	})
}

func (c *Context) LoadBuiltins() {
	c.LoadFunc("import", c.mgnImport)
	c.LoadFunc("type", c.mgnType)
	c.LoadFunc("string", c.mgnString)
	c.LoadFunc("int", c.mgnInt)
	c.LoadFunc("float", c.mgnFloat)
	c.LoadFunc("codepoint", c.mgnCodepoint)
	c.LoadFunc("char", c.mgnChar)
	c.LoadFunc("len", c.mgnLen)
	c.LoadFunc("print", c.mgnPrint)
	c.LoadFunc("keys", c.mgnKeys)
}

func (c *Context) mgnString(args []Value) (Value, error) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return arg, nil
	default:
		return MakeString(arg.String()), nil
	}
}

func (c *Context) mgnInt(args []Value) (Value, error) {
	if err := c.requireArgLen("int", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		return arg, nil
	case FloatValue:
		return IntValue(math.Trunc(float64(arg))), nil
	case *StringValue:
		n, err := strconv.ParseInt(string(*arg), 10, 64)
		if err != nil {
			return null, nil
		}
		return IntValue(n), nil
	default:
		return null, nil
	}
}

func (c *Context) mgnFloat(args []Value) (Value, error) {
	if err := c.requireArgLen("float", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		return FloatValue(arg), nil
	case FloatValue:
		return arg, nil
	case *StringValue:
		f, err := strconv.ParseFloat(string(*arg), 64)
		if err != nil {
			return null, nil
		}
		return FloatValue(f), nil
	default:
		return null, nil
	}
}

func (c *Context) mgnCodepoint(args []Value) (Value, error) {
	if err := c.requireArgLen("codepoint", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		if len(*arg) != 1 {
			return null, nil
		}
		return IntValue(int64((*arg)[0])), nil
	default:
		return null, nil
	}
}

func (c *Context) mgnChar(args []Value) (Value, error) {
	if err := c.requireArgLen("char", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		codepoint := int64(arg)
		if codepoint < 0 {
			codepoint = 0
		}
		if codepoint > 255 {
			codepoint = 255
		}
		bytes := StringValue([]byte{byte(codepoint)})
		return &bytes, nil
	default:
		return null, nil
	}
}

func (c *Context) mgnType(args []Value) (Value, error) {
	if err := c.requireArgLen("type", args, 1); err != nil {
		return nil, err
	}

	switch args[0].(type) {
	case NullValue:
		return AtomValue("null"), nil
	case EmptyValue:
		return AtomValue("empty"), nil
	case IntValue:
		return AtomValue("int"), nil
	case FloatValue:
		return AtomValue("float"), nil
	case BoolValue:
		return AtomValue("bool"), nil
	case AtomValue:
		return AtomValue("atom"), nil
	case *StringValue:
		return AtomValue("string"), nil
	case *ListValue:
		return AtomValue("list"), nil
	case ObjectValue:
		return AtomValue("object"), nil
	case FnValue, BuiltinFnValue:
		return AtomValue("function"), nil
	}

	panic("Unreachable!")
}

func (c *Context) mgnImport(args []Value) (Value, error) {
	if err := c.requireArgLen("import", args, 1); err != nil {
		return nil, err
	}

	pathBytes, ok := args[0].(*StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("path to import() must be a string, got %s", args[0]),
		}
	}
	pathStr := string(*pathBytes)

	// if a stdlib, just import the library from binary
	if isStdLib(pathStr) {
		return c.LoadLib(pathStr)
	}

	filePath := pathStr + ".mgn"
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(c.rootPath, filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, runtimeError{
			reason: fmt.Sprintf("could not open %s, %s", filePath, err.Error()),
		}
	}
	defer file.Close()

	if imported, ok := c.importMap[filePath]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := NewContext(path.Dir(filePath))
	ctx.importMap = c.importMap
	ctx.LoadBuiltins()
	_, err = ctx.Eval(file)
	if err != nil {
		return nil, err
	}

	c.importMap[filePath] = ctx.scope
	return ObjectValue(ctx.scope.vars), nil
}

func (c *Context) mgnLen(args []Value) (Value, error) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return IntValue(len(*arg)), nil
	case *ListValue:
		return IntValue(len(*arg)), nil
	case ObjectValue:
		return IntValue(len(arg)), nil
	default:
		return nil, runtimeError{
			reason: fmt.Sprintf("%s does not support a len() call", arg),
		}
	}
}

func (c *Context) mgnPrint(args []Value) (Value, error) {
	if err := c.requireArgLen("print", args, 1); err != nil {
		return nil, err
	}

	outputString, ok := args[0].(*StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("unexpected argument to print: %s", args[0]),
		}
	}

	n, _ := os.Stdout.Write(*outputString)
	return IntValue(n), nil
}

func makeIntListUpTo(max int) Value {
	list := make(ListValue, max)
	for i := 0; i < max; i++ {
		list[i] = IntValue(i)
	}
	return &list
}

func (c *Context) mgnKeys(args []Value) (Value, error) {
	if err := c.requireArgLen("print", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return makeIntListUpTo(len(*arg)), nil
	case *ListValue:
		return makeIntListUpTo(len(*arg)), nil
	case ObjectValue:
		keys := make(ListValue, len(arg))
		i := 0
		for key := range arg {
			keys[i] = MakeString(key)
			i++
		}
		return &keys, nil
	default:
		return MakeList(), nil
	}
}
