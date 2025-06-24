package normalizer

import (
	"errors"

	"github.com/bagaswh/mysql-toolkit/pkg/bytes"
	"github.com/bagaswh/mysql-toolkit/pkg/lexer"
)

type Case byte

const (
	CaseDefault Case = iota
	CaseLower
	CaseUpper
)

type Config struct {
	KeywordCase    Case
	RemoveLiterals bool
	// PutBacktickOnKeywords    bool
	// RemoveBacktickOnKeywords bool
	// PutSpaceBeforeOpenParen bool
}

// Determine if current and previous token is space-able.
func isSpaceAble(config Config, prev lexer.Token, token lexer.Token) bool {
	if prev == (lexer.Token{}) {
		return false
	}
	if token.Type == lexer.TokenComment {
		return false
	}
	if prev.Type == lexer.TokenDot || token.Type == lexer.TokenDot {
		return false
	}
	if prev.IsKeyword() && token.Type == lexer.TokenOpenParen {
		return false
	}
	if token.Type == lexer.TokenComma || token.Type == lexer.TokenCloseParen {
		return false
	}
	if prev.Type == lexer.TokenOpenParen {
		return false
	}
	// if config.PutSpaceBeforeOpenParen && token.Type == lexer.TokenOpenParen {
	// 	return true
	// }
	return true
}

var (
	questionMark = []byte("?")
)

var (
	ErrBufferTooSmall = errors.New("buffer too small")
)

func Normalize(config Config, lex *lexer.Lexer, sql []byte, result []byte) (int, []byte, error) {
	lex.Parse(sql)
	lex.Reset()
	var prev lexer.Token
	off := 0

	for {
		tok := lex.NextToken()
		if tok.Type == lexer.TokenEOF {
			break
		}

		var n int

		if isSpaceAble(config, prev, tok) {
			n = copy(result[off:], []byte{' '})
			if n == 0 {
				return off, result[:off], ErrBufferTooSmall
			}
			off += n
		}

		n = 0
		if tok.Type == lexer.TokenLiteral && config.RemoveLiterals {
			n = copy(result[off:], questionMark)
		} else if tok.Type == lexer.TokenComment {
			// skip comments
			goto end
		} else {
			n, _ = tok.LexemeWithRemovedBacktick(sql, result[off:])
			n, _ = tok.Lexeme(sql, result[off:])
		}
		if n == 0 {
			return off, result[:off], ErrBufferTooSmall
		}
		off += n

		if tok.IsKeyword() {
			if config.KeywordCase == CaseLower {
				bytes.ToLowerInPlace(result[off-n : off])
			} else if config.KeywordCase == CaseUpper {
				bytes.ToUpperInPlace(result[off-n : off])
			}
		}
	end:
		prev = tok
	}
	return off, result[:off], nil
}
