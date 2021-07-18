package main

import (
	"fmt"
	"strings"
	"testing"
)

func expectProgramToReturn(t *testing.T, program string, expected Value) {
	ctx := NewContext("<input>", "/tmp")
	ctx.LoadBuiltins()
	val, err := ctx.Eval(strings.NewReader(program))
	if err != nil {
		t.Errorf("Did not expect program to exit with error: %s", err.Error())
	}
	if val == nil {
		t.Errorf("Return valuue of program should not be nil")
	} else if !val.Eq(expected) {
		t.Errorf(fmt.Sprintf("Expected and returned values don't match: %s != %s", expected, val))
	}
}

func TestEvalEmptyProgram(t *testing.T) {
	expectProgramToReturn(t, "", null)
}

func TestEmptyLiteral(t *testing.T) {
	expectProgramToReturn(t, "_", empty)
}

func TestNullLiteral(t *testing.T) {
	expectProgramToReturn(t, "?", null)
}

func TestStringLiteral(t *testing.T) {
	expectProgramToReturn(t, "'Hello, World!\\n'", MakeString("Hello, World!\n"))
}

func TestIntegerLiteral(t *testing.T) {
	expectProgramToReturn(t, "64710", IntValue(64710))
}

func TestFloatLiteral(t *testing.T) {
	expectProgramToReturn(t, "3.141592", FloatValue(3.141592))
}

func TestAtomLiteral(t *testing.T) {
	expectProgramToReturn(t, ":not_found_404", AtomValue("not_found_404"))
}

func TestFunctionDefAndCall(t *testing.T) {
	expectProgramToReturn(t, `fn getThree() { x := 4, 3 }, getThree()`, IntValue(3))
}

func TestLocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `x := 100, y := 200, x`, IntValue(100))
}

func TestNonlocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `
	x := 100
	y := 200
	fn do {
		x <- x + 100
		y := y + 100
	}
	do()
	x + y
	`, IntValue(400))
}

func TestBasicBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `1 + 2 * 3`, IntValue(7))
}

func TestBinaryExprWithParens(t *testing.T) {
	// NOTE: this program, in most other languages like C and Node, evaluate to
	// 2. -2 is the result we get in Magnolia (and Ink) due to our specific
	// operator precedence. We might change this later, but for now this is the
	// designed behavior.
	expectProgramToReturn(t, `(1 + 2) / 3 - 1 + (10 + (20 / 5)) % 3`, IntValue(-2))
}

func TestLongBinaryExprWithPrecedence(t *testing.T) {
	expectProgramToReturn(t, `x := 1 + 2 * 3 + 4 / 2 + 10 % 4, x % 5 + x`, IntValue(12))
}

func TestBinaryExprWithComplexTerms(t *testing.T) {
	expectProgramToReturn(t, `
	fn double(n) 2 * n
	fn decrement(n) n - 1
	double(10) + if decrement(10) { 9 -> 2, _ -> 1 } + 8
	`, IntValue(30))
}

func TestEmptyIfExpr(t *testing.T) {
	expectProgramToReturn(t, `if 100 {}`, null)
}

func TestBasicIfExpr(t *testing.T) {
	expectProgramToReturn(t, `if 2 * 2 {
		? -> 100
		{ a: 'b' } -> 200
		5 -> 'five'
		4 -> 'four'
	}`, MakeString("four"))
}

func TestIfExprWithEmpty(t *testing.T) {
	expectProgramToReturn(t, `if 10 + 2 {
		12 -> 'twelve'
		_ -> 'wrong'
	}`, MakeString("twelve"))
}

func TestIfExprInFunction(t *testing.T) {
	expectProgramToReturn(t, `
	fn even?(n) if n % 2 {
		0 -> true
		_ -> false
	}
	even?(100)
	`, BoolValue(true))
}

func TestBasicWithExpr(t *testing.T) {
	expectProgramToReturn(t, `fn add(a, b) { a + b }, with add(10) 40`, IntValue(50))
}

func TestWithExprWithCallback(t *testing.T) {
	expectProgramToReturn(t, `fn applyThrice(x, f) f(f(f(x))), with applyThrice(10) fn(n) n + 1`, IntValue(13))
}

func TestRecursiveFunction(t *testing.T) {
	expectProgramToReturn(t, `
	fn times(n, f) {
		fn sub(i) if i {
			n -> ?
			_ -> {
				f(i)
				sub(i + 1)
			}
		}
		sub(0)
	}

	counter := 0
	with times(10) fn(i) {
		counter <- counter + i * 10
	}
	counter
	`, IntValue(450))
}

func TestREcursiveFunctionOnList(t *testing.T) {
	expectProgramToReturn(t, `
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

	sum := 0
	list := [1, 2, 3, 4, 5]
	with each(list) fn(it) {
		sum <- sum + it
	}
	sum
	`, IntValue(15))
}

func TestCurriedFunctionDef(t *testing.T) {
	expectProgramToReturn(t, `
	addThree := fn(a) fn(b) fn(c) {
		a + b + c
	}

	almost := addThree(15)(20)
	almost(8)
	`, IntValue(15+20+8))
}
