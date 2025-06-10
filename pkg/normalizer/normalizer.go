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
	KeywordCase              Case
	RemoveLiterals           bool
	PutBacktickOnKeywords    bool
	RemoveBacktickOnKeywords bool
	PutSpaceBeforeOpenParen  bool
}

// Determine if current and previous token is space-able.
func isSpaceAble(config Config, prev lexer.Token, token lexer.Token) bool {
	if prev == (lexer.Token{}) {
		return false
	}
	if config.PutSpaceBeforeOpenParen && token.Type == lexer.TokenOpenParen {
		return true
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

	// Putting space before paren can cause trouble if keywords are not bactick'ed.
	// For example, COUNT (*) is invalid syntax, but COUNT (`*`) and COUNT(*) are valid.
	if config.PutSpaceBeforeOpenParen && !config.PutBacktickOnKeywords {
		return 0, nil, errors.New("PutSpaceBeforeOpenParen requires PutBacktickOnKeywords to be true")
	}

	if config.PutBacktickOnKeywords && config.RemoveBacktickOnKeywords {
		return 0, nil, errors.New("PutBacktickOnKeywords and RemoveBacktickOnKeywords cannot be both true")
	}

	for {
		tok := lex.NextToken()
		if tok.Type == lexer.TokenEOF {
			break
		}

		var n int

		if isSpaceAble(config, prev, tok) {
			n, _ = bytes.PutBytes(result[off:], []byte{' '})
			if n == 0 {
				return off, result[:off], ErrBufferTooSmall
			}
			off += n
		}

		// origLen := len(result)]

		n = 0
		if tok.Type == lexer.TokenLiteral && config.RemoveLiterals {
			n, _ = bytes.PutBytes(result[off:], questionMark)
		} else {
			isBacktickable := tok.IsBacktickAble() || (tok.IsKeyword() && prev.Type == lexer.TokenDot)
			if isBacktickable && config.PutBacktickOnKeywords {
				n, _ = tok.LexemeWithBacktick(sql, result[off:])
			} else if isBacktickable && config.RemoveBacktickOnKeywords {
				n, _ = tok.LexemeWithRemovedBacktick(sql, result[off:])
			} else {
				n, _ = tok.Lexeme(sql, result[off:])
			}
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

		prev = tok
	}
	return off, result[:off], nil
}
