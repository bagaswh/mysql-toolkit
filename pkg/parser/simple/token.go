package simple

type TokenType uint16

const (
	_ TokenType = iota
	TokenOpenParen
	TokenCloseParen
	TokenComma
	TokenKeyword
	TokenLiteral
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
	case TokenOperator:
		return "TokenOperator"
	}
	return "unknown"
}

var EOF = Token{Type: TokenEOF, Pos: Pos{0, 0}}
var EmptyToken Token

type Token struct {
	Type TokenType
	Pos  Pos
}

func (t Token) Lexeme(source []byte) []byte {
	if t.Pos.start >= 0 && t.Pos.end <= len(source) {
		return source[t.Pos.start:t.Pos.end]
	}
	return []byte{}
}
