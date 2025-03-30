package spanner

import (
	"fmt"
	"io"

	"github.com/cloudspannerecosystem/memefish"
	spannerast "github.com/cloudspannerecosystem/memefish/ast"
	"github.com/sqlc-dev/sqlc/internal/source"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct {
}

func (p *Parser) Parse(r io.Reader) ([]ast.Statement, error) {
	blob, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// TODO input file name?
	spStmts, err := memefish.ParseStatements("sqlc-internal.sql", string(blob))
	if err != nil {
		return nil, err
	}

	var stmts []ast.Statement
	for _, spStmt := range spStmts {
		n, err := translate(spStmt)
		if err != nil {
			return nil, err
		}
		if n == nil {
			return nil, fmt.Errorf("unexpected nil node")
		}
		stmts = append(stmts, ast.Statement{
			Raw: &ast.RawStmt{
				Stmt:         n,
				StmtLocation: int(spStmt.Pos()),
				StmtLen:      len(spStmt.SQL()),
			},
		})
	}

	return stmts, nil
}

func translate(stmt spannerast.Statement) (ast.Node, error) {
	switch expr := stmt.(type) {
	case *spannerast.CreateTable:
		return convertCreateTable(expr)
	case *spannerast.CreateIndex:
		return convertCreateIndex(expr), nil
	case *spannerast.QueryStatement:
		return convertQueryStatement(expr)
	default:
		return nil, fmt.Errorf("unexpected statement type: %T. %q", expr, expr.SQL())
	}
}

func (p *Parser) CommentSyntax() source.CommentSyntax {
	// https://cloud.google.com/spanner/docs/reference/standard-sql/lexical#comments
	return source.CommentSyntax{
		Dash:      true,
		Hash:      true,
		SlashStar: true,
	}
}
