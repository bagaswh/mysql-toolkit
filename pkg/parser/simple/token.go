package simple

type TokenType uint16

const (
	TokenOpenParen TokenType = iota
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

var EOF = &Token{Type: TokenEOF, Lexeme: "", Pos: Pos{0, 0}}

type Token struct {
	Type   TokenType
	Lexeme string
	Pos    Pos
}
