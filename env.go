package main

import (
	"fmt"
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

func (c *Context) LoadBuiltins() {
	c.scope.put("print", BuiltinFnValue{
		name: "print",
		fn:   c.mgnPrint,
	})
}

// TODO: temp helper, remove when it's time
func (c *Context) mgnPrint(args []Value) (Value, error) {
	err := c.requireArgLen("print", args, 1)
	if err != nil {
		return nil, err
	}

	outputString, ok := args[0].(StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("unexpected argument to print"),
		}
	}

	fmt.Print(string(outputString))
	return null, nil
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
