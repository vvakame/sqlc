package spanner

import (
	"fmt"

	spannerast "github.com/cloudspannerecosystem/memefish/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

func convertCreateTable(node *spannerast.CreateTable) (*ast.CreateTableStmt, error) {
	create := &ast.CreateTableStmt{
		Name: &ast.TableName{
			Name: node.Name.SQL(),
		},
		IfNotExists: node.IfNotExists,
	}

	for _, col := range node.Columns {
		column, err := convertColumnDef(col)
		if err != nil {
			return nil, err
		}
		create.Cols = append(create.Cols, column)
	}

	return create, nil
}

func convertColumnDef(node *spannerast.ColumnDef) (*ast.ColumnDef, error) {
	colDef := &ast.ColumnDef{
		Colname:    node.Name.Name,
		IsNotNull:  node.NotNull,
		PrimaryKey: node.PrimaryKey,
		Location:   int(node.Pos()),
	}

	typeName, err := convertSchemaType(node.Type)
	if err != nil {
		return nil, err
	}
	colDef.TypeName = typeName

	return colDef, nil
}

func convertSchemaType(node spannerast.SchemaType) (*ast.TypeName, error) {
	switch node := node.(type) {
	case *spannerast.ScalarSchemaType:
		return &ast.TypeName{
			Name:     string(node.Name),
			Location: int(node.Pos()),
		}, nil

	case *spannerast.SizedSchemaType:
		return &ast.TypeName{
			Name:     string(node.Name),
			Location: int(node.Pos()),
		}, nil

	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}
