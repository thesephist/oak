package main

import (
	"strings"
	"testing"
)

func newTestContext() Context {
	return NewContext("<input>", "/tmp")
}

func TestEvalEmptyProgram(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader(""))
	if err != nil || val == nil || !val.Eq(null) {
		t.Errorf("Expected empty program to evaluate to ?")
	}
}

func TestEmptyLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader("_"))
	if err != nil || val == nil || !val.Eq(empty) {
		t.Errorf("Expected empty value to evaluate to _")
	}
}

func TestNullLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader("?"))
	if err != nil || val == nil || !val.Eq(null) {
		t.Errorf("Expected null value to evaluate to ?")
	}
}

func TestStringLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader("'Hello, World!\\n'"))
	if err != nil || val == nil || !val.Eq(StringValue([]byte("Hello, World!\n"))) {
		t.Errorf("Expected string literal to parse and evaluate correctly")
	}
}

func TestIntegerLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader("12314"))
	if err != nil || val == nil || !val.Eq(IntValue(12314)) {
		t.Errorf("Expected integer literal to parse and evaluate correctly")
	}
}

func TestFloatLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader("3.141592"))
	if err != nil || val == nil || !val.Eq(FloatValue(3.141592)) {
		t.Errorf("Expected float literal to parse and evaluate correctly")
	}
}

func TestAtomLiteral(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader(":not_found_404"))
	if err != nil || val == nil || !val.Eq(AtomValue("not_found_404")) {
		t.Errorf("Expected basic atom to parse and evaluate correctly")
	}
}

func TestFunctionDefAndCall(t *testing.T) {
	ctx := newTestContext()
	val, err := ctx.Eval(strings.NewReader(`fn getThree() { x := 4, 3 }, getThree()`))
	if err != nil || val == nil || !val.Eq(IntValue(3)) {
		t.Errorf("Expected basic function definition and call to work")
	}
}
