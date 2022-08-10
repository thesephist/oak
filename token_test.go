package main

import "testing"

func TestBack(t *testing.T) {
	token := tokenizer{
		source: []rune("hello\nworld"),
		line:   1,
		col:    0,
		index:  6,
	}
	token.back()
	if token.line != 0 && token.col != 5 {
		t.Errorf("Expected line 0 and column 5, received line %d and column %d", token.line, token.col)
	}
}
