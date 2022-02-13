package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/chzyer/readline"
)

//go:embed cmd/version.oak
var cmdversion string

//go:embed cmd/help.oak
var cmdhelp string

//go:embed cmd/cat.oak
var cmdcat string

//go:embed cmd/fmt.oak
var cmdfmt string

//go:embed cmd/build.oak
var cmdbuild string

var cliCommands = map[string]string{
	"version": cmdversion,
	"help":    cmdhelp,
	"cat":     cmdcat,
	"fmt":     cmdfmt,
	"build":   cmdbuild,
}

func isStdinReadable() bool {
	stdin, _ := os.Stdin.Stat()
	return (stdin.Mode() & os.ModeCharDevice) == 0
}

func performCommandIfExists(command string) bool {
	switch command {
	case "repl":
		runRepl()
		return true
	case "eval":
		runEval()
		return true
	case "pipe":
		runPipe()
		return true
	}

	commandProgram, ok := cliCommands[command]
	if !ok {
		return false
	}

	ctx := NewContextWithCwd()
	defer ctx.Wait()
	ctx.LoadBuiltins()

	if _, err := ctx.Eval(strings.NewReader(commandProgram)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return true
}

func runFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", filePath, err)
		os.Exit(1)
	}
	defer file.Close()

	ctx := NewContext(path.Dir(filePath))
	defer ctx.Wait()
	ctx.LoadBuiltins()

	if _, err = ctx.Eval(file); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runStdin() {
	ctx := NewContextWithCwd()
	defer ctx.Wait()
	ctx.LoadBuiltins()

	if _, err := ctx.Eval(os.Stdin); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runRepl() {
	var historyFilePath string
	homeDir, err := os.UserHomeDir()
	if err == nil {
		historyFilePath = path.Join(homeDir, ".oak_history")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "> ",
		HistoryFile: historyFilePath,
	})
	if err != nil {
		fmt.Println("Could not open the repl")
		os.Exit(1)
	}
	defer rl.Close()

	ctx := NewContextWithCwd()
	ctx.LoadBuiltins()
	ctx.mustLoadAllLibs()

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

		// keep last evaluated result as __ in REPL
		ctx.scope.put("__", val)
	}
}

func runEval() {
	ctx := NewContextWithCwd()
	defer ctx.Wait()
	ctx.LoadBuiltins()
	ctx.mustLoadAllLibs()

	if isStdinReadable() {
		allInput, _ := io.ReadAll(os.Stdin)
		allInputValue := StringValue(allInput)
		ctx.scope.put("stdin", &allInputValue)
	}

	prog := strings.Join(os.Args[2:], " ")
	if val, err := ctx.Eval(strings.NewReader(prog)); err == nil {
		if stringVal, ok := val.(*StringValue); ok {
			fmt.Println(string(*stringVal))
		} else {
			fmt.Println(val)
		}
	} else {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runPipe() {
	if !isStdinReadable() {
		return
	}

	ctx := NewContextWithCwd()
	defer ctx.Wait()
	ctx.LoadBuiltins()
	ctx.mustLoadAllLibs()

	rootScope := ctx.scope
	stdin := bufio.NewReader(os.Stdin)
	prog := strings.Join(os.Args[2:], " ")
	for i := 0; ; i++ {
		line, err := stdin.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Could not read piped input")
			os.Exit(1)
		}

		line = bytes.TrimSuffix(line, []byte{'\n'})
		lineValue := StringValue(line)
		// each line gets its own top-level subscope to avoid collisions
		ctx.subScope(&rootScope)
		ctx.scope.put("line", &lineValue)
		ctx.scope.put("i", IntValue(i))

		// NOTE: currently, the same program is re-tokenized and re-parsed on
		// every line. This is not efficient, and can be optimized in the
		// future by parsing once and reusing a single AST.
		outValue, err := ctx.Eval(strings.NewReader(prog))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var outLine []byte
		switch v := outValue.(type) {
		case NullValue:
			// lines that return ? are filtered out entirely, which lets Oak's
			// shorthand `if pattern -> action` notation become very useful
			continue
		case *StringValue:
			outLine = []byte(*v)
		default:
			outLine = []byte(outValue.String())
		}
		if _, err := os.Stdout.Write(append(outLine, '\n')); err != nil {
			fmt.Println("Could not write piped output")
			os.Exit(1)
		}
	}
}
