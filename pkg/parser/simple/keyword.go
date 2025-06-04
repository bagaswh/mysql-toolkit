package simple

func (p *Parser) keyword(c byte) Token {
	if c == backtick {
		return p.quotedIdentifier()
	}

	if !isAlpha(c) && !isDigit(c) && c != underscore {
		return Token{}
	}

	start := p.start
	for {
		c = p.peek()
		ahead := p.ahead()
		if isAlpha(c) ||
			isDigit(c) ||
			c == underscore ||
			(c == dot && (isAlpha(ahead) || isDigit(ahead))) {
			p.advance()
			continue
		}
		break
	}
	return Token{
		Type: TokenKeyword,
		Pos:  Pos{start, p.curr},
	}
}

func (p *Parser) quotedIdentifier() Token {
	start := p.start
	escaped := false

	for {
		c := p.advance()
		if c == 0 {
			// Unterminated quoted identifier
			return Token{
				Type: TokenKeyword,
				Pos:  Pos{start, p.curr},
			}
		}

		if c == backtick {
			if escaped {
				escaped = false
				continue
			}
			// Check for escaped backtick (double backtick)
			if p.peek() == backtick {
				escaped = true
				continue
			}
			// End of quoted identifier
			return Token{
				Type: TokenKeyword,
				Pos:  Pos{start, p.curr},
			}
		}

		escaped = false
	}
}
