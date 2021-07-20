package main

import (
	"fmt"
	"strings"
	"testing"
)

func expectProgramToReturn(t *testing.T, program string, expected Value) {
	ctx := NewContext("/tmp")
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
	expectProgramToReturn(t, "100.0", FloatValue(100))
	expectProgramToReturn(t, "3.141592", FloatValue(3.141592))
}

func TestAtomLiteral(t *testing.T) {
	expectProgramToReturn(t, ":not_found_404", AtomValue("not_found_404"))
}

func TestListLiteral(t *testing.T) {
	expectProgramToReturn(t, `[1, [2, 'three'], :four]`, MakeList(
		IntValue(1),
		MakeList(
			IntValue(2),
			MakeString("three"),
		),
		AtomValue("four"),
	))
}

func TestObjectLiteral(t *testing.T) {
	expectProgramToReturn(t, `{a: 'ay', b: 200, c: {d: 'dee'}}`, ObjectValue{
		"a": MakeString("ay"),
		"b": IntValue(200),
		"c": ObjectValue{
			"d": MakeString("dee"),
		},
	})
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

func TestUnaryExpr(t *testing.T) {
	expectProgramToReturn(t, `!true`, mgnFalse)
	expectProgramToReturn(t, `!(false | true)`, mgnFalse)

	expectProgramToReturn(t, `-546`, IntValue(-546))
	expectProgramToReturn(t, `-3.250`, FloatValue(-3.25))
}

func TestUnaryBindToProperty(t *testing.T) {
	expectProgramToReturn(t, `!!false`, mgnFalse)
	expectProgramToReturn(t, `--3`, IntValue(3))
	expectProgramToReturn(t, `
	obj := {k: false, n: 10}
	[!obj.k, -obj.n]
	`, MakeList(
		mgnTrue,
		IntValue(-10),
	))
}

func TestBasicBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `2 * 3 + 1`, IntValue(7))
	expectProgramToReturn(t, `1 + 2 * 3`, IntValue(7))
}

func TestOrderedBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `-1.5 + -3.5 - 5 / 5 * 2`, FloatValue(-7))
	expectProgramToReturn(t, `(-1.5 + -3.5 - 5) / 5 * 2`, FloatValue(-4))
}

func TestBinaryExprWithParens(t *testing.T) {
	expectProgramToReturn(t, `(1 + 2) / 3 - 1 + (10 + (20 / 5)) % 3`, IntValue(2))
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

func TestIfExprWithAssignmentCond(t *testing.T) {
	expectProgramToReturn(t, `if x := 2 + 4 {
		6 -> x * x
		_ -> x
	}`, IntValue(36))
}

func TestIfExprInFunction(t *testing.T) {
	expectProgramToReturn(t, `
	fn even?(n) if n % 2 {
		0 -> true
		_ -> false
	}
	even?(100)
	`, mgnTrue)
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

// string ops

func TestStringAccess(t *testing.T) {
	expectProgramToReturn(t, `
	s := 'Hello, World!'
	[
		s.0 + s.2
		s.-2
		s.15
	]
	`, MakeList(MakeString("Hl"), null, null))
}

func TestStringAssign(t *testing.T) {
	expectProgramToReturn(t, `
	s := {
		payload: 'Magnolia'
	}
	t := s.payload
	[s.payload.3 := 'pie', t]
	`, MakeList(
		MakeString("Magpieia"),
		MakeString("Magpieia"),
	))
}

func TestStringAppend(t *testing.T) {
	expectProgramToReturn(t, `
	s := {
		payload: 'Magnolia'
	}
	t := s.payload
	[s.payload.(len(t)) := ' language', t]
	`, MakeList(
		MakeString("Magnolia language"),
		MakeString("Magnolia language"),
	))
}

// list ops

func TestListAccess(t *testing.T) {
	expectProgramToReturn(t, `
	s := [1, 2, 3, 4, 5]
	[
		s.0 + s.3
		s.-2
		s.15
	]
	`, MakeList(IntValue(5), null, null))
}

func TestListAssign(t *testing.T) {
	result := MakeList(
		IntValue(1),
		IntValue(2),
		MakeString("three"),
		IntValue(4),
	)

	expectProgramToReturn(t, `
	s := {
		numbers: [1, 2, 3, 4]
	}
	t := s.numbers
	[s.numbers.2 := 'three', t]
	`, MakeList(result, result))
}

func TestListAppend(t *testing.T) {
	result := MakeList(
		IntValue(1),
		IntValue(2),
		IntValue(3),
		IntValue(4),
		IntValue(100),
	)

	expectProgramToReturn(t, `
	s := {
		numbers: [1, 2, 3, 4]
	}
	t := s.numbers
	[s.numbers.(len(t)) := 100, t]
	`, MakeList(result, result))
}

// object ops

func TestObjectAccess(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: ['see', {
			d: 'd'
		}]
	}
	obj.c.(1).d
	`, MakeString("d"))
}

func TestObjectAssign(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: ['see', {
			d: 'd'
		}]
	}
	[
		obj.c.(1).e := 'hello'
		obj.c
	]
	`, MakeList(
		ObjectValue{
			"d": MakeString("d"),
			"e": MakeString("hello"),
		},
		MakeList(MakeString("see"), ObjectValue{
			"d": MakeString("d"),
			"e": MakeString("hello"),
		}),
	))
}

func TestObjectDelete(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: {
			d: 'dee'
			e: 'ee'
		}
	}
	[
		obj.c.d := _
		obj.c
	]
	`, MakeList(
		ObjectValue{
			"e": MakeString("ee"),
		},
		ObjectValue{
			"e": MakeString("ee"),
		},
	))
}

func TestSinglePipe(t *testing.T) {
	expectProgramToReturn(t, `
	fn append(a, b) a + b
	'hello' |> append('world')
	`, MakeString("helloworld"))
}

func TestMultiPipe(t *testing.T) {
	expectProgramToReturn(t, `
	fn append(a, b) a + b
	'hello' |> append('world') |> append('!')
	`, MakeString("helloworld!"))
}

func TestComplexPipe(t *testing.T) {
	expectProgramToReturn(t, `
	lib := {
		add1: fn(n) n + 1
		double: fn(n) 2 * n
	}
	fn getAdder(env) { env.add1 }
	100 |> lib.add1() |> lib.double() |> getAdder(lib)()
	`, IntValue(203))
}

func TestExtraArgs(t *testing.T) {
	expectProgramToReturn(t, `
	fn getExtra(a, b, c) {
		[b, c]
	}
	getExtra(1, ?)
	`, MakeList(null, null))
}

func TestRestArgs(t *testing.T) {
	expectProgramToReturn(t, `
	fn getRest(first, rest...) {
		rest
	}
	getRest(1, 2, 3, 4, 5)
	`, MakeList(
		IntValue(2),
		IntValue(3),
		IntValue(4),
		IntValue(5),
	))
}
