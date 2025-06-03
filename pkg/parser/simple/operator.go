package simple

var validOperators = map[string]struct{}{}

// init valid operators
func init() {
	_validOperators := []string{
		">", ">=", "<", "<=",
		"&", ">>", "<<", "^", "|", "~",
		"<>", "!", "!=", "&&", "||",
		"<=>",
		"%", "+", "-", "*", "/",
		"->", "->>",
		":=", "=",
		"*",
	}
	for _, op := range _validOperators {
		validOperators[op] = struct{}{}
	}
}

func (p *Parser) operator(c string) *Token {
	for {
		if p.isAtEnd() {
			goto ret
		}
		next := p.peek()
		_, ok1 := validOperators[c]
		_, ok2 := validOperators[next]
		if !(ok1 && ok2) {
			goto ret
		}
		c = p.advance()
	}

ret:
	op := p.sql[p.start:p.curr]
	if _, ok := validOperators[op]; ok {
		return &Token{
			Type:   TokenOperator,
			Lexeme: p.sql[p.start:p.curr],
			Pos:    Pos{p.start, p.curr},
		}
	}
	return nil
}
