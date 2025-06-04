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

func (p *Parser) operator(c byte) Token {
	for {
		// if p.isAtEnd() {
		// 	goto ret
		// }
		next := p.peek()
		// the string conversion here is free
		// source: https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html#using_byte_as_a_map_key
		_, ok1 := validOperators[string(c)]
		_, ok2 := validOperators[string(next)]
		if !(ok1 && ok2) {
			goto ret
		}
		c = p.advance()
	}

ret:
	op := p.sql[p.start:p.curr]
	if _, ok := validOperators[string(op)]; ok {
		return Token{
			Type: TokenOperator,
			Pos:  Pos{p.start, p.curr},
		}
	}
	return Token{}
}
