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

type builtinFnValue func(*Context, ...[]Value) (Value, error)

// TODO: temp helper, remove when it's time
func (c *Context) mgnPrint(s StringValue) (Value, error) {
	fmt.Print(string(s))
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
