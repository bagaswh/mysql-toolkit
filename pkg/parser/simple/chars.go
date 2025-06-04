package simple

import (
	"bytes"
	"slices"
)

var (
	backtick    = byte('`')
	singleQuote = byte('\'')
	doubleQuote = byte('"')

	underscore = byte('_')
	dot        = byte('.')
	plus       = byte('+')

	backslash = byte('\\')
	dash      = byte('-')

	whitespaces = []byte(" \t\n\r")

	bytes_SQLKeyword_NULL  = []byte("null")
	bytes_SQLKeyword_True  = []byte("true")
	bytes_SQLKeyword_False = []byte("false")
)

func maybeLiteral(c byte) bool {
	switch c {
	case 'n', 'N', 't', 'T', 'f', 'F',
		'b', 'B', 'x', 'X', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		singleQuote, doubleQuote,
		dash, plus, dot:
		return true
	// to prevent from trying to lex these as literals, making it faster
	// in my case, it's 2x faster when benchmarking with ComplexQuery (see parser_test.go)
	default:
		return false
	}
}

func isEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}

func isAlpha(c byte) bool {
	return ((c >= 65 && c <= 90) || (c >= 97 && c <= 122))
}

func isDigit(c byte) bool {
	return c >= 48 && c <= 57
}

func isWhiteSpace(c byte) bool {
	if slices.Contains(whitespaces, c) {
		return true
	}
	return false
}
