package simple

type state byte

type Pos struct {
	start, end int
}

type Parser struct {
	literalsPos []Pos
	keywordsPos []Pos
	commas      []Pos
	openParens  []Pos
	closeParens []Pos
	operators   []Pos

	sql         string
	start, curr int

	state state

	onToken func(*Token, *Token) bool
}

func NewParser(onToken func(prev, curr *Token) bool) *Parser {
	return &Parser{
		onToken: onToken,
	}
}

func (p *Parser) Parse(sql string) {
	p.sql = sql
	p.scanTokens()
}

func (p *Parser) Reset() {
	p.curr = 0
}

func (p *Parser) scanTokens() {
	var lastTok *Token
	for !p.isAtEnd() {
		p.start = p.curr
		tok := p.scanToken()
		if tok != nil {
			if tok.Type == TokenEOF {
				return
			}
			if !p.onToken(lastTok, tok) {
				return
			}
			lastTok = tok
		}
	}
}

func (p *Parser) scanToken() *Token {
	c := p.advance()
	switch c {
	case ",":
		pos := Pos{p.curr, p.curr + 1}
		p.commas = append(p.commas, pos)
		return &Token{Type: TokenComma, Lexeme: c, Pos: pos}
	case "(":
		pos := Pos{p.curr, p.curr + 1}
		p.openParens = append(p.openParens, pos)
		return &Token{Type: TokenOpenParen, Lexeme: c, Pos: pos}
	case ")":
		pos := Pos{p.curr, p.curr + 1}
		p.closeParens = append(p.closeParens, pos)
		return &Token{Type: TokenCloseParen, Lexeme: c, Pos: pos}
	case "":
		return EOF
	default:
		if p.isWhiteSpace(c) {
			return nil
		}

		// try literal
		tok := p.literal(c)
		if tok != nil {
			p.literalsPos = append(p.literalsPos, tok.Pos)
			return tok
		}

		tok = p.keyword(c)
		if tok != nil {
			p.keywordsPos = append(p.keywordsPos, tok.Pos)
			return tok
		}

		// try operator
		tok = p.operator(c)
		if tok != nil {
			p.operators = append(p.operators, tok.Pos)
			return tok
		}
	}
	return nil
}

func (p *Parser) advance() string {
	if p.isAtEnd() {
		return ""
	}
	c := p.sql[p.curr]
	p.curr++
	return string(c)
}

func (p *Parser) peek() string {
	if p.isAtEnd() {
		return ""
	}
	return string(p.sql[p.curr])
}

func (p *Parser) prev() string {
	if p.curr-1 < 0 {
		return ""
	}
	return string(p.sql[p.curr-1])
}

func (p *Parser) ahead() string {
	if p.isAtEnd() || p.curr+1 >= len(p.sql) {
		return ""
	}
	return string(p.sql[p.curr+1])
}

func (p *Parser) aheadN(n int) string {
	if p.isAtEnd() || p.curr+n >= len(p.sql) {
		return ""
	}
	return string(p.sql[p.curr+1 : p.curr+n+1])
}

func (p *Parser) isAlpha(c string) bool {
	if len(c) == 0 {
		return false
	}
	return ((c[0] >= 65 && c[0] <= 90) || (c[0] >= 97 && c[0] <= 122))
}

func (p *Parser) isDigit(c string) bool {
	if len(c) == 0 {
		return false
	}
	return c[0] >= 48 && c[0] <= 57
}

func (p *Parser) isWhiteSpace(c string) bool {
	if len(c) == 0 {
		return false
	}
	switch c[0] {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func (p *Parser) isAtEnd() bool {
	return p.curr >= (len(p.sql))
}
