package lexer

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
