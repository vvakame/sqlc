package spanner

import (
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/catalog"
)

func defaultSchema(name string) *catalog.Schema {
	s := &catalog.Schema{Name: name}
	s.Funcs = []*catalog.Function{
		{
			Name:       "CURRENT_TIMESTAMP",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "TIMESTAMP"},
		},
	}

	return s
}
