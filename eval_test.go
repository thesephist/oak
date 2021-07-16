package main

import (
	"fmt"
	"strings"
	"testing"
)

func expectProgramToReturn(t *testing.T, program string, expected Value) {
	ctx := NewContext("<input>", "/tmp")
	val, err := ctx.Eval(strings.NewReader(program))
	if err != nil {
		t.Errorf("Did not expect program to exit with an error")
	}
	if val == nil {
		t.Errorf("Return valuue of program should not be nil")
	}
	if !val.Eq(expected) {
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
	expectProgramToReturn(t, "'Hello, World!\\n'", StringValue([]byte("Hello, World!\n")))
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

// TODO: write some more tests covering prog3, prog4, and prog5 which are
// roughly if expressions, nested functions, recursion, and binary expressions
// (equality), etc.

func TestBasicBinaryExpr(t *testing.T) {
}

func TestBinaryExprWithParens(t *testing.T) {
}

func TestLongBinaryExprWithPrecedence(t *testing.T) {
}

func TestBinaryExprWithComplexTerms(t *testing.T) {
	// complex terms are function calls, property accesses
}

func TestBasicIfExpr(t *testing.T) {
}

func TestIfExprWithEmpty(t *testing.T) {
}

func TestIfExprInFunction(t *testing.T) {
}

func TestRecursiveFunction(t *testing.T) {
}

func TestBasicWithExpr(t *testing.T) {
	expectProgramToReturn(t, `fn add(a, b) { a + b }, with add(10) 40`, IntValue(50))
}

func TestWithExprWithCallback(t *testing.T) {
}
