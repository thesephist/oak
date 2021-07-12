package main

import "fmt"

const prog = `
fn say(a, b, c...) {
	std.println.do('Hello, World!')
}

say(1, 2, 3)(xyz)`

func main() {
	tokenizer := newTokenizer(prog)
	tokens := tokenizer.tokenize()
	fmt.Println(tokens)

	parser := newParser(tokens)
	nodes, err := parser.parse()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(nodes)

	ctx := NewContext("<input>", "/tmp")
	val, err := ctx.evalProgram(nodes)
	if err != nil {
		fmt.Println("%v", err)
	}
	fmt.Println(val)
}
