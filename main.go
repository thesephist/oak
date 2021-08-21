package main

import (
	"fmt"
	"os"
	"path"
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
		arg := os.Args[1]

		// first, attempt to run argv[1] as a CLI command, and
		// fallback to reading a file
		isCommand := performCommandIfExists(arg)
		if !isCommand {
			runFile(arg)
		}
		return
	}

	runRepl()
}

func runRepl() {
	rl, err := readline.New("> ")
	if err != nil {
		fmt.Println("Could not open the repl.")
		os.Exit(1)
	}
	defer rl.Close()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get working directory")
		os.Exit(1)
	}
	ctx := NewContext(cwd)
	ctx.LoadBuiltins()

	// pre-load standard libraries into global scope
	for libname := range stdlibs {
		ctx.Eval(strings.NewReader(fmt.Sprintf("%s := import('%s')", libname, libname)))
	}

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		// if no input, don't print the null output
		if strings.TrimSpace(line) == "" {
			continue
		}

		val, err := ctx.Eval(strings.NewReader(line))
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(val)
	}
}

func runFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", filePath, err)
		os.Exit(1)
	}
	defer file.Close()

	ctx := NewContext(path.Dir(filePath))
	ctx.LoadBuiltins()
	defer ctx.Wait()

	_, err = ctx.Eval(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
