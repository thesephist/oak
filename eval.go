package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
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

// TODO: values

// Null need not contain any data, so we use the most compact data
// representation we can.
type NullValue byte

// interned "null"
const null NullValue = 0

func (v NullValue) String() string {
	return "?"
}
func (v NullValue) Eq(u Value) bool {
	if _, ok := u.(NullValue); ok {
		return true
	}
	return false
}

type StringValue []byte

var emptyString = StringValue("")

func (v StringValue) String() string {
	return fmt.Sprintf("\"%v\"", string(v))
}
func (v StringValue) Eq(u Value) bool {
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
	if w, ok := u.(IntValue); ok {
		return v == w
	}
	return false
}

type FnValue struct {
	defn *fnNode
	scope
}

func (v FnValue) String() string {
	return "TODO" // TODO: fix this!
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

type Context struct {
	Cwd        string
	SourcePath string

	scope
}

func NewContext(path string, cwd string) Context {
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
	_, err := io.ReadAll(programReader)
	if err != nil {
		return nil, err
	}

	// TODO: tokenize and eval

	return null, nil
}

func (c *Context) evalProgram(nodes []astNode) (Value, error) {
	var val Value
	var err error
	for _, expr := range nodes {
		val, err = c.evalExpr(expr)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (c *Context) evalExpr(node astNode) (Value, error) {
	switch node.(type) {
	case nullNode:
		// TODO: evalers
		return null, nil
	}
	return null, nil
}
