package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed lib/std.mgn
var libstd string

//go:embed lib/str.mgn
var libstr string

//go:embed lib/math.mgn
var libmath string

//go:embed lib/sort.mgn
var libsort string

//go:embed lib/fs.mgn
var libfs string

//go:embed lib/fmt.mgn
var libfmt string

//go:embed lib/test.mgn
var libtest string

var stdlibs = map[string]string{
	"std":  libstd,
	"str":  libstr,
	"math": libmath,
	"sort": libsort,
	"fs":   libfs,
	"fmt":  libfmt,
	"test": libtest,
}

func isStdLib(name string) bool {
	_, ok := stdlibs[name]
	return ok
}

func (c *Context) LoadLib(name string) (Value, error) {
	program, ok := stdlibs[name]
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("%s is not a valid standard library; could not import", name),
		}
	}

	if imported, ok := c.importMap[name]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := NewContext(c.rootPath)
	ctx.importMap = c.importMap
	ctx.LoadBuiltins()
	_, err := ctx.Eval(strings.NewReader(program))
	if err != nil {
		return nil, err
	}

	c.importMap[name] = ctx.scope
	return ObjectValue(ctx.scope.vars), nil
}
