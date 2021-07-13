package main

import "fmt"

const progSample = `
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

const prog = `
fn main {
	print('Hello, World!\n')
}

curried := fn (a) fn (b) fn (c) {
	print(a)
	print(b)
	print(c)
	print('\n')
}

main()

curried('first')('second')('third')
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
	ctx.LoadBuiltins()

	val, err := ctx.evalProgram(nodes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val)
}
