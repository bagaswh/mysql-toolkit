package simple

func (p *Parser) keyword(c string) *Token {
	if !p.isAlpha(c) {
		return nil
	}

	for {
		c = p.peek()
		if p.isAlpha(c) || c == "_" {
			p.advance()
			continue
		}

		return &Token{
			Type:   TokenKeyword,
			Lexeme: p.sql[p.start:p.curr],
			Pos:    Pos{p.start, p.curr},
		}
	}
}
