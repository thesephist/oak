package main

import "os"

func main() {
	if runPackFile() {
		return
	}

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
