package lexer

import (
	"github.com/bagaswh/mysql-toolkit/pkg/bytes"
)

type TokenType uint16

const (
	_ TokenType = iota
	TokenOpenParen
	TokenCloseParen
	TokenComma
	TokenDot
	TokenStar
	TokenComment
	TokenKeyword
	TokenLiteral
	TokenBackslash
	TokenOperator
	TokenEOF
)

func (t TokenType) String() string {
	switch t {
	case TokenOpenParen:
		return "TokenOpenParen"
	case TokenCloseParen:
		return "TokenCloseParen"
	case TokenComma:
		return "TokenComma"
	case TokenKeyword:
		return "TokenKeyword"
	case TokenLiteral:
		return "TokenLiteral"
	case TokenBackslash:
		return "TokenBackslash"
	case TokenOperator:
		return "TokenOperator"
	case TokenStar:
		return "TokenStar"
	case TokenComment:
		return "TokenComment"
	case TokenEOF:
		return "TokenEOF"
	}
	return "unknown"
}

var EOF = Token{Type: TokenEOF, Pos: Pos{0, 0}}
var EmptyToken Token

type Attr uint64

const (
	TokenAttrBuiltIn Attr = 1 << iota
	TokenKeywordIdentifierWithDot
)

type Token struct {
	Pos  Pos
	Attr Attr
	Type TokenType
}

func (t Token) Lexeme(source []byte, result []byte) (int, []byte) {
	if t.Pos.start >= 0 && t.Pos.end <= len(source) && t.Pos.end > t.Pos.start {
		return bytes.PutBytes(result, source[t.Pos.start:t.Pos.end])
	}
	return 0, result
}

func (t Token) LexemeRef(source []byte) []byte {
	if t.Pos.start >= 0 && t.Pos.end <= len(source) && t.Pos.end > t.Pos.start {
		return source[t.Pos.start:t.Pos.end]
	}
	return []byte{}
}

func (t Token) LexemeLen() int {
	return t.Pos.end - t.Pos.start
}

func (t Token) IsKeyword() bool {
	return t.Type == TokenKeyword
}

func (t Token) IsLiteral() bool {
	return t.Type == TokenLiteral
}

func (t Token) IsBuiltInKeyword() bool {
	return t.IsKeyword() && t.Attr&TokenAttrBuiltIn != 0
}

func (t Token) IsIdentifier() bool {
	return t.IsKeyword() && !t.IsBuiltInKeyword()
}

func (t Token) IsBuiltInFunction() bool {
	return t.IsKeyword() && t.Attr&KeywordAttrBuiltInFunction != 0
}

func (t Token) IsQuotedWithBacktick(source []byte) bool {
	if t.Pos.start >= 0 && t.Pos.end <= len(source) && t.Pos.end > t.Pos.start {
		return source[t.Pos.start] == '`' && source[t.Pos.end-1] == '`'
	}
	return false
}

func (t Token) LexemeWithRemovedBacktick(source []byte, result []byte) (int, []byte) {
	if t.Pos.end-t.Pos.start > 1 {
		if t.IsQuotedWithBacktick(source) {
			return bytes.PutBytes(result, source[t.Pos.start+1:t.Pos.end-1])
		}
	}
	return t.Lexeme(source, result)
}

func (t Token) LexemeWithBacktick(source []byte, result []byte) (int, []byte) {
	if t.Pos.end-t.Pos.start > 1 {
		if t.IsQuotedWithBacktick(source) {
			return t.Lexeme(source, result)
		}
	}
	return bytes.PutBytes(result, []byte{'`'}, source[t.Pos.start:t.Pos.end], []byte{'`'})
}

func (t Token) IsBacktickAble() bool {
	return t.IsIdentifier()
}
