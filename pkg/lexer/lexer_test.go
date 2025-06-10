package lexer

import (
	"bytes"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func init() {
	runtime.GOMAXPROCS(1)
}

func addIntoTokenSlice(toks []Token, lexer *Lexer) []Token {
	for {
		tok := lexer.NextToken()
		if tok.Type == TokenEOF {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

func TestLexer_BasicSQL(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	l.Parse([]byte("SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)"))
	tokens = addIntoTokenSlice(tokens, l)

	expected := []struct {
		tokenType TokenType
		tokenAttr Attr
		lexeme    string
	}{
		{TokenKeyword, TokenAttrBuiltIn, "SELECT"},
		{TokenStar, 0, "*"},
		{TokenKeyword, TokenAttrBuiltIn, "FROM"},
		{TokenKeyword, 0, "users"},
		{TokenKeyword, TokenAttrBuiltIn, "WHERE"},
		{TokenKeyword, 0, "id"},
		{TokenKeyword, 0, "IN"},
		{TokenOpenParen, 0, "("},
		{TokenKeyword, TokenAttrBuiltIn, "SELECT"},
		{TokenKeyword, 0, "user_id"},
		{TokenKeyword, TokenAttrBuiltIn, "FROM"},
		{TokenKeyword, 0, "orders"},
		{TokenKeyword, TokenAttrBuiltIn, "WHERE"},
		{TokenKeyword, 0, "total"},
		{TokenOperator, 0, ">"},
		{TokenLiteral, 0, "100"},
		{TokenCloseParen, 0, ")"},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i].tokenType {
			t.Errorf("Token %d: expected type %s, got %s, tokens: %v", i, expected[i].tokenType.String(), tok.Type.String(), tokens)
		}
		if expected[i].tokenAttr != 0 && (tok.Attr&expected[i].tokenAttr) == 0 {
			t.Errorf("Token %d: expected attr %d, got %d", i, expected[i].tokenAttr, tok.Attr)
		}
		if string(l.GetLexeme(tok)) != expected[i].lexeme {
			t.Errorf("Token %d: expected lexeme '%s', got '%s'", i, expected[i].lexeme, string(l.GetLexeme(tok)))
		}
	}
}

func TestLexer_StringLiterals(t *testing.T) {
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

	l := NewLexer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.Reset()
			l.Parse([]byte(tt.input))

			var tokens []Token

			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(string(l.GetLexeme(tok))))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestLexer_CrazyStringLiterals(t *testing.T) {
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestLexer_NumberLiterals(t *testing.T) {
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestLexer_Fuzzy(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLexemes []string
		expectedTokens  []TokenType
	}{
		{
			name:            "number literals",
			input:           "SELECT 123)12e-4+12-31233()",
			expectedLexemes: []string{"SELECT", "123", ")", "12e-4", "+12-31233", "(", ")"},
			expectedTokens: []TokenType{
				TokenKeyword, TokenLiteral, TokenCloseParen, TokenLiteral, TokenLiteral, TokenOpenParen, TokenCloseParen,
			},
		},
		{
			name:            "Binary data literals",
			input:           "SELECT * FROM files WHERE content = 'binary\x00\x01\x02'",
			expectedLexemes: []string{"SELECT", "*", "FROM", "files", "WHERE", "content", "=", "'binary\x00\x01\x02'"},
			expectedTokens: []TokenType{
				TokenKeyword, TokenStar, TokenKeyword, TokenKeyword, TokenKeyword, TokenKeyword, TokenOperator, TokenLiteral,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			if len(tt.expectedTokens) != len(tokens) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expectedTokens), len(tokens))
			}

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expectedTokens[i] {
					t.Errorf("Token %d: expected type %s, got %s", i, tt.expectedTokens[i].String(), tok.Type.String())
				}
				if !bytes.Equal(l.GetLexeme(tok), []byte(tt.expectedLexemes[i])) {
					t.Errorf("Token %d: expected lexeme '%s', got '%s'", i, tt.expectedLexemes[i], string(l.GetLexeme(tok)))
				}
			}
		})
	}
}

func TestLexer_Operators(t *testing.T) {
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
			expected: []TokenType{TokenOperator, TokenOperator, TokenStar, TokenOperator, TokenOperator},
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_HexLiterals(t *testing.T) {
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestLexer_BitValueLiterals(t *testing.T) {
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func TestLexer_SpecialLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "NULL literal",
			input:    "SELECT NULL",
			expected: []TokenType{TokenKeyword, TokenKeyword},
		},
		{
			name:     "Boolean literals",
			input:    "SELECT TRUE, FALSE",
			expected: []TokenType{TokenKeyword, TokenKeyword, TokenComma, TokenKeyword},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_Parentheses(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	l.Parse([]byte("SELECT COUNT(*) FROM (SELECT id FROM users)"))
	tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_Commas(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	l.Parse([]byte("SELECT id, name, email FROM users"))
	tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_ComplexQuery(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	query := `SELECT u.id, u.name, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.created_at >= '2023-01-01' 
			  GROUP BY u.id, u.name 
			  HAVING COUNT(o.id) > 5 
			  ORDER BY order_count DESC`

	l.Parse([]byte(query))
	tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_EmptyString(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	l.Parse([]byte(""))
	tokens = addIntoTokenSlice(tokens, l)

	if len(tokens) != 0 {
		t.Errorf("Expected no tokens for empty string, got %d", len(tokens))
	}
}

func TestLexer_WhitespaceHandling(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	// Test with various whitespace
	l.Parse([]byte("   SELECT   *   FROM   table   "))
	tokens = addIntoTokenSlice(tokens, l)

	expected := []string{"SELECT", "*", "FROM", "table"}
	var lexemes []string
	for _, tok := range tokens {
		lexemes = append(lexemes, string(l.GetLexeme(tok)))
	}

	if !reflect.DeepEqual(lexemes, expected) {
		t.Errorf("Expected %v, got %v", expected, lexemes)
	}
}

func TestLexer_TokenPositions(t *testing.T) {
	var tokens []Token
	l := NewLexer()

	l.Parse([]byte("SELECT id"))
	tokens = addIntoTokenSlice(tokens, l)

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

func TestLexer_Reset(t *testing.T) {
	l := NewLexer()
	l.Parse([]byte("SELECT * FROM users"))
	addIntoTokenSlice([]Token{}, l)

	// Check that curr position was advanced
	if l.curr == 0 {
		t.Error("Expected curr to be advanced after parsing")
	}

	l.Reset()

	// Check that Reset() works
	if l.curr != 0 {
		t.Errorf("Expected curr to be 0 after Reset(), got %d", l.curr)
	}
}

func TestLexer_CrazyKeywords(t *testing.T) {
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
			expected: []string{"SELECT", "t1", ".", "id", ",", "t2", ".", "name", "FROM", "table1", "AS", "t1", "JOIN", "table2", "t2"},
		},
		{
			name:     "Database qualified names",
			input:    "SELECT db.table.column FROM db.table",
			expected: []string{"SELECT", "db", ".", "table", ".", "column", "FROM", "db", ".", "table"},
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
			l := NewLexer()

			l.Parse([]byte(tt.input))
			tokens = addIntoTokenSlice(tokens, l)

			var lexemes []string
			for _, tok := range tokens {
				lexemes = append(lexemes, string(l.GetLexeme(tok)))
			}

			if !reflect.DeepEqual(lexemes, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, lexemes)
			}
		})
	}
}

func BenchmarkLexer_SimpleQuery(b *testing.B) {
	query := []byte("SELECT * FROM users WHERE id = 1")

	l := NewLexer()
	l.Parse(query)
	for i := 0; i < b.N; i++ {
		l.Reset()
		for {
			tok := l.NextToken()
			if tok.Type == TokenEOF {
				break
			}
		}
	}
}

func BenchmarkLexer_ComplexQuery(b *testing.B) {
	query := []byte(`
		WITH recent_orders AS (
			SELECT user_id, MAX(created_at) as last_order
			FROM orders
			WHERE created_at >= NOW() - INTERVAL 6 MONTH
			GROUP BY user_id
		),
		active_users AS (
			SELECT u.id, u.name, COUNT(o.id) as total_orders
			FROM users u
			INNER JOIN orders o ON u.id = o.user_id
			WHERE u.status = 'active' AND o.status IN ('shipped', 'delivered')
			GROUP BY u.id, u.name
			HAVING COUNT(o.id) > 10
		)
		SELECT au.id, au.name, ro.last_order, COUNT(p.id) as total_payments,
		       SUM(p.amount) as total_spent, AVG(p.amount) as avg_payment,
		       MAX(p.amount) as max_payment
		FROM active_users au
		LEFT JOIN recent_orders ro ON au.id = ro.user_id
		LEFT JOIN payments p ON au.id = p.user_id
		WHERE p.status = 'completed'
		GROUP BY au.id, au.name, ro.last_order
		HAVING total_spent > 1000
		ORDER BY total_spent DESC

		UNION ALL

		SELECT u.id, u.name, NULL, 0, 0, 0, 0
		FROM users u
		WHERE u.status = 'inactive'
	`)

	l := NewLexer()
	l.Parse(query)
	for i := 0; i < b.N; i++ {
		l.Reset()
		for {
			tok := l.NextToken()
			if tok.Type == TokenEOF {
				break
			}
		}
	}
}
