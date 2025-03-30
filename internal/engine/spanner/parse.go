package spanner

import (
	"errors"
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

	var ddls []spannerast.DDL
	// TODO input file name?
	ddls, err = memefish.ParseDDLs("sqlc-ddls.sql", string(blob))
	var tryDMLs bool
	{
		var spErr *memefish.Error
		// TODO correct error handling
		if errors.As(err, &spErr); spErr != nil {
			tryDMLs = true
			ddls = nil
		} else if err != nil {
			return nil, err
		}
	}

	var dmls []spannerast.DML
	if tryDMLs {
		var spErr *memefish.Error
		// TODO input file name?
		dmls, err = memefish.ParseDMLs("sqlc-internal.sql", string(blob))
		// TODO correct error handling
		if errors.As(err, &spErr); spErr != nil {
			dmls = nil
		} else if err != nil {
			return nil, err
		}
	}

	var stmts []ast.Statement
	for _, ddl := range ddls {
		n, err := translateDDL(ddl)
		if err != nil {
			return nil, err
		}
		if n == nil {
			return nil, fmt.Errorf("unexpected nil node")
		}
		stmts = append(stmts, ast.Statement{
			Raw: &ast.RawStmt{
				Stmt:         n,
				StmtLocation: int(ddl.Pos()),
				StmtLen:      len(ddl.SQL()),
			},
		})
	}

	for _, dml := range dmls {
		n, err := translateDML(dml)
		if err != nil {
			return nil, err
		}
		if n == nil {
			return nil, fmt.Errorf("unexpected nil node")
		}
		stmts = append(stmts, ast.Statement{
			Raw: &ast.RawStmt{
				Stmt:         n,
				StmtLocation: int(dml.Pos()),
				StmtLen:      len(dml.SQL()),
			},
		})
	}

	return stmts, nil
}

func translateDDL(stmt spannerast.DDL) (ast.Node, error) {
	switch expr := stmt.(type) {
	case *spannerast.CreateTable:
		return convertCreateTable(expr)
	default:
		return &ast.TODO{}, nil
	}
}

func translateDML(stmt spannerast.DML) (ast.Node, error) {
	// TODO
	return &ast.TODO{}, nil
}

func (p *Parser) CommentSyntax() source.CommentSyntax {
	// https://cloud.google.com/spanner/docs/reference/standard-sql/lexical#comments
	return source.CommentSyntax{
		Dash:      true,
		Hash:      true,
		SlashStar: true,
	}
}
