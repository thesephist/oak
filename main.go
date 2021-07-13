package main

import "fmt"

const prog = `
x := 42

fn getTwo 2

fn getThree() {
	3
}

fn getFour {
	a := 4
	b := 10
	a
}

fn getSecondArg(a, b) {
	b
}

getFive := fn() {
	getSecondArg(2, 5)
}

getFive()

[1, 2, 3]
{
	a: 12
	b: 'hello'
	c: {
		d: getFour()
	}
}
`

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
