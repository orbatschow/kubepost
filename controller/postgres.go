package controller

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

var pool *pgxpool.Pool

type Postgres struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode string
}

// TODO: make sslmode configurable
func (p *Postgres) GetConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", p.Username, p.Password, p.Host, p.Port, p.Database))
	if err != nil {
		log.Errorf("Unable to connect to database: '%s' on host '%s' with user '%s' : '%s'", p.Database, p.Host, p.Username, err)
		return nil, err
	}
	return conn, err
}
