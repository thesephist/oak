package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

const version = "0.1"

const helpMsg = `
Magnolia is a small, dynamic, functional programming language.

By default, mgn interprets from standard input.
	mgn < main.mgn
Run Magnolia programs from a source file by passing it to mgn.
	mgn main.mgn
Run mgn with no arguments to start an interactive repl.
	mgn
	>
`

func main() {
	if len(os.Args) > 1 {
		runFile()
		return
	}

	runRepl()
}

func newOsCtx() Context {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get working directory")
		os.Exit(1)
	}
	ctx := NewContext("<input>", cwd)
	ctx.LoadBuiltins()
	return ctx
}

func runRepl() {
	rl, err := readline.New("> ")
	if err != nil {
		fmt.Println("Could not open the repl.")
		os.Exit(1)
	}
	defer rl.Close()

	ctx := newOsCtx()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		val, err := ctx.Eval(strings.NewReader(line))
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(val)
	}
}

func runFile() {
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", filePath, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	ctx := newOsCtx()
	_, err = ctx.Eval(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
