package repository

import (
	"github.com/jackc/pgx/v4"
)

func SanitizeString(input string) string {
	var ids pgx.Identifier
	ids = append(ids, input)
	return ids.Sanitize()
}
