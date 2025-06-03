package simple

func (p *Parser) keyword(c string) *Token {
	if c == "`" {
		return p.quotedIdentifier()
	}

	if !p.isAlpha(c) && !p.isDigit(c) && c != "_" {
		return nil
	}

	start := p.start
	for {
		c = p.peek()
		ahead := p.ahead()
		if p.isAlpha(c) ||
			p.isDigit(c) ||
			c == "_" ||
			(c == "." && (p.isAlpha(ahead) || p.isDigit(ahead))) {
			p.advance()
			continue
		}
		break
	}
	return &Token{
		Type:   TokenKeyword,
		Lexeme: p.sql[start:p.curr],
		Pos:    Pos{start, p.curr},
	}
}

func (p *Parser) quotedIdentifier() *Token {
	start := p.start
	escaped := false

	for {
		c := p.advance()
		if c == "" {
			// Unterminated quoted identifier
			return &Token{
				Type:   TokenKeyword,
				Lexeme: p.sql[start:p.curr],
				Pos:    Pos{start, p.curr},
			}
		}

		if c == "`" {
			if escaped {
				escaped = false
				continue
			}
			// Check for escaped backtick (double backtick)
			if p.peek() == "`" {
				escaped = true
				continue
			}
			// End of quoted identifier
			return &Token{
				Type:   TokenKeyword,
				Lexeme: p.sql[start:p.curr],
				Pos:    Pos{start, p.curr},
			}
		}

		escaped = false
	}
}
