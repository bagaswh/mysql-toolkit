package simple

import (
	"reflect"
	"testing"
)

func TestParser_BasicSQL(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	p.Parse("SELECT * FROM users WHERE id = 1")

	expected := []struct {
		tokenType TokenType
		lexeme    string
	}{
		{TokenKeyword, "SELECT"},
		{TokenOperator, "*"},
		{TokenKeyword, "FROM"},
		{TokenKeyword, "users"},
		{TokenKeyword, "WHERE"},
		{TokenKeyword, "id"},
		{TokenOperator, "="},
		{TokenLiteral, "1"},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i].tokenType {
			t.Errorf("Token %d: expected type %s, got %s", i, expected[i].tokenType.String(), tok.Type.String())
		}
		if tok.Lexeme != expected[i].lexeme {
			t.Errorf("Token %d: expected lexeme '%s', got '%s'", i, expected[i].lexeme, tok.Lexeme)
		}
	}
}

func TestParser_StringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single quotes",
			input:    "SELECT 'hello world'",
			expected: []string{"SELECT", "'hello world'"},
		},
		{
			name:     "Double quotes",
			input:    "SELECT \"hello world\"",
			expected: []string{"SELECT", "\"hello world\""},
		},
		{
			name:     "Mixed quotes",
			input:    "SELECT 'single' AND \"double\"",
			expected: []string{"SELECT", "'single'", "AND", "\"double\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, tok.Lexeme)
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestParser_NumberLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Integer",
			input:    "SELECT 123",
			expected: []string{"SELECT", "123"},
		},
		{
			name:     "Float",
			input:    "SELECT 123.456",
			expected: []string{"SELECT", "123.456"},
		},
		{
			name:     "Negative number",
			input:    "SELECT -42",
			expected: []string{"SELECT", "-42"},
		},
		{
			name:     "Positive number",
			input:    "SELECT +42",
			expected: []string{"SELECT", "+42"},
		},
		{
			name:     "Scientific notation",
			input:    "SELECT 1.23E-4",
			expected: []string{"SELECT", "1.23E-4"},
		},
		{
			name:     "Decimal starting with dot",
			input:    "SELECT .123",
			expected: []string{"SELECT", ".123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, tok.Lexeme)
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestParser_Operators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "Comparison operators",
			input:    "> >= < <= <> !=",
			expected: []TokenType{TokenOperator, TokenOperator, TokenOperator, TokenOperator, TokenOperator, TokenOperator},
		},
		{
			name:     "Arithmetic operators",
			input:    "+ - * / %",
			expected: []TokenType{TokenOperator, TokenOperator, TokenOperator, TokenOperator, TokenOperator},
		},
		{
			name:     "Bitwise operators",
			input:    "& | ^ ~ >> <<",
			expected: []TokenType{TokenOperator, TokenOperator, TokenOperator, TokenOperator, TokenOperator, TokenOperator},
		},
		{
			name:     "JSON operators",
			input:    "-> ->>",
			expected: []TokenType{TokenOperator, TokenOperator},
		},
		{
			name:     "Assignment operators",
			input:    "= :=",
			expected: []TokenType{TokenOperator, TokenOperator},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var types []TokenType
			for _, tok := range tokens {
				types = append(types, tok.Type)
			}

			if !reflect.DeepEqual(types, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, types)
			}
		})
	}
}

func TestParser_HexLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Hex with x' prefix",
			input:    "SELECT x'48656C6C6F'",
			expected: []string{"SELECT", "x'48656C6C6F'"},
		},
		{
			name:     "Hex with 0x prefix",
			input:    "SELECT 0xFF",
			expected: []string{"SELECT", "0xFF"},
		},
		{
			name:     "Uppercase X",
			input:    "SELECT X'ABCD'",
			expected: []string{"SELECT", "X'ABCD'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				// fmt.Println(token.Type.String(), token.Lexeme)
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, tok.Lexeme)
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestParser_BitValueLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Binary with b' prefix",
			input:    "SELECT b'101010'",
			expected: []string{"SELECT", "b'101010'"},
		},
		{
			name:     "Binary with 0b prefix",
			input:    "SELECT 0b101010",
			expected: []string{"SELECT", "0b101010"},
		},
		{
			name:     "Uppercase B",
			input:    "SELECT B'111000'",
			expected: []string{"SELECT", "B'111000'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, tok.Lexeme)
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestParser_SpecialLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "NULL literal",
			input:    "SELECT NULL",
			expected: []TokenType{TokenKeyword, TokenLiteral},
		},
		{
			name:     "Boolean literals",
			input:    "SELECT TRUE, FALSE",
			expected: []TokenType{TokenKeyword, TokenLiteral, TokenComma, TokenLiteral},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []*Token
			p := NewParser(func(token *Token) bool {
				tokens = append(tokens, token)
				return true
			})

			p.Parse(tt.input)

			var types []TokenType
			for _, tok := range tokens {
				types = append(types, tok.Type)
			}

			if !reflect.DeepEqual(types, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, types)
			}
		})
	}
}

func TestParser_Parentheses(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	p.Parse("SELECT COUNT(*) FROM (SELECT id FROM users)")

	var parenTypes []TokenType
	for _, tok := range tokens {
		if tok.Type == TokenOpenParen || tok.Type == TokenCloseParen {
			parenTypes = append(parenTypes, tok.Type)
		}
	}

	expected := []TokenType{TokenOpenParen, TokenCloseParen, TokenOpenParen, TokenCloseParen}
	if !reflect.DeepEqual(parenTypes, expected) {
		t.Errorf("Expected parentheses %v, got %v", expected, parenTypes)
	}
}

func TestParser_Commas(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	p.Parse("SELECT id, name, email FROM users")

	commaCount := 0
	for _, tok := range tokens {
		if tok.Type == TokenComma {
			commaCount++
		}
	}

	if commaCount != 2 {
		t.Errorf("Expected 2 commas, got %d", commaCount)
	}
}

func TestParser_ComplexQuery(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	query := `SELECT u.id, u.name, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.created_at >= '2023-01-01' 
			  GROUP BY u.id, u.name 
			  HAVING COUNT(o.id) > 5 
			  ORDER BY order_count DESC`

	p.Parse(query)

	// Count different token types
	counts := make(map[TokenType]int)
	for _, tok := range tokens {
		counts[tok.Type]++
	}

	// Basic sanity checks
	if counts[TokenKeyword] == 0 {
		t.Error("Expected keywords in complex query")
	}
	if counts[TokenOperator] == 0 {
		t.Error("Expected operators in complex query")
	}
	if counts[TokenLiteral] == 0 {
		t.Error("Expected literals in complex query")
	}
}

func TestParser_EmptyString(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	p.Parse("")

	if len(tokens) != 0 {
		t.Errorf("Expected no tokens for empty string, got %d", len(tokens))
	}
}

func TestParser_WhitespaceHandling(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	// Test with various whitespace
	p.Parse("   SELECT   *   FROM   table   ")

	expected := []string{"SELECT", "*", "FROM", "table"}
	var lexemes []string
	for _, tok := range tokens {
		lexemes = append(lexemes, tok.Lexeme)
	}

	if !reflect.DeepEqual(lexemes, expected) {
		t.Errorf("Expected %v, got %v", expected, lexemes)
	}
}

func TestParser_TokenPositions(t *testing.T) {
	var tokens []*Token
	p := NewParser(func(token *Token) bool {
		tokens = append(tokens, token)
		return true
	})

	p.Parse("SELECT id")

	if len(tokens) != 2 {
		t.Fatalf("Expected 2 tokens, got %d", len(tokens))
	}

	// Check that positions are reasonable
	for i, tok := range tokens {
		if tok.Pos.start < 0 || tok.Pos.end < tok.Pos.start {
			t.Errorf("Token %d has invalid position: start=%d, end=%d",
				i, tok.Pos.start, tok.Pos.end)
		}
	}
}

func TestParser_EarlyTermination(t *testing.T) {
	callCount := 0
	p := NewParser(func(token *Token) bool {
		callCount++
		return callCount < 3 // Stop after 2 tokens
	})

	p.Parse("SELECT * FROM users WHERE id = 1")

	if callCount != 3 {
		t.Errorf("Expected parser to stop after 3 calls, but got %d calls", callCount)
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TokenOpenParen, "TokenOpenParen"},
		{TokenCloseParen, "TokenCloseParen"},
		{TokenComma, "TokenComma"},
		{TokenKeyword, "TokenKeyword"},
		{TokenLiteral, "TokenLiteral"},
		{TokenOperator, "TokenOperator"},
		{TokenEOF, "unknown"}, // Based on the String() method implementation
	}

	for _, tt := range tests {
		if got := tt.tokenType.String(); got != tt.expected {
			t.Errorf("TokenType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestParser_Reset(t *testing.T) {
	p := NewParser(func(token *Token) bool { return true })
	p.Parse("SELECT * FROM users")

	// Check that curr position was advanced
	if p.curr == 0 {
		t.Error("Expected curr to be advanced after parsing")
	}

	p.Reset()

	// Check that Reset() works
	if p.curr != 0 {
		t.Errorf("Expected curr to be 0 after Reset(), got %d", p.curr)
	}
}

func BenchmarkParser_SimpleQuery(b *testing.B) {
	query := "SELECT * FROM users WHERE id = 1"

	for i := 0; i < b.N; i++ {
		p := NewParser(func(token *Token) bool { return true })
		p.Parse(query)
	}
}

func BenchmarkParser_ComplexQuery(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.created_at >= '2023-01-01' 
			  GROUP BY u.id, u.name 
			  HAVING COUNT(o.id) > 5 
			  ORDER BY order_count DESC`

	for i := 0; i < b.N; i++ {
		p := NewParser(func(token *Token) bool { return true })
		p.Parse(query)
	}
}
