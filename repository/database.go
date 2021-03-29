package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type databaseRepository struct {
	conn *pgx.Conn
}

func NewDatabaseRepository(conn *pgx.Conn) databaseRepository {
	return databaseRepository{
		conn: conn,
	}
}

func (r *databaseRepository) Create(name string) error {

	_, err := r.conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", name))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "42P04" {
				log.Infof("database '%s' already exists, skipping creation", name)
				return nil
			}
			log.Errorf("unable to create database '%s', failed with code: '%s' and message: '%s'", name, pgErr.Code, pgErr.Message)
			return err
		}
	}

	return nil
}

func (r *databaseRepository) Delete(name string) error {

	_, err := r.conn.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", name))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Errorf("unable to delete database '%s', failed with code: '%s' and message: '%s'", name, pgErr.Code, pgErr.Message)
			return err
		}
	}

	return nil
}

func (r *databaseRepository) ReconcileExtensions(extensions []string) error {

	// TODO
	_, err := r.conn.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", name))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Errorf("unable to delete database '%s', failed with code: '%s' and message: '%s'", name, pgErr.Code, pgErr.Message)
			return err
		}
	}

	return nil
}
