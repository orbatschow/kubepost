package postgres

import "github.com/jackc/pgx/v4/pgxpool"

var pool *pgxpool.Pool

type Postgres struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string
}
