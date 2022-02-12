package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed lib/std.oak
var libstd string

//go:embed lib/str.oak
var libstr string

//go:embed lib/math.oak
var libmath string

//go:embed lib/sort.oak
var libsort string

//go:embed lib/random.oak
var librandom string

//go:embed lib/fs.oak
var libfs string

//go:embed lib/fmt.oak
var libfmt string

//go:embed lib/json.oak
var libjson string

//go:embed lib/datetime.oak
var libdatetime string

//go:embed lib/path.oak
var libpath string

//go:embed lib/http.oak
var libhttp string

//go:embed lib/test.oak
var libtest string

//go:embed lib/debug.oak
var libdebug string

//go:embed lib/cli.oak
var libcli string

//go:embed lib/md.oak
var libmd string

//go:embed lib/crypto.oak
var libcrypto string

//go:embed lib/syntax.oak
var libsyntax string

var stdlibs = map[string]string{
	"std":      libstd,
	"str":      libstr,
	"math":     libmath,
	"sort":     libsort,
	"random":   librandom,
	"fs":       libfs,
	"fmt":      libfmt,
	"json":     libjson,
	"datetime": libdatetime,
	"path":     libpath,
	"http":     libhttp,
	"test":     libtest,
	"debug":    libdebug,
	"cli":      libcli,
	"md":       libmd,
	"crypto":   libcrypto,
	"syntax":   libsyntax,
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

func (c *Context) loadAllLibs() error {
	for libname := range stdlibs {
		_, err := c.Eval(strings.NewReader(fmt.Sprintf("%s := import('%s')", libname, libname)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) mustLoadAllLibs() {
	if err := c.loadAllLibs(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
