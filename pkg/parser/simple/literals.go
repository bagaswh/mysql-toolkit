package simple

import (
	"strings"
)

func (p *Parser) literal(c string) *Token {
	curr := p.peek()
	if c == "'" || c == "\"" {
		return p.stringLiteral()
	}
	if (c == "b" || c == "B") && curr == "'" {
		p.advance()
		return p.bitValueLiterals("b'")
	}
	if c == "0" && (curr == "B" || curr == "b") {
		p.advance()
		return p.bitValueLiterals("0b")
	}
	if (c == "X" || c == "x") && curr == "'" {
		p.advance()
		return p.hexLiteral("x'")
	}
	if c == "0" && (curr == "X" || curr == "x") {
		p.advance()
		return p.hexLiteral("0x")
	}
	if ((c == "-" || c == "+" || c == ".") && p.isDigit(curr)) ||
		p.isDigit(c) {
		return p.numberLiteral()
	}

	var tok *Token
	advanceN := 0
	ahead := c + curr + p.aheadN(2)

	if ahead == "NULL" || strings.ToLower(ahead) == "true" {
		advanceN = 3
		tok = &Token{
			Type:   TokenLiteral,
			Lexeme: p.sql[p.start : p.curr+3],
			Pos:    Pos{p.start, p.curr},
		}
		goto ret
	}

	ahead = c + curr + p.aheadN(3)
	if strings.ToLower(ahead) == "false" {
		advanceN = 4
		tok = &Token{
			Type:   TokenLiteral,
			Lexeme: p.sql[p.start : p.curr+4],
			Pos:    Pos{p.start, p.curr},
		}
		goto ret
	}
ret:
	if tok != nil {
		for i := 0; i < advanceN; i++ {
			p.advance()
		}
		return tok
	}

	return nil
}

func (p *Parser) stringLiteral() *Token {
	for {
		c := p.advance()
		if (c == "'" || c == "\"") && p.prev() != "\\" {
			return &Token{
				Type:   TokenLiteral,
				Lexeme: p.sql[p.start:p.curr],
				Pos:    Pos{p.start, p.curr},
			}
		}
	}
}

func (p *Parser) numberLiteral() *Token {
	for {
		c := p.advance()
		curr := p.peek()
		if (p.isDigit(c) ||
			(c == "-" && p.isDigit(curr)) ||
			(c == "." && p.isDigit(curr))) ||
			(c == "E" && (p.isDigit(curr) || curr == "-")) {
			continue
		}
		if !p.isDigit(c) {
			return &Token{
				Type:   TokenLiteral,
				Lexeme: p.sql[p.start:p.curr],
				Pos:    Pos{p.start, p.curr},
			}
		}
	}
}

func (p *Parser) bitValueLiterals(start string) *Token {
	var tok *Token
	for {
		c := p.advance()
		if start == "b'" {
			if c == "'" {
				tok = &Token{
					Type:   TokenLiteral,
					Lexeme: p.sql[p.start:p.curr],
					Pos:    Pos{p.start, p.curr},
				}
				break
			}
		} else if start == "0b" {
			if p.isWhiteSpace(c) || p.isAtEnd() {
				tok = &Token{
					Type:   TokenLiteral,
					Lexeme: p.sql[p.start:p.curr],
					Pos:    Pos{p.start, p.curr},
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

func (p *Parser) hexLiteral(hexStart string) *Token {
	var tok *Token
	for {
		c := p.advance()
		if hexStart == "x'" {
			if c == "'" {
				tok = &Token{
					Type:   TokenLiteral,
					Lexeme: p.sql[p.start:p.curr],
					Pos:    Pos{p.start, p.curr},
				}
				break
			}
		} else if hexStart == "0x" {
			if p.isWhiteSpace(c) || p.isAtEnd() {
				tok = &Token{
					Type:   TokenLiteral,
					Lexeme: p.sql[p.start:p.curr],
					Pos:    Pos{p.start, p.curr},
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
