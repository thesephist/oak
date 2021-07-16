package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func (c *Context) requireArgLen(fnName string, args []Value, count int) error {
	if len(args) != count {
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
	c.LoadFunc("string", c.mgnString)
	c.LoadFunc("len", c.mgnLen)
	c.LoadFunc("print", c.mgnPrint)
}

func (c *Context) mgnString(args []Value) (Value, error) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case StringValue:
		return arg, nil
	default:
		return StringValue([]byte(arg.String())), nil
	}
}

func (c *Context) mgnImport(args []Value) (Value, error) {
	if err := c.requireArgLen("import", args, 1); err != nil {
		return nil, err
	}

	pathBytes, ok := args[0].(StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("path to import() must be a string, got %s", args[0]),
		}
	}
	path := string(pathBytes) + ".mgn"
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.rootPath, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, runtimeError{
			reason: fmt.Sprintf("could not open %s, %s", path, err.Error()),
		}
	}
	defer file.Close()

	if imported, ok := c.importMap[path]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := NewContext(path, c.rootPath)
	ctx.importMap = c.importMap
	ctx.LoadBuiltins()
	_, err = ctx.Eval(file)
	if err != nil {
		return nil, err
	}

	c.importMap[path] = ctx.scope
	return ObjectValue(ctx.scope.vars), nil
}

func (c *Context) mgnLen(args []Value) (Value, error) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case StringValue:
		return IntValue(len(arg)), nil
	case ListValue:
		return IntValue(len(arg)), nil
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

	outputString, ok := args[0].(StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("unexpected argument to print: %s", args[0]),
		}
	}

	n, _ := os.Stdout.Write(outputString)
	return IntValue(n), nil
}

func (c *Context) mgnOpen(
	path StringValue, mode IntValue, cb FnValue) (Value, error) {
	return emptyString, nil
}

func (c *Context) mgnRead(
	fd IntValue, offset IntValue, length IntValue, cb FnValue) (Value, error) {
	return emptyString, nil
}

func (c *Context) mgnWrite(
	fd IntValue, offset IntValue, data StringValue, cb FnValue) (Value, error) {
	return emptyString, nil
}
