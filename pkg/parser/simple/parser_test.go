package simple

import (
	"reflect"
	"strings"
	"testing"
)

func addIntoTokenSlice(toks []Token, parser *Parser) []Token {
	for {
		tok := parser.NextToken()
		if tok.Type == TokenEOF {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

func TestParser_BasicSQL(t *testing.T) {
	var tokens []Token
	p := NewParser()

	p.Parse([]byte("SELECT * FROM users WHERE id = 1"))
	tokens = addIntoTokenSlice(tokens, p)

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
		if string(p.GetLexeme(tok)) != expected[i].lexeme {
			t.Errorf("Token %d: expected lexeme '%s', got '%s'", i, expected[i].lexeme, string(p.GetLexeme(tok)))
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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(string(p.GetLexeme(tok))))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestParser_CrazyStringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty strings",
			input:    "SELECT '' AND \"\"",
			expected: []string{"SELECT", "''", "AND", "\"\""},
		},
		{
			name:     "Escaped quotes",
			input:    "SELECT '\\'escaped\\'' AND \"\\\"escaped\\\"\"",
			expected: []string{"SELECT", "'\\'escaped\\''", "AND", "\"\\\"escaped\\\"\""},
		},
		{
			name:     "Mixed quotes with spaces",
			input:    "SELECT 'hello \"world\"' AND \"hello 'world'\"",
			expected: []string{"SELECT", "'hello \"world\"'", "AND", "\"hello 'world'\""},
		},
		{
			name:     "Multiple escaped characters",
			input:    "SELECT '\\n\\t\\r\\b\\\\'",
			expected: []string{"SELECT", "'\\n\\t\\r\\b\\\\'"},
		},
		{
			name:     "Unicode characters",
			input:    "SELECT '你好世界' AND \"🎉🎈🎂\"",
			expected: []string{"SELECT", "'你好世界'", "AND", "\"🎉🎈🎂\""},
		},
		{
			name:     "Very long string",
			input:    "SELECT '" + strings.Repeat("a", 1000) + "'",
			expected: []string{"SELECT", "'" + strings.Repeat("a", 1000) + "'"},
		},
		{
			name:     "String with special SQL characters",
			input:    "SELECT 'SELECT * FROM table WHERE id = 1'",
			expected: []string{"SELECT", "'SELECT * FROM table WHERE id = 1'"},
		},
		{
			name:     "String with numbers and operators",
			input:    "SELECT '123 + 456 = 579' AND \"10 * 20 = 200\"",
			expected: []string{"SELECT", "'123 + 456 = 579'", "AND", "\"10 * 20 = 200\""},
		},
		{
			name:     "String with multiple escaped backslashes",
			input:    "SELECT '\\\\\\\\'",
			expected: []string{"SELECT", "'\\\\\\\\'"},
		},
		{
			name:     "String with mixed line endings",
			input:    "SELECT 'line1\\nline2\\r\\nline3'",
			expected: []string{"SELECT", "'line1\\nline2\\r\\nline3'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(p.GetLexeme(tok)))
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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(p.GetLexeme(tok)))
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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(p.GetLexeme(tok)))
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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(p.GetLexeme(tok)))
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
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

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
	var tokens []Token
	p := NewParser()

	p.Parse([]byte("SELECT COUNT(*) FROM (SELECT id FROM users)"))
	tokens = addIntoTokenSlice(tokens, p)

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
	var tokens []Token
	p := NewParser()

	p.Parse([]byte("SELECT id, name, email FROM users"))
	tokens = addIntoTokenSlice(tokens, p)

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
	var tokens []Token
	p := NewParser()

	query := `SELECT u.id, u.name, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.created_at >= '2023-01-01' 
			  GROUP BY u.id, u.name 
			  HAVING COUNT(o.id) > 5 
			  ORDER BY order_count DESC`

	p.Parse([]byte(query))
	tokens = addIntoTokenSlice(tokens, p)

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
	var tokens []Token
	p := NewParser()

	p.Parse([]byte(""))
	tokens = addIntoTokenSlice(tokens, p)

	if len(tokens) != 0 {
		t.Errorf("Expected no tokens for empty string, got %d", len(tokens))
	}
}

func TestParser_WhitespaceHandling(t *testing.T) {
	var tokens []Token
	p := NewParser()

	// Test with various whitespace
	p.Parse([]byte("   SELECT   *   FROM   table   "))
	tokens = addIntoTokenSlice(tokens, p)

	expected := []string{"SELECT", "*", "FROM", "table"}
	var lexemes []string
	for _, tok := range tokens {
		lexemes = append(lexemes, string(p.GetLexeme(tok)))
	}

	if !reflect.DeepEqual(lexemes, expected) {
		t.Errorf("Expected %v, got %v", expected, lexemes)
	}
}

func TestParser_TokenPositions(t *testing.T) {
	var tokens []Token
	p := NewParser()

	p.Parse([]byte("SELECT id"))
	tokens = addIntoTokenSlice(tokens, p)

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

// func TestParser_EarlyTermination(t *testing.T) {
// 	callCount := 0
// 	p := NewParser(func(_, token Token) bool {
// 		callCount++
// 		return callCount < 3 // Stop after 2 tokens
// 	})

// 	p.Parse([]byte("SELECT * FROM users WHERE id = 1"))

// 	if callCount != 3 {
// 		t.Errorf("Expected parser to stop after 3 calls, but got %d calls", callCount)
// 	}
// }

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
	p := NewParser()
	p.Parse([]byte("SELECT * FROM users"))
	addIntoTokenSlice([]Token{}, p)

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
	query := []byte("SELECT * FROM users WHERE id = 1")

	p := NewParser()
	for i := 0; i < b.N; i++ {
		p.Parse(query)
		for {
			tok := p.NextToken()
			if tok.Type == TokenEOF {
				break
			}
		}
	}
}

func BenchmarkParser_ComplexQuery(b *testing.B) {
	query := []byte(`SELECT u.id, u.name, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.created_at >= '2023-01-01' 
			  GROUP BY u.id, u.name 
			  HAVING COUNT(o.id) > 5 
			  ORDER BY order_count DESC`)

	p := NewParser()
	p.Reset()
	for i := 0; i < b.N; i++ {
		p.Parse(query)
		for {
			tok := p.NextToken()
			if tok.Type == TokenEOF {
				break
			}
		}
	}
}

func TestParser_CrazyKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Table names with special characters",
			input:    "SELECT * FROM `my-table` WHERE `my.column` = 1",
			expected: []string{"SELECT", "*", "FROM", "`my-table`", "WHERE", "`my.column`", "=", "1"},
		},
		{
			name:     "Table names with numbers",
			input:    "SELECT * FROM table_123 WHERE col_456 = 789",
			expected: []string{"SELECT", "*", "FROM", "table_123", "WHERE", "col_456", "=", "789"},
		},
		{
			name:     "Reserved keywords as identifiers",
			input:    "SELECT `select` `from` `where` FROM `table`",
			expected: []string{"SELECT", "`select`", "`from`", "`where`", "FROM", "`table`"},
		},
		{
			name:     "Mixed case keywords and identifiers",
			input:    "SeLeCt * FrOm TaBlE WhErE iD = 1",
			expected: []string{"SeLeCt", "*", "FrOm", "TaBlE", "WhErE", "iD", "=", "1"},
		},
		{
			name:     "Table aliases",
			input:    "SELECT t1.id, t2.name FROM table1 AS t1 JOIN table2 t2",
			expected: []string{"SELECT", "t1.id", ",", "t2.name", "FROM", "table1", "AS", "t1", "JOIN", "table2", "t2"},
		},
		{
			name:     "Database qualified names",
			input:    "SELECT db.table.column FROM db.table",
			expected: []string{"SELECT", "db.table.column", "FROM", "db.table"},
		},
		{
			name:     "Keywords with underscores",
			input:    "SELECT * FROM my_table WHERE is_deleted = FALSE",
			expected: []string{"SELECT", "*", "FROM", "my_table", "WHERE", "is_deleted", "=", "FALSE"},
		},
		{
			name:     "Keywords with dots",
			input:    "SELECT * FROM `my.schema.table` WHERE `col.name` = 'value'",
			expected: []string{"SELECT", "*", "FROM", "`my.schema.table`", "WHERE", "`col.name`", "=", "'value'"},
		},
		{
			name:     "Keywords with special characters",
			input:    "SELECT * FROM `my@table` WHERE `col#name` = 'value'",
			expected: []string{"SELECT", "*", "FROM", "`my@table`", "WHERE", "`col#name`", "=", "'value'"},
		},
		{
			name:     "Keywords with spaces",
			input:    "SELECT * FROM `my table` WHERE `my column` = 'value'",
			expected: []string{"SELECT", "*", "FROM", "`my table`", "WHERE", "`my column`", "=", "'value'"},
		},
		{
			name:     "Keywords with unicode",
			input:    "SELECT * FROM `表名` WHERE `列名` = '值'",
			expected: []string{"SELECT", "*", "FROM", "`表名`", "WHERE", "`列名`", "=", "'值'"},
		},
		{
			name:     "Keywords with emojis",
			input:    "SELECT * FROM `my_😊_table` WHERE `col_🎉_name` = 'value'",
			expected: []string{"SELECT", "*", "FROM", "`my_😊_table`", "WHERE", "`col_🎉_name`", "=", "'value'"},
		},
		{
			name:     "Keywords with multiple dots",
			input:    "SELECT * FROM `db.schema.table.column`",
			expected: []string{"SELECT", "*", "FROM", "`db.schema.table.column`"},
		},
		{
			name:     "Keywords with backticks in names",
			input:    "SELECT * FROM `table``name` WHERE `col``name` = 'value'",
			expected: []string{"SELECT", "*", "FROM", "`table``name`", "WHERE", "`col``name`", "=", "'value'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			p := NewParser()

			p.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, p)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(p.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}
