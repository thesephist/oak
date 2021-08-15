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

//go:embed lib/http.mgn
var libhttp string

//go:embed lib/syntax.mgn
var libsyntax string

//go:embed lib/test.mgn
var libtest string

var stdlibs = map[string]string{
	"std":    libstd,
	"str":    libstr,
	"math":   libmath,
	"sort":   libsort,
	"fs":     libfs,
	"fmt":    libfmt,
	"http":   libhttp,
	"syntax": libsyntax,
	"test":   libtest,
}

func isStdLib(name string) bool {
	_, ok := stdlibs[name]
	return ok
}

func (c *Context) LoadLib(name string) (Value, *runtimeError) {
	program, ok := stdlibs[name]
	if !ok {
		return nil, &runtimeError{
			reason: fmt.Sprintf("%s is not a valid standard library; could not import", name),
		}
	}

	if imported, ok := c.eng.importMap[name]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := c.ChildContext(c.rootPath)
	ctx.LoadBuiltins()

	ctx.Unlock()
	_, err := ctx.Eval(strings.NewReader(program))
	ctx.Lock()
	if err != nil {
		if runtimeErr, ok := err.(*runtimeError); ok {
			return nil, runtimeErr
		} else {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Error loading %s: %s", name, err.Error()),
			}
		}
	}

	c.eng.importMap[name] = ctx.scope
	return ObjectValue(ctx.scope.vars), nil
}
