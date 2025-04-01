package spanner

import (
	"fmt"
	"strconv"

	spannerast "github.com/cloudspannerecosystem/memefish/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

func notImplemented() ast.Node {
	return &ast.TODO{}
}

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

	case *spannerast.ArraySchemaType:
		// TODO: Temporary implementation to prevent errors. Will be properly implemented later.
		return &ast.TypeName{
			Name:     node.Item.SQL(),
			Location: int(node.Pos()),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %T, %q", node, node.SQL())
	}
}

func convertCreateIndex(node *spannerast.CreateIndex) *ast.IndexStmt {
	strPtr := func(s string) *string {
		return &s
	}
	indexStmt := &ast.IndexStmt{
		Idxname:     strPtr(node.Name.SQL()), // TODO
		Unique:      node.Unique,
		IfNotExists: node.IfNotExists,
	}

	return indexStmt
}

func convertQueryStatement(node *spannerast.QueryStatement) (*ast.SelectStmt, error) {
	// NOTE: node.Hint can be ignored with sqlc context.
	return convertQueryExpr(node.Query)
}

func convertQueryExpr(node spannerast.QueryExpr) (*ast.SelectStmt, error) {
	switch node := node.(type) {
	case *spannerast.Select:
		return convertSelect(node)
	default:
		return nil, fmt.Errorf("unsupported node type: %T, %q", node, node.SQL())
	}
}

func convertSelect(node *spannerast.Select) (*ast.SelectStmt, error) {
	if node == nil {
		return nil, nil
	}

	selectStmt := &ast.SelectStmt{
		TargetList: &ast.List{
			Items: convertSelectItemList(node.Results),
		},
		FromClause:  convertFrom(node.From),
		WhereClause: convertWhere(node.Where),
	}

	switch node.AllOrDistinct {
	case "":
		// ignore
	case spannerast.AllOrDistinctAll:
		selectStmt.All = true
	default:
		return nil, fmt.Errorf("unsupported node type: %T, %q", node, node.AllOrDistinct)
	}

	// TODO: support more attributes

	return selectStmt, nil
}

func convertSelectItemList(list []spannerast.SelectItem) []ast.Node {
	var items []ast.Node
	for _, item := range list {
		items = append(items, convertSelectItem(item))
	}
	return items
}

func convertSelectItem(node spannerast.SelectItem) ast.Node {
	switch node := node.(type) {
	case *spannerast.Star:
		star := &ast.A_Star{}
		return star
	case *spannerast.DotStar:
		_ = node
		// TODO
		return notImplemented()
	default:
		// TODO
		return notImplemented()
	}
}

func convertFrom(node *spannerast.From) *ast.List {
	if node == nil {
		return nil
	}

	list := &ast.List{}

	tableExpr, err := convertTableExpr(node.Source)
	if err != nil {
		return nil
	}

	list.Items = append(list.Items, tableExpr)

	return list
}

func convertTableExpr(node spannerast.TableExpr) (ast.Node, error) {
	if node == nil {
		return nil, nil
	}

	switch node := node.(type) {
	case *spannerast.TableName:
		if node.As == nil {
			tableName := &ast.TableName{
				Name: node.Table.Name,
			}
			return tableName, nil
		}
		// TODO
		return notImplemented(), nil
	default:
		// TODO
		return notImplemented(), nil
	}
}

func convertWhere(node *spannerast.Where) ast.Node {
	if node == nil {
		return nil
	}

	return convertExpr(node.Expr)
}

func convertExpr(node spannerast.Expr) ast.Node {
	switch node := node.(type) {
	case *spannerast.Ident:
		return &ast.A_Expr{
			Name:     &ast.List{Items: []ast.Node{&ast.String{Str: node.Name}}},
			Location: int(node.Pos()),
		}
	case *spannerast.IntLiteral:
		v, err := strconv.ParseInt(node.Value, node.Base, 64)
		if err != nil {
			// TODO
			return notImplemented()
		}
		return &ast.A_Expr{
			Name:     &ast.List{Items: []ast.Node{&ast.Integer{Ival: v}}},
			Location: int(node.Pos()),
		}
	case *spannerast.BinaryExpr:
		aExpr := &ast.A_Expr{
			Name:     &ast.List{Items: []ast.Node{&ast.String{Str: string(node.Op)}}},
			Lexpr:    convertExpr(node.Left),
			Rexpr:    convertExpr(node.Right),
			Location: int(node.Pos()),
		}
		return aExpr
	case *spannerast.Param:
		// TODO
		return &ast.Param{
			Xpr:      &ast.String{Str: node.Name},
			Location: int(node.Pos()),
		}
	default:
		// TODO
		return notImplemented()
	}
}
