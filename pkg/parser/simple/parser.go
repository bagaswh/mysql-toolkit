package simple

type state byte

type Pos struct {
	start, end int
}

type Parser struct {
	// literalsPos []Pos
	// keywordsPos []Pos
	// commas      []Pos
	// openParens  []Pos
	// closeParens []Pos
	// operators   []Pos

	sql []byte

	// small "arena"
	arena []byte

	start, curr int

	onToken func(Token, Token) bool
}

func NewParser() *Parser {
	return &Parser{
		arena: make([]byte, 1024),
	}
}

func (p *Parser) Parse(sql []byte) {
	p.sql = sql
}

func (p *Parser) Reset() {
	p.curr = 0
}

func (p *Parser) GetLexeme(token Token) []byte {
	return token.Lexeme(p.sql)
}

func (p *Parser) NextToken() Token {
	for p.curr < len(p.sql) {
		p.start = p.curr
		tok := p.scanToken()
		if tok != (Token{}) {
			return tok
		}
	}
	return EOF
}

// func (p *Parser) scanTokens() {
// 	var lastTok Token
// 	sql := p.sql
// 	for p.curr < len(sql) {
// 		p.start = p.curr
// 		tok := p.scanToken(sql)
// 		if tok != (Token{}) {
// 			if tok.Type == TokenEOF {
// 				return
// 			}
// 			if p.onToken != nil && !p.onToken(lastTok, tok) {
// 				return
// 			}
// 			lastTok = tok
// 		}
// 	}
// }

func (p *Parser) scanToken() Token {
	c := p.advance()
	switch c {
	case ',':
		pos := Pos{p.curr - 1, p.curr}
		// p.commas = append(p.commas, pos)
		return Token{Type: TokenComma, Pos: pos}
	case '(':
		pos := Pos{p.curr - 1, p.curr}
		// p.openParens = append(p.openParens, pos)
		return Token{Type: TokenOpenParen, Pos: pos}
	case ')':
		pos := Pos{p.curr - 1, p.curr}
		// p.closeParens = append(p.closeParens, pos)
		return Token{Type: TokenCloseParen, Pos: pos}
	case 0:
		return EOF
	default:
		if isWhiteSpace(c) {
			return Token{}
		}

		// try literal
		tok := p.literal(c)
		if tok != (Token{}) {
			// p.literalsPos = append(p.literalsPos, tok.Pos)
			return tok
		}

		tok = p.keyword(c)
		if tok != (Token{}) {
			// p.keywordsPos = append(p.keywordsPos, tok.Pos)
			return tok
		}

		// try operator
		tok = p.operator(c)
		if tok != (Token{}) {
			// p.operators = append(p.operators, tok.Pos)
			return tok
		}
	}
	return (Token{})
}

func (p *Parser) advance() byte {
	if p.isAtEnd() {
		return 0
	}
	c := p.sql[p.curr]
	p.curr++
	return c
}

func (p *Parser) peek() byte {
	if p.isAtEnd() {
		return 0
	}
	return p.sql[p.curr]
}

func (p *Parser) ahead() byte {
	if p.curr+1 >= len(p.sql) {
		return 0
	}
	return p.sql[p.curr+1]
}

func (p *Parser) aheadN(n int) []byte {
	if p.curr+n >= len(p.sql) {
		return nil
	}
	return p.sql[p.curr+1 : p.curr+n+1]
}

func (p *Parser) isAtEnd() bool {
	return p.curr >= (len(p.sql))
}
