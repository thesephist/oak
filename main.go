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

const prog3 = `
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

const prog4 = `
fn println(x) print(string(x) + '\n')

fn fib(n) if n <= 1 {
	true -> 1
	_ -> fib(n - 2) + fib(n - 1)
}

fn each(list, f) {
	fn sub(i) if i {
		len(list) -> ?
		_ -> {
			f(list.(i))
			sub(i + 1)
		}
	}

	sub(0)
}

list := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
each(list, fn(i) println(fib(i)))
`

const prog5 = `
fn println(x) print(string(x) + '\n')

fn count(max) {
	fn sub(i) if i {
		max -> ?
		_ -> {
			println(i)
			sub(i + 1)
		}
	}

	sub(0)
}

count(20)
`

const prog = `
fn add(a, b) {
	a + b
}

with add(10) 40
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
