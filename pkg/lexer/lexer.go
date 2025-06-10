package lexer

import "github.com/bagaswh/mysql-toolkit/pkg/bytes"

type state byte

type Pos struct {
	start, end int
}

type Lexer struct {
	sql []byte

	// reusable buffer
	arena []byte

	start, curr int
}

func NewLexer() *Lexer {
	return &Lexer{}
}

func (l *Lexer) Parse(sql []byte) {
	l.sql = sql
	l.initArena()
}

func (l *Lexer) initArena() {
	if l.arena == nil || cap(l.arena) < len(l.sql) {
		l.arena = make([]byte, 0, max(len(l.sql), 1024))
	}
}

func (l *Lexer) Reset() {
	l.curr = 0
}

func (l *Lexer) GetLexeme(token Token) []byte {
	return token.LexemeRef(l.sql)
}

func (l *Lexer) GetLexemeCopy(token Token) []byte {
	result := make([]byte, token.LexemeLen())
	_, result = token.Lexeme(l.sql, result)
	return result
}

func (l *Lexer) GetLexemeNoAlloc(token Token, result []byte) (int, []byte) {
	return token.Lexeme(l.sql, result)
}

func (l *Lexer) NextToken() Token {
	for l.curr < len(l.sql) {
		l.start = l.curr
		tok := l.scanToken()
		if tok != (Token{}) {
			return tok
		}
	}
	return EOF
}

func (l *Lexer) scanToken() Token {
	ch := l.advance()
	if ch == nil {
		return EOF
	}
	c := ch[0]
	switch c {
	case '*':
		pos := Pos{l.curr - 1, l.curr}
		return Token{Type: TokenStar, Pos: pos}
	case ',':
		pos := Pos{l.curr - 1, l.curr}
		return Token{Type: TokenComma, Pos: pos}
	case '(':
		pos := Pos{l.curr - 1, l.curr}
		return Token{Type: TokenOpenParen, Pos: pos}
	case ')':
		pos := Pos{l.curr - 1, l.curr}
		return Token{Type: TokenCloseParen, Pos: pos}
	case 0:
		return EOF
	default:
		if isWhiteSpace(c) {
			return Token{}
		}

		// try literal
		tok := l.literal(c)
		if tok != (Token{}) {
			return tok
		}

		if c == '.' {
			pos := Pos{l.curr - 1, l.curr}
			return Token{Type: TokenDot, Pos: pos}
		}

		tok = l.keyword(c)
		if tok != (Token{}) {
			return tok
		}

		// try operator
		tok = l.operator(c)
		if tok != (Token{}) {
			return tok
		}
	}
	return (Token{})
}

func (l *Lexer) operator(c byte) Token {
	for {
		// if l.isAtEnd() {
		// 	goto ret
		// }
		next := l.peek()
		// the string conversion here is free
		// source: https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html#using_byte_as_a_map_key
		_, ok1 := validOperators[string(c)]
		_, ok2 := validOperators[string(next)]
		if !(ok1 && ok2) {
			goto ret
		}
		ch := l.advance()
		if ch == nil {
			break
		}
	}

ret:
	op := l.sql[l.start:l.curr]
	if _, ok := validOperators[string(op)]; ok {
		return Token{
			Type: TokenOperator,
			Pos:  Pos{l.start, l.curr},
		}
	}
	return Token{}
}

func (l *Lexer) keyword(c byte) Token {
	if c == backtick {
		return l.quotedIdentifier()
	}

	if !isAlpha(c) {
		return Token{}
	}

	var attr Attr

	start := l.start
	for {
		c = l.peek()
		// ahead := l.ahead()
		if isAlpha(c) ||
			isDigit(c) ||
			c == underscore {
			l.advance()
			continue
		}
		// if c == dot && (isAlpha(ahead) || isDigit(ahead)) {
		// 	// dots in keyword indicate it's a db/table/column reference
		// 	l.advance()
		// 	continue
		// }
		break
	}
	return Token{
		Type: TokenKeyword,
		Pos:  Pos{start, l.curr},
		Attr: l.getKeywordAttr(attr, start),
	}
}

func (l *Lexer) getKeywordAttr(currentAttr Attr, start int) Attr {
	var attr Attr
	arena := l.arena[:l.curr-start]
	bytes.PutBytes(arena, l.sql[start:l.curr])
	bytes.ToUpperInPlace(arena)
	if keyTyp, ok := isBuiltInKeyword(arena); ok {
		attr |= TokenAttrBuiltIn
		attr |= keyTyp.Attr
	}
	return currentAttr | attr
}

func (l *Lexer) quotedIdentifier() Token {
	start := l.start
	escaped := false

	var attr Attr

	for {
		ch := l.advance()
		if ch == nil {
			return Token{
				Type: TokenKeyword,
				Pos:  Pos{},
				Attr: l.getKeywordAttr(attr, start),
			}
		}
		c := ch[0]

		if c == backtick {
			if escaped {
				escaped = false
				continue
			}
			// Check for escaped backtick (double backtick)
			if l.peek() == backtick {
				escaped = true
				continue
			}
			// End of quoted identifier
			return Token{
				Type: TokenKeyword,
				Pos:  Pos{start, l.curr},
				Attr: l.getKeywordAttr(attr, start),
			}
		}

		escaped = false
	}
}

func (l *Lexer) literal(c byte) Token {

	curr := l.peek()

	if c == singleQuote || c == doubleQuote {
		return l.stringLiteral(c)
	}
	if (c == 'b' || c == 'B') && curr == singleQuote {
		l.advance()
		return l.bitValueLiterals("b'")
	}
	if c == '0' && (curr == 'B' || curr == 'b') {
		l.advance()
		return l.bitValueLiterals("0b")
	}
	if (c == 'X' || c == 'x') && curr == singleQuote {
		l.advance()
		return l.hexLiteral("x'")
	}
	if c == '0' && (curr == 'X' || curr == 'x') {
		l.advance()
		return l.hexLiteral("0x")
	}
	if ((c == dash || c == plus || c == dot) && isDigit(curr)) ||
		isDigit(c) {
		return l.numberLiteral()
	}
	return Token{}
}

func (l *Lexer) stringLiteral(c byte) Token {
	quoteType := c // either ' or "
	escaped := false

	for {
		ch := l.advance()
		if ch == nil {
			// Unterminated string literal
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{l.start, l.curr},
			}
		}
		c := ch[0]

		if c == backslash {
			escaped = !escaped
			continue
		}

		if c == quoteType && !escaped {
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{l.start, l.curr},
			}
		}

		escaped = false
	}
}

func (l *Lexer) numberLiteral() Token {
	for {
		curr := l.peek()
		currIsDigit := isDigit(curr)
		next := l.ahead()
		nextIsDigit := isDigit(next)
		if (currIsDigit ||
			(curr == dash && nextIsDigit) ||
			(curr == dot && nextIsDigit)) ||
			((curr == 'E' || curr == 'e') && (nextIsDigit || next == dash)) {
			l.stepForward()
			continue
		}
		if !currIsDigit {
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{l.start, l.curr},
			}
		}
	}
}

func (l *Lexer) bitValueLiterals(start string) Token {
	var tok Token
	for {
		ch := l.advance()
		if ch == nil {
			break
		}
		c := ch[0]
		if start == "b'" {
			if c == '\'' {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{l.start, l.curr},
				}
				break
			}
		} else if start == "0b" {
			if isWhiteSpace(c) || l.isAtEnd() {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{l.start, l.curr},
				}
				break
			}
		}
		if l.isAtEnd() {
			break
		}
	}
	return tok
}

func (l *Lexer) hexLiteral(hexStart string) Token {
	var tok Token
	for {
		ch := l.advance()
		if ch == nil || (ch[0] == '\'' && hexStart == "x'") || (hexStart == "0x" && isWhiteSpace(ch[0])) {
			tok = Token{
				Type: TokenLiteral,
				Pos:  Pos{l.start, l.curr},
			}
			break
		}
	}
	return tok
}

func (l *Lexer) stepForward() {
	if l.isAtEnd() {
		return
	}
	l.curr++
}

func (l *Lexer) advance() []byte {
	if l.isAtEnd() {
		return nil
	}
	// TODO: Why this is not BCE'd?
	c := l.sql[l.curr]
	l.curr++
	return []byte{c}
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.sql[l.curr]
}

func (l *Lexer) ahead() byte {
	if l.curr+1 >= len(l.sql) {
		return 0
	}
	return l.sql[l.curr+1]
}

func (l *Lexer) isAtEnd() bool {
	return l.curr >= (len(l.sql))
}

var (
	backtick    = byte('`')
	singleQuote = byte('\'')
	doubleQuote = byte('"')

	underscore = byte('_')
	dot        = byte('.')
	plus       = byte('+')

	backslash = byte('\\')
	dash      = byte('-')

	maybeLiterals = [][]byte{
		[]byte("n"), []byte("NU"), []byte("tr"), []byte("TR"), []byte("fa"), []byte("FA"),
	}
)

func isAlpha(c byte) bool {
	return ((c >= 65 && c <= 90) || (c >= 97 && c <= 122))
}

func isDigit(c byte) bool {
	return c-48 <= 9
}

func isWhiteSpace(c byte) bool {
	return c == '\t' || c == '\n' || c == '\r' || c == ' '
}
