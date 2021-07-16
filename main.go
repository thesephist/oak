package main

import (
	"fmt"
	"strings"
)

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

const prog2 = `
// main loop
fn main print('Hello, World!\n')
main()

fn println(x) {
	print(x), print('\n')
}
curried := fn(a) fn(b) fn(c) {
	println(a)
	println(b)
	println(c)
}

curried('first')('second')('third')
`

const prog = `
fn println(x) {
	print(x)
	print('\n')
}

fn one?(n) if n {
	1 -> println('is true')
	_ -> println('is false')
}

one?(1)
one?(2)
one?(3)
`

func main() {
	ctx := NewContext("<input>", "/tmp")
	ctx.LoadBuiltins()

	val, err := ctx.Eval(strings.NewReader(prog))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(val)
}
