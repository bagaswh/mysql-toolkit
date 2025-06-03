package queryfingerprint

import (
	"fmt"
	"reflect"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
)

type FingerprintOptions struct {
	ValidateSQL      bool
	KeywordsCase     string
	MergeWhitespaces bool
	OneLiner         bool
}

type position struct {
	start, end int
}

type fingerprintVisitor struct {
	keywordPositions []position
	valuePositions   []position
}

func (v *fingerprintVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch typ := in.(type) {
	case ast.ValueExpr:
		fmt.Println(reflect.TypeOf(typ))
	}
	return in, true
}

func (v *fingerprintVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func Fingerprint(q string, options ...FingerprintOptions) (string, error) {
	p := parser.New()

	stmtNodes, _, err := p.ParseSQL(q)
	if err != nil {
		return "", fmt.Errorf("failed parsing sql: %w", err)
	}

	for _, stmtNode := range stmtNodes {
		visitor := &fingerprintVisitor{}
		stmtNode.Accept(visitor)
	}

	return "", nil
}
