package golang

import (
	"github.com/sqlc-dev/sqlc/internal/codegen/golang/opts"
	"github.com/sqlc-dev/sqlc/internal/codegen/sdk"
	"github.com/sqlc-dev/sqlc/internal/plugin"
)

func spannerType(req *plugin.GenerateRequest, options *opts.Options, col *plugin.Column) string {
	columnType := sdk.DataType(col.Type)
	notNull := col.NotNull || col.IsArray
	// unsigned := col.Unsigned

	// TODO: just copied from mysql. needs reimplementation
	// https://cloud.google.com/spanner/docs/reference/standard-sql/data-definition-language#data_types

	switch columnType {
	case "STRING":
		if notNull {
			return "string"
		}
		return "spanner.NullString"

	default:
		panic("unknown column type")
	}
}
