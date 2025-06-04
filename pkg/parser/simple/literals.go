package simple

import (
	"bytes"
)

func (p *Parser) literal(c byte) Token {
	if !maybeLiteral(c) {
		return Token{}
	}

	curr := p.peek()

	if c == singleQuote || c == doubleQuote {
		return p.stringLiteral(c)
	}
	if (c == 'b' || c == 'B') && curr == singleQuote {
		p.advance()
		return p.bitValueLiterals("b'")
	}
	if c == '0' && (curr == 'B' || curr == 'b') {
		p.advance()
		return p.bitValueLiterals("0b")
	}
	if (c == 'X' || c == 'x') && curr == singleQuote {
		p.advance()
		return p.hexLiteral("x'")
	}
	if c == '0' && (curr == 'X' || curr == 'x') {
		p.advance()
		return p.hexLiteral("0x")
	}
	if ((c == dash || c == plus || c == dot) && isDigit(curr)) ||
		isDigit(c) {
		return p.numberLiteral()
	}

	var tok Token
	advanceN := 0
	p.arena = p.arena[:5]
	copy(p.arena[:2], []byte{c, curr})
	copy(p.arena[2:], p.aheadN(2))
	toLowerInPlace(p.arena)

	if bytes.HasPrefix(p.arena, bytes_SQLKeyword_NULL) ||
		bytes.HasPrefix(p.arena, bytes_SQLKeyword_True) {
		advanceN = 3
		tok = Token{
			Type: TokenLiteral,
			Pos:  Pos{p.start, p.curr + 3},
		}
		goto ret
	}

	copy(p.arena[2:], p.aheadN(3))
	toLowerInPlace(p.arena)

	if bytes.HasPrefix(p.arena, bytes_SQLKeyword_False) {
		advanceN = 4
		tok = Token{
			Type: TokenLiteral,
			Pos:  Pos{p.start, p.curr + 4},
		}
		goto ret
	}
ret:
	if tok != (Token{}) {
		p.curr += advanceN
		return tok
	}

	return Token{}
}

func (p *Parser) stringLiteral(c byte) Token {
	quoteType := c // either ' or "
	escaped := false

	for {
		c := p.advance()
		if c == 0 {
			// Unterminated string literal
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{p.start, p.curr},
			}
		}

		if c == backslash {
			escaped = !escaped
			continue
		}

		if c == quoteType && !escaped {
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{p.start, p.curr},
			}
		}

		escaped = false
	}
}

func (p *Parser) numberLiteral() Token {
	for {
		c := p.advance()
		curr := p.peek()
		if (isDigit(c) ||
			(c == dash && isDigit(curr)) ||
			(c == dot && isDigit(curr))) ||
			(c == 'E' && (isDigit(curr) || curr == dash)) {
			continue
		}
		if !isDigit(c) {
			return Token{
				Type: TokenLiteral,
				Pos:  Pos{p.start, p.curr},
			}
		}
	}
}

func (p *Parser) bitValueLiterals(start string) Token {
	var tok Token
	for {
		c := p.advance()
		if start == "b'" {
			if c == '\'' {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{p.start, p.curr},
				}
				break
			}
		} else if start == "0b" {
			if isWhiteSpace(c) || p.isAtEnd() {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{p.start, p.curr},
				}
				break
			}
		}
		if p.isAtEnd() {
			break
		}
	}
	return tok
}

func (p *Parser) hexLiteral(hexStart string) Token {
	var tok Token
	for {
		c := p.advance()
		if hexStart == "x'" {
			if c == '\'' {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{p.start, p.curr},
				}
				break
			}
		} else if hexStart == "0x" {
			if isWhiteSpace(c) || p.isAtEnd() {
				tok = Token{
					Type: TokenLiteral,
					Pos:  Pos{p.start, p.curr},
				}
				break
			}
		}
		if p.isAtEnd() {
			break
		}
	}
	return tok
}
