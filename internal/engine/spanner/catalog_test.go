package spanner

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/catalog"
)

func TestDDLs(t *testing.T) {
	p := NewParser()

	for i, tc := range []struct {
		stmt string
		s    *catalog.Schema
	}{
		{
			`
			CREATE TABLE Foos ( Bar STRING(MAX) );
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "Foos"},
						Columns: []*catalog.Column{
							{
								Name: "Bar",
								Type: ast.TypeName{Name: "STRING"},
							},
						},
					},
				},
			},
		},
	} {
		test := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stmts, err := p.Parse(strings.NewReader(test.stmt))
			if err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			c := newTestCatalog()
			if err := c.Build(stmts); err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			e := newTestCatalog()
			if test.s != nil {
				var replaced bool
				for i := range e.Schemas {
					if e.Schemas[i].Name == test.s.Name {
						e.Schemas[i] = test.s
						replaced = true
						break
					}
				}
				if !replaced {
					e.Schemas = append(e.Schemas, test.s)
				}
			}

			if diff := cmp.Diff(e, c, cmpDiffOpts()...); diff != "" {
				t.Log(test.stmt)
				t.Errorf("catalog mismatch:\n%s", diff)
			}
		})
	}
}

func TestDMLs(t *testing.T) {
	p := NewParser()

	for i, tc := range []struct {
		stmt   string
		parsed []ast.Statement
	}{
		{
			`
			SELECT 1
			`,
			[]ast.Statement{
				{
					Raw: &ast.RawStmt{
						Stmt: &ast.SelectStmt{},
					},
				},
			},
		},
		{
			`
			SELECT * FROM Singers WHERE SingerId = 1
			`,
			[]ast.Statement{
				{
					Raw: &ast.RawStmt{
						Stmt: &ast.SelectStmt{
							FromClause: &ast.List{
								Items: []ast.Node{
									&ast.TableName{Name: "Singers"},
								},
							},
						},
					},
				},
			},
		},
		{
			`
			SELECT * FROM Singers WHERE SingerId = @singerId
			`,
			[]ast.Statement{
				{
					Raw: &ast.RawStmt{
						Stmt: &ast.SelectStmt{
							FromClause: &ast.List{
								Items: []ast.Node{
									&ast.TableName{Name: "Singers"},
								},
							},
						},
					},
				},
			},
		},
	} {
		test := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stmts, err := p.Parse(strings.NewReader(test.stmt))
			if err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.parsed, stmts, cmpDiffOpts()...); diff != "" {
				t.Log(test.stmt)
				t.Errorf("catalog mismatch:\n%s", diff)
			}
		})
	}
}

func cmpDiffOpts() cmp.Options {
	return cmp.Options{
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreUnexported(catalog.Column{}),
		cmpopts.IgnoreFields(ast.ColumnDef{}, "Location"),
		cmpopts.IgnoreFields(ast.TypeName{}, "Location"),
		cmpopts.IgnoreFields(ast.RawStmt{}, "StmtLocation"),
		cmpopts.IgnoreFields(ast.RawStmt{}, "StmtLen"),
	}
}
