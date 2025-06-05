package normalizer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/bagaswh/mysql-toolkit/pkg/lexer"
)

// Benchmark configurations
var benchConfigs = []struct {
	name   string
	config Config
}{
	{
		name: "DefaultCase",
		config: Config{
			KeywordCase:    CaseDefault,
			RemoveLiterals: false,
		},
	},
	{
		name: "UpperCase_RemoveLiterals",
		config: Config{
			KeywordCase:    CaseUpper,
			RemoveLiterals: true,
		},
	},
	{
		name: "LowerCase_RemoveLiterals",
		config: Config{
			KeywordCase:    CaseLower,
			RemoveLiterals: true,
		},
	},
	{
		name: "FullFeatures",
		config: Config{
			KeywordCase:             CaseUpper,
			RemoveLiterals:          true,
			PutBacktickOnKeywords:   true,
			PutSpaceBeforeOpenParen: true,
		},
	},
	{
		name: "RemoveBackticks",
		config: Config{
			KeywordCase:              CaseLower,
			RemoveLiterals:           true,
			RemoveBacktickOnKeywords: true,
		},
	},
}

// SQL templates for generating test queries
var sqlTemplates = []string{
	// Simple SELECT
	"SELECT %s FROM %s WHERE %s = %s",

	// Complex JOIN
	"SELECT %s.id, %s.name, %s.title FROM %s %s JOIN %s %s ON %s.id = %s.user_id WHERE %s.age > %s AND %s.status = %s",

	// Subquery
	"SELECT * FROM %s WHERE id IN (SELECT user_id FROM %s WHERE %s > %s)",

	// UNION
	"SELECT %s FROM %s WHERE %s = %s UNION ALL SELECT %s FROM %s WHERE %s = %s",

	// Window function
	"SELECT %s, ROW_NUMBER() OVER (PARTITION BY %s ORDER BY %s DESC) as rank FROM %s WHERE %s IS NOT NULL",

	// Complex CASE
	"SELECT %s, CASE WHEN %s < %s THEN %s WHEN %s > %s THEN %s ELSE %s END as category FROM %s",

	// INSERT with multiple values
	"INSERT INTO %s (%s, %s, %s) VALUES (%s, %s, %s), (%s, %s, %s), (%s, %s, %s)",

	// UPDATE with subquery
	"UPDATE %s SET %s = (SELECT MAX(%s) FROM %s WHERE %s = %s.%s) WHERE %s IN (%s, %s, %s)",

	// DELETE with JOIN
	"DELETE %s FROM %s %s INNER JOIN %s %s ON %s.id = %s.user_id WHERE %s.%s = %s",

	// Complex aggregation
	"SELECT %s, COUNT(*), AVG(%s), SUM(%s), MIN(%s), MAX(%s) FROM %s GROUP BY %s HAVING COUNT(*) > %s ORDER BY %s LIMIT %s",
}

// Data generators
var tableNames = []string{"users", "products", "orders", "customers", "inventory", "sales", "accounts", "profiles", "settings", "logs"}
var columnNames = []string{"id", "name", "email", "age", "price", "quantity", "status", "created_at", "updated_at", "title", "description", "category"}
var stringLiterals = []string{"'active'", "'inactive'", "'pending'", "'John Doe'", "'admin'", "'user'", "'premium'", "'basic'", "'completed'", "'cancelled'"}
var numericLiterals = []string{"1", "10", "100", "1000", "18", "25", "99.99", "0", "5", "50"}

func generateRandomIdentifier() string {
	return columnNames[rand.Intn(len(columnNames))]
}

func generateRandomTable() string {
	return tableNames[rand.Intn(len(tableNames))]
}

func generateRandomLiteral() string {
	if rand.Float32() < 0.5 {
		return stringLiterals[rand.Intn(len(stringLiterals))]
	}
	return numericLiterals[rand.Intn(len(numericLiterals))]
}

func generateRandomAlias() string {
	aliases := []string{"a", "b", "c", "u", "p", "o", "t1", "t2", "t3"}
	return aliases[rand.Intn(len(aliases))]
}

// Generate SQL query from template
func generateSQLFromTemplate(template string) string {
	// Count placeholders
	placeholderCount := strings.Count(template, "%s")
	args := make([]interface{}, placeholderCount)

	for i := 0; i < placeholderCount; i++ {
		switch rand.Intn(4) {
		case 0:
			args[i] = generateRandomIdentifier()
		case 1:
			args[i] = generateRandomTable()
		case 2:
			args[i] = generateRandomLiteral()
		case 3:
			args[i] = generateRandomAlias()
		}
	}

	return fmt.Sprintf(template, args...)
}

// Generate large SQL query by combining multiple templates
func generateLargeSQL(size int) string {
	var parts []string
	for i := 0; i < size; i++ {
		template := sqlTemplates[rand.Intn(len(sqlTemplates))]
		sql := generateSQLFromTemplate(template)
		parts = append(parts, sql)
	}
	return strings.Join(parts, "; ")
}

// Pathological cases that might stress the normalizer
var pathologicalCases = []string{
	// Very long identifier names
	"SELECT " + strings.Repeat("very_long_column_name_", 50) + "1 FROM " + strings.Repeat("very_long_table_name_", 50) + "1",

	// Many nested subqueries
	"SELECT * FROM (SELECT * FROM (SELECT * FROM (SELECT * FROM users WHERE id > 1) t1 WHERE name IS NOT NULL) t2 WHERE age > 18) t3 WHERE status = 'active'",

	// Large number of columns
	"SELECT " + strings.Join(generateManyColumns(100), ", ") + " FROM users",

	// Large IN clause
	"SELECT * FROM users WHERE id IN (" + strings.Join(generateManyLiterals(200), ", ") + ")",

	// Many JOINs
	generateManyJoins(20),

	// Complex CASE with many WHENs
	generateComplexCase(50),

	// Large INSERT with many values
	generateLargeInsert(100),

	// Query with many string literals containing special characters
	"SELECT * FROM users WHERE name LIKE '%john%' AND description LIKE '%it''s a \"test\"% ' AND comment = 'O''Reilly & Sons' AND data = '{\"key\": \"value\", \"array\": [1, 2, 3]}'",
}

func generateManyColumns(count int) []string {
	columns := make([]string, count)
	for i := 0; i < count; i++ {
		columns[i] = fmt.Sprintf("col_%d", i)
	}
	return columns
}

func generateManyLiterals(count int) []string {
	literals := make([]string, count)
	for i := 0; i < count; i++ {
		literals[i] = fmt.Sprintf("%d", i)
	}
	return literals
}

func generateManyJoins(count int) string {
	sql := "SELECT u.id FROM users u"
	for i := 1; i < count; i++ {
		sql += fmt.Sprintf(" JOIN table_%d t%d ON u.id = t%d.user_id", i, i, i)
	}
	return sql + " WHERE u.active = 1"
}

func generateComplexCase(count int) string {
	sql := "SELECT id, CASE"
	for i := 0; i < count; i++ {
		sql += fmt.Sprintf(" WHEN col_%d = %d THEN 'value_%d'", i, i, i)
	}
	sql += " ELSE 'default' END as result FROM table1"
	return sql
}

func generateLargeInsert(count int) string {
	sql := "INSERT INTO users (id, name, email) VALUES"
	values := make([]string, count)
	for i := 0; i < count; i++ {
		values[i] = fmt.Sprintf("(%d, 'name_%d', 'email_%d@example.com')", i, i, i)
	}
	return sql + " " + strings.Join(values, ", ")
}

// Benchmark simple queries
func BenchmarkNormalize_Simple(b *testing.B) {
	lex := lexer.NewLexer()

	for _, config := range benchConfigs {
		b.Run(config.name, func(b *testing.B) {
			sql := "SELECT id, name FROM users WHERE age = 25"
			sqlBytes := []byte(sql)
			result := make([]byte, len(sql)*2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Normalize(config.config, lex, sqlBytes, result)
			}
		})
	}
}

// Benchmark complex queries
func BenchmarkNormalize_Complex(b *testing.B) {
	lex := lexer.NewLexer()

	complexSQL := `
		SELECT u.id, u.name, u.email, p.title, p.content, 
		       COUNT(c.id) as comment_count,
		       AVG(r.rating) as avg_rating,
		       CASE 
		           WHEN u.age < 18 THEN 'minor'
		           WHEN u.age < 65 THEN 'adult'
		           ELSE 'senior'
		       END as age_category
		FROM users u
		LEFT JOIN posts p ON u.id = p.user_id
		LEFT JOIN comments c ON p.id = c.post_id
		LEFT JOIN ratings r ON p.id = r.post_id
		WHERE u.active = 1 
		  AND u.created_at > '2020-01-01'
		  AND (u.role = 'admin' OR u.role = 'moderator')
		  AND p.status IN ('published', 'featured')
		GROUP BY u.id, p.id
		HAVING COUNT(c.id) > 5 OR AVG(r.rating) > 4.0
		ORDER BY avg_rating DESC, comment_count DESC
		LIMIT 100 OFFSET 0
	`

	for _, config := range benchConfigs {
		b.Run(config.name, func(b *testing.B) {
			sqlBytes := []byte(complexSQL)
			result := make([]byte, len(complexSQL)*2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Normalize(config.config, lex, sqlBytes, result)
			}
		})
	}
}

// Benchmark pathological cases
func BenchmarkNormalize_Pathological(b *testing.B) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}

	for i, sql := range pathologicalCases {
		b.Run(fmt.Sprintf("Case_%d", i), func(b *testing.B) {
			sqlBytes := []byte(sql)
			result := make([]byte, len(sql)*3) // Extra space for pathological cases

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				Normalize(config, lex, sqlBytes, result)
			}
		})
	}
}

// Benchmark with varying SQL sizes
func BenchmarkNormalize_VaryingSizes(b *testing.B) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}

	sizes := []int{1, 5, 10, 25, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			sql := generateLargeSQL(size)
			sqlBytes := []byte(sql)
			result := make([]byte, len(sql)*2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Normalize(config, lex, sqlBytes, result)
			}
		})
	}
}

// Benchmark memory allocation patterns
func BenchmarkNormalize_MemoryAllocation(b *testing.B) {
	lex := lexer.NewLexer()
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}
	sql := "SELECT u.id, u.name FROM users u JOIN posts p ON u.id = p.user_id WHERE u.age > 18 AND p.status = 'published'"
	sqlBytes := []byte(sql)

	b.Run("PreAllocated", func(b *testing.B) {
		result := make([]byte, len(sql)*2)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Normalize(config, lex, sqlBytes, result)
		}
	})

	b.Run("NewAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := make([]byte, len(sql)*2)
			Normalize(config, lex, sqlBytes, result)
		}
	})

	b.Run("ExactSize", func(b *testing.B) {
		result := make([]byte, len(sql))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Normalize(config, lex, sqlBytes, result)
		}
	})
}

// Benchmark lexer reuse vs new instances
func BenchmarkNormalize_LexerReuse(b *testing.B) {
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}
	sql := "SELECT id, name FROM users WHERE age = 25 AND status = 'active'"
	sqlBytes := []byte(sql)
	result := make([]byte, len(sql)*2)

	b.Run("ReuseLexer", func(b *testing.B) {
		lex := lexer.NewLexer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Normalize(config, lex, sqlBytes, result)
		}
	})

	b.Run("NewLexer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			lex := lexer.NewLexer()
			Normalize(config, lex, sqlBytes, result)
		}
	})
}

// Benchmark different keyword case scenarios
func BenchmarkNormalize_KeywordIntensive(b *testing.B) {
	lex := lexer.NewLexer()

	// Query with many keywords
	keywordIntensiveSQL := `
		SELECT DISTINCT u.id, u.name 
		FROM users u 
		INNER JOIN orders o ON u.id = o.user_id 
		LEFT OUTER JOIN products p ON o.product_id = p.id 
		WHERE u.active = TRUE 
		  AND o.status IN ('completed', 'shipped') 
		  AND p.category IS NOT NULL 
		  AND EXISTS (
		      SELECT 1 FROM reviews r 
		      WHERE r.product_id = p.id 
		        AND r.rating >= 4
		  )
		GROUP BY u.id, u.name 
		HAVING COUNT(o.id) > 5 
		ORDER BY u.name ASC 
		LIMIT 100
	`

	configs := []struct {
		name   string
		config Config
	}{
		{"NoCase", Config{KeywordCase: CaseDefault}},
		{"Lower", Config{KeywordCase: CaseLower}},
		{"Upper", Config{KeywordCase: CaseUpper}},
	}

	for _, config := range configs {
		b.Run(config.name, func(b *testing.B) {
			sqlBytes := []byte(keywordIntensiveSQL)
			result := make([]byte, len(keywordIntensiveSQL)*2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Normalize(config.config, lex, sqlBytes, result)
			}
		})
	}
}

// Stress test with concurrent access (if needed)
func BenchmarkNormalize_Concurrent(b *testing.B) {
	config := Config{KeywordCase: CaseUpper, RemoveLiterals: true}
	sql := "SELECT u.id, u.name FROM users u WHERE u.age > 18 AND u.status = 'active'"
	sqlBytes := []byte(sql)

	b.RunParallel(func(pb *testing.PB) {
		lex := lexer.NewLexer()
		result := make([]byte, len(sql)*2)

		for pb.Next() {
			Normalize(config, lex, sqlBytes, result)
		}
	})
}
