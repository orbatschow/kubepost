package postgres

import "github.com/jackc/pgx/v4"

const (
	TABLE    = "TABLE"
	FUNCTION = "FUNCTION"
	SCHEMA   = "SCHEMA"
	COLUMN   = "COLUMN"
	VIEW     = "VIEW"
	SEQUENCE = "SEQUENCE"
)

func SanitizeString(input string) string {
	var ids pgx.Identifier
	ids = append(ids, input)
	return ids.Sanitize()
}
