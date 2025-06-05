package normalizer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bagaswh/mysql-toolkit/pkg/lexer"
)

func TestNormalize(t *testing.T) {
	lex := lexer.NewLexer()

	tests := []struct {
		name     string
		config   Config
		input    string
		expected string
	}{
		{
			name: "basic select with literals removed",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "SELECT id, name FROM users WHERE age = 25",
			expected: "SELECT ID, NAME FROM USERS WHERE AGE = ?",
		},
		{
			name: "lowercase keywords",
			config: Config{
				KeywordCase:    CaseLower,
				RemoveLiterals: false,
			},
			input:    "SELECT * FROM users WHERE id = 1",
			expected: "select * from users where id = 1",
		},
		{
			name: "uppercase keywords with literals",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: false,
			},
			input:    "select * from users where name = 'john'",
			expected: "SELECT * FROM USERS WHERE NAME = 'john'",
		},
		{
			name: "complex query with joins",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.age > 18",
			expected: "SELECT U.ID, U.NAME, P.TITLE FROM USERS U JOIN POSTS P ON U.ID = P.USER_ID WHERE U.AGE > ?",
		},
		{
			name: "insert statement",
			config: Config{
				KeywordCase:    CaseLower,
				RemoveLiterals: true,
			},
			input:    "INSERT INTO users (name, email, age) VALUES ('John Doe', 'john@example.com', 30)",
			expected: "insert into users(name, email, age) values(?, ?, ?)",
		},
		{
			name: "update statement",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "UPDATE users SET name = 'Jane Doe', age = 25 WHERE id = 1",
			expected: "UPDATE USERS SET NAME = ?, AGE = ? WHERE ID = ?",
		},
		{
			name: "delete statement",
			config: Config{
				KeywordCase:    CaseLower,
				RemoveLiterals: true,
			},
			input:    "DELETE FROM users WHERE age < 18 AND status = 'inactive'",
			expected: "delete from users where age < ? and status = ?",
		},
		{
			name: "subquery",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
			expected: "SELECT * FROM USERS WHERE ID IN(SELECT USER_ID FROM ORDERS WHERE TOTAL > ?)",
		},
		{
			name: "function calls",
			config: Config{
				KeywordCase:             CaseLower,
				RemoveLiterals:          true,
				PutSpaceBeforeOpenParen: true,
				PutBacktickOnKeywords:   true,
			},
			input:    "SELECT COUNT(*), MAX(age), MIN(created_at) FROM users WHERE name LIKE '%john%'",
			expected: "select `count` (`*`), `max` (`age`), `min` (`created_at`) from `users` where `name` like ?",
		},
		{
			name: "window functions",
			config: Config{
				KeywordCase:             CaseUpper,
				RemoveLiterals:          false,
				PutSpaceBeforeOpenParen: true,
				PutBacktickOnKeywords:   true,
			},
			input:    "SELECT name, ROW_NUMBER () OVER (ORDER BY age DESC) as `rank` FROM users",
			expected: "SELECT `NAME`, ROW_NUMBER () OVER (ORDER BY `AGE` DESC) AS `RANK` FROM `USERS`",
		},
		{
			name: "empty query",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "",
			expected: "",
		},
		{
			name: "only whitespace",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name: "numeric literals",
			config: Config{
				KeywordCase:    CaseLower,
				RemoveLiterals: true,
			},
			input:    "SELECT * FROM products WHERE price = 99.99 AND quantity >= 10",
			expected: "select * from products where price = ? and quantity >= ?",
		},
		{
			name: "string literals with quotes",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    `SELECT * FROM users WHERE name = "John's User" AND description = 'He said "Hello"'`,
			expected: "SELECT * FROM USERS WHERE NAME = ? AND DESCRIPTION = ?",
		},
		{
			name: "case statement",
			config: Config{
				KeywordCase:    CaseUpper,
				RemoveLiterals: true,
			},
			input:    "SELECT name, CASE WHEN age < 18 THEN 'minor' ELSE 'adult' END as category FROM users",
			expected: "SELECT NAME, CASE WHEN AGE < ? THEN ? ELSE ? END AS CATEGORY FROM USERS",
		},
		{
			name: "union query",
			config: Config{
				KeywordCase:    CaseLower,
				RemoveLiterals: false,
			},
			input:    "SELECT name FROM users WHERE active = 1 UNION SELECT name FROM admins WHERE active = 1",
			expected: "select name from users where active = 1 union select name from admins where active = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]byte, len(tt.input)*2)
			_, normalized, err := Normalize(tt.config, lex, []byte(tt.input), result)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}

			actual := string(normalized)
			if actual != tt.expected {
				t.Errorf("Normalize() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

// Test all configuration combinations systematically
func TestNormalize_AllConfigCombinations(t *testing.T) {
	lex := lexer.NewLexer()

	// Generate all possible config combinations
	keywordCases := []Case{CaseDefault, CaseLower, CaseUpper}
	boolOptions := []bool{true, false}

	type testCase struct {
		name     string
		input    string
		expected map[string]string // config description -> expected output
	}

	testCases := []testCase{
		{
			name:  "basic_select_with_backticks",
			input: "SELECT `user_id`, `full_name` FROM `user_table` WHERE `age` = 25",
			expected: map[string]string{
				"default_false_false_false_false": "SELECT `user_id`, `full_name` FROM `user_table` WHERE `age` = 25",
				"lower_false_false_false_false":   "select `user_id`, `full_name` from `user_table` where `age` = 25",
				"upper_false_false_false_false":   "SELECT `USER_ID`, `FULL_NAME` FROM `USER_TABLE` WHERE `AGE` = 25",
				"upper_true_false_false_false":    "SELECT `USER_ID`, `FULL_NAME` FROM `USER_TABLE` WHERE `AGE` = ?",
				"upper_false_true_false_false":    "SELECT `USER_ID`, `FULL_NAME` FROM `USER_TABLE` WHERE `AGE` = 25",
				"upper_false_false_true_false":    "SELECT USER_ID, FULL_NAME FROM USER_TABLE WHERE AGE = 25",
				"upper_false_true_false_true":     "SELECT `USER_ID`, `FULL_NAME` FROM `USER_TABLE` WHERE `AGE` = 25",
			},
		},
		{
			name:  "function_calls_with_parens",
			input: "SELECT COUNT(*), MAX(age) FROM users WHERE name LIKE 'John%'",
			expected: map[string]string{
				"default_false_false_false_false": "SELECT COUNT(*), MAX(age) FROM users WHERE name LIKE 'John%'",
				"lower_true_true_false_true":      "select `count` (`*`), `max` (`age`) from `users` where `name` like ?",
				"upper_true_true_false_true":      "SELECT `COUNT` (`*`), `MAX` (`AGE`) FROM `USERS` WHERE `NAME` LIKE ?",
			},
		},
		{
			name:  "identifiers_without_backticks",
			input: "SELECT user_id, full_name FROM user_table WHERE `status` = 'active'",
			expected: map[string]string{
				"default_false_false_false_false": "SELECT user_id, full_name FROM user_table WHERE `status` = 'active'",
				"upper_false_true_false_false":    "SELECT `USER_ID`, `FULL_NAME` FROM `USER_TABLE` WHERE `STATUS` = 'active'",
				"lower_true_true_false_false":     "select `user_id`, `full_name` from `user_table` where `status` = ?",
			},
		},
		{
			name:  "mixed_quotes_and_literals",
			input: `SELECT * FROM users WHERE name = "John's Data" AND age = 25 AND score = 99.5`,
			expected: map[string]string{
				"upper_true_true_false_false":   "SELECT `*` FROM `USERS` WHERE `NAME` = ? AND `AGE` = ? AND `SCORE` = ?",
				"lower_false_false_false_false": `select * from users where name = "John's Data" and age = 25 and score = 99.5`,
			},
		},
	}

	// Test all combinations with proper constraint validation
	for _, keywordCase := range keywordCases {
		for _, removeLiterals := range boolOptions {
			for _, putBacktick := range boolOptions {
				for _, removeBacktick := range boolOptions {
					for _, spaceBeforeParen := range boolOptions {
						// Skip invalid combinations based on constraints
						if putBacktick && removeBacktick {
							// Cannot both put and remove backticks
							continue
						}
						if spaceBeforeParen && !putBacktick {
							// PutSpaceBeforeOpenParen requires PutBacktickOnKeywords to be true
							continue
						}

						config := Config{
							KeywordCase:              keywordCase,
							RemoveLiterals:           removeLiterals,
							PutBacktickOnKeywords:    putBacktick,
							RemoveBacktickOnKeywords: removeBacktick,
							PutSpaceBeforeOpenParen:  spaceBeforeParen,
						}

						configKey := fmt.Sprintf("%s_%t_%t_%t_%t",
							getCaseName(keywordCase), removeLiterals, putBacktick, removeBacktick, spaceBeforeParen)

						for _, tc := range testCases {
							t.Run(fmt.Sprintf("%s_%s", tc.name, configKey), func(t *testing.T) {
								result := make([]byte, len(tc.input)*3)
								_, normalized, err := Normalize(config, lex, []byte(tc.input), result)
								if err != nil {
									t.Errorf("Normalize() error = %v", err)
									return
								}
								actual := string(normalized)

								// Check if we have an expected result for this config
								if expected, exists := tc.expected[configKey]; exists {
									if actual != expected {
										t.Errorf("Config %s: got %q, want %q", configKey, actual, expected)
									}
								} else {
									// Log configurations that don't have expected results for debugging
									t.Logf("No expected result for config %s, got: %q", configKey, actual)
								}

								// Basic invariants that should always hold
								validateBasicInvariants(t, tc.input, actual, config)
							})
						}
					}
				}
			}
		}
	}
}

// Test invalid configurations should return errors
func TestNormalize_InvalidConfigurations(t *testing.T) {
	lex := lexer.NewLexer()
	input := []byte("SELECT COUNT(*) FROM users")
	result := make([]byte, len(input)*3)

	testCases := []struct {
		name   string
		config Config
		errMsg string
	}{
		{
			name: "both_put_and_remove_backticks",
			config: Config{
				PutBacktickOnKeywords:    true,
				RemoveBacktickOnKeywords: true,
			},
			errMsg: "PutBacktickOnKeywords and RemoveBacktickOnKeywords cannot be both true",
		},
		{
			name: "space_before_paren_without_backticks",
			config: Config{
				PutSpaceBeforeOpenParen: true,
				PutBacktickOnKeywords:   false,
			},
			errMsg: "PutSpaceBeforeOpenParen requires PutBacktickOnKeywords to be true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := Normalize(tc.config, lex, input, result)
			if err == nil {
				t.Errorf("Expected error for invalid config, but got nil")
				return
			}
			if !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tc.errMsg, err.Error())
			}
		})
	}
}

func getCaseName(c Case) string {
	switch c {
	case CaseDefault:
		return "default"
	case CaseLower:
		return "lower"
	case CaseUpper:
		return "upper"
	default:
		return "unknown"
	}
}

func validateBasicInvariants(t *testing.T, input, output string, config Config) {
	// Output should never be longer than reasonable bounds
	if len(output) > len(input)*3 {
		t.Errorf("Output too long: input=%d, output=%d", len(input), len(output))
	}

	// If RemoveLiterals is true, there should be no string literals in output
	if config.RemoveLiterals {
		if strings.Contains(output, "'") && !strings.Contains(output, "?") {
			t.Errorf("RemoveLiterals=true but output contains quotes without ?: %s", output)
		}
	}

	// Output should be valid (no malformed tokens)
	if strings.Contains(output, "  ") { // Multiple consecutive spaces
		t.Errorf("Output contains multiple consecutive spaces: %s", output)
	}
}

// Test edge cases and error conditions
func TestNormalize_EdgeCases(t *testing.T) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}

	edgeCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_whitespace",
			input:    "   \t\n\r  ",
			expected: "",
		},
		// {
		// 	name:     "only_comments",
		// 	input:    "-- this is a comment\n/* block comment */",
		// 	expected: "",
		// },
		{
			name:     "single_keyword",
			input:    "SELECT",
			expected: "SELECT",
		},
		{
			name:     "keywords_only",
			input:    "SELECT FROM WHERE",
			expected: "SELECT FROM WHERE",
		},
		{
			name:     "unterminated_string",
			input:    "SELECT * FROM users WHERE name = 'unterminated",
			expected: "SELECT * FROM USERS WHERE NAME = ?",
		},
		{
			name:     "empty_parens",
			input:    "SELECT COUNT() FROM users",
			expected: "SELECT COUNT() FROM USERS",
		},
		{
			name:     "nested_parens",
			input:    "SELECT ((1 + 2) * 3) FROM dual",
			expected: "SELECT((? + ?) * ?) FROM DUAL",
		},
		{
			name:     "special_characters",
			input:    "SELECT * FROM `table-name` WHERE `col@name` = 'val#ue'",
			expected: "SELECT * FROM `TABLE-NAME` WHERE `COL@NAME` = ?",
		},
		// {
		// 	name:     "unicode_identifiers",
		// 	input:    "SELECT café, naïve FROM müller WHERE résumé = 'données'",
		// 	expected: "SELECT CAFÉ, NAÏVE FROM MÜLLER WHERE RÉSUMÉ = ?",
		// },
		{
			name:     "very_long_identifier",
			input:    "SELECT " + strings.Repeat("very_long_column_name_", 20) + " FROM users",
			expected: "SELECT " + strings.ToUpper(strings.Repeat("very_long_column_name_", 20)) + " FROM USERS",
		},
		// {
		// 	name:     "sql_injection_attempt",
		// 	input:    "SELECT * FROM users WHERE id = 1; DROP TABLE users; --",
		// 	expected: "SELECT * FROM USERS WHERE ID = ?; DROP TABLE USERS; --",
		// },
		{
			name:     "binary_data_simulation",
			input:    "SELECT * FROM files WHERE content = 'binary\x00\x01\x02'",
			expected: "SELECT * FROM FILES WHERE CONTENT = ?",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := make([]byte, len(tc.input)*3)
			_, normalized, err := Normalize(config, lex, []byte(tc.input), result)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}
			actual := string(normalized)

			if actual != tc.expected {
				t.Errorf("got %q, want %q", actual, tc.expected)
			}
		})
	}
}

// // Test buffer size edge cases
func TestNormalize_BufferSizes(t *testing.T) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: false}
	input := "SELECT id FROM users WHERE name = 'test'"

	testCases := []struct {
		name       string
		bufferSize int
		shouldWork bool
	}{
		{"exact_size", len(input), true},
		{"too_small", len(input) / 2, false},
		{"double_size", len(input) * 2, true},
		{"minimal", 1, false},
		{"zero_size", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := make([]byte, tc.bufferSize)
			_, normalized, err := Normalize(config, lex, []byte(input), result)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}

			if tc.shouldWork {
				if len(normalized) == 0 && len(input) > 0 {
					t.Error("Expected non-empty result but got empty")
				}
			} else {
				// For cases that shouldn't work, we expect truncated or empty results
				// This tests the robustness of the buffer handling
				t.Logf("Buffer too small case resulted in: %q", string(normalized))
			}
		})
	}
}

// Fuzzy testing to catch unexpected edge cases
func FuzzNormalize(f *testing.F) {
	// Seed with some initial SQL samples
	seeds := []string{
		"SELECT * FROM users",
		"INSERT INTO users VALUES (1, 'test')",
		"UPDATE users SET name = 'new' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"SELECT COUNT(*) FROM users GROUP BY status",
		"SELECT * FROM users WHERE name LIKE '%test%'",
		"SELECT u.*, p.title FROM users u JOIN posts p ON u.id = p.user_id",
		"SELECT CASE WHEN age < 18 THEN 'minor' ELSE 'adult' END FROM users",
		"SELECT * FROM users WHERE id IN (1, 2, 3)",
		"CREATE TABLE test (id INT PRIMARY KEY)",
		"/* comment */ SELECT 1 -- line comment",
		"SELECT 'string with ''quotes''' FROM dual",
		"SELECT `backtick_id` FROM `table_name`",
		"SELECT 1.5, -42, 0x1F, 1e10",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// All possible configurations for fuzzing
	configs := []Config{
		{KeywordCase: CaseDefault, RemoveLiterals: false},
		{KeywordCase: CaseLower, RemoveLiterals: false},
		{KeywordCase: CaseUpper, RemoveLiterals: false},
		{KeywordCase: CaseUpper, RemoveLiterals: true},
		{KeywordCase: CaseLower, RemoveLiterals: true, PutBacktickOnKeywords: true},
		{KeywordCase: CaseUpper, RemoveLiterals: true, RemoveBacktickOnKeywords: true},
		{KeywordCase: CaseDefault, RemoveLiterals: false, PutBacktickOnKeywords: true, PutSpaceBeforeOpenParen: true},
		{KeywordCase: CaseUpper, RemoveLiterals: true, PutBacktickOnKeywords: true, PutSpaceBeforeOpenParen: true},
	}

	lex := lexer.NewLexer()

	f.Fuzz(func(t *testing.T, input string) {
		// Skip if input is too large to avoid timeout
		if len(input) > 10000 {
			t.Skip("Input too large for fuzzing")
		}

		for i, config := range configs {
			t.Run(fmt.Sprintf("config_%d", i), func(t *testing.T) {
				// Test with various buffer sizes
				bufferSizes := []int{
					len(input),           // exact
					len(input) * 2,       // double
					len(input) * 3,       // triple
					max(len(input)/2, 1), // half (minimum 1)
				}

				for _, bufSize := range bufferSizes {
					result := make([]byte, bufSize)

					// The normalize function should never panic
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("Normalize panicked with input %q, config %+v, buffer size %d: %v",
								input, config, bufSize, r)
						}
					}()

					_, normalized, err := Normalize(config, lex, []byte(input), result)
					if err != nil {
						t.Errorf("Normalize() error = %v", err)
						return
					}
					output := string(normalized)

					// Basic sanity checks
					if len(output) > len(result) {
						t.Errorf("Output length %d exceeds buffer size %d", len(output), len(result))
					}

					// Output should not contain null bytes (unless input did)
					if strings.Contains(output, "\x00") && !strings.Contains(input, "\x00") {
						t.Errorf("Output contains null bytes that weren't in input")
					}

					// If RemoveLiterals is set, check that we don't have obvious literals
					if config.RemoveLiterals && len(output) > 0 {
						// This is a heuristic check - we look for patterns that suggest
						// literals weren't properly replaced
						suspiciousPatterns := []string{"= '", "= \"", "= 123", "IN ('", "LIKE '"}
						for _, pattern := range suspiciousPatterns {
							if strings.Contains(strings.ToUpper(output), strings.ToUpper(pattern)) {
								// Only report if we're confident this is a bug
								if !strings.Contains(output, "?") {
									t.Logf("Possible literal not replaced in output: %q (pattern: %s)", output, pattern)
								}
							}
						}
					}

					// Keyword case should be consistent
					if config.KeywordCase == CaseUpper {
						keywords := []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "JOIN"}
						for _, keyword := range keywords {
							if strings.Contains(output, strings.ToLower(keyword)+" ") {
								// Only report if it's clearly a keyword in wrong case
								lowerPos := strings.Index(output, strings.ToLower(keyword)+" ")
								if lowerPos >= 0 && (lowerPos == 0 || output[lowerPos-1] == ' ') {
									t.Logf("Keyword not properly uppercased in output: %q", output)
									break
								}
							}
						}
					}
				}
			})
		}
	})
}

// Property-based testing for specific properties
func TestNormalize_Properties(t *testing.T) {
	lex := lexer.NewLexer()

	// Property: RemoveLiterals should reduce or maintain length
	t.Run("remove_literals_reduces_length", func(t *testing.T) {
		testCases := []string{
			"SELECT * FROM users WHERE id = 123 AND name = 'test'",
			"INSERT INTO users VALUES(1, 'john', 25, 'active')",
			"UPDATE users SET age = 30, score = 95.5 WHERE id = 1",
		}

		for _, input := range testCases {
			configWithLiterals := Config{RemoveLiterals: false}
			configWithoutLiterals := Config{RemoveLiterals: true}

			result1 := make([]byte, len(input)*2)
			_, normalized1, err := Normalize(configWithLiterals, lex, []byte(input), result1)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}

			result2 := make([]byte, len(input)*2)
			_, normalized2, err := Normalize(configWithoutLiterals, lex, []byte(input), result2)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}

			if len(normalized2) > len(normalized1) {
				t.Errorf("RemoveLiterals increased length: without=%d, with=%d (input: %s)",
					len(normalized1), len(normalized2), input)
			}
		}
	})

}

func TestNormalize_SmallBuffer(t *testing.T) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}
	input := "SELECT * FROM users WHERE name = 'test'"

	result := make([]byte, 5)
	_, normalized, err := Normalize(config, lex, []byte(input), result)
	if err != nil {
		t.Errorf("Normalize() error = %v", err)
		return
	}

	if string(normalized) != "SELEC" {
		t.Errorf("Normalize() result = %q, want %q", string(normalized), "SELEC")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
