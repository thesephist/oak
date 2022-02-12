package main

import (
	"os"
)

const helpMsg = `
Oak is a small, dynamic, functional programming language.

By default, oak interprets from standard input.
	oak < main.oak
Run Oak programs from a source file by passing it to oak.
	oak main.oak
Run oak with no arguments to start an interactive repl.
	oak
	>
`

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if isCommand := performCommandIfExists(arg); !isCommand {
			runFile(arg)
		}
		return
	}

	if isStdinReadable() {
		runStdin()
		return
	}

	runRepl()
}
