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

func (r *databaseRepository) DoesDatabaseExist(name string) (bool, error) {

	var exist bool
	err := r.conn.QueryRow(
		context.Background(),
		"SELECT true FROM pg_database WHERE datname = $1",
		name,
	).Scan(&exist)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			log.Errorf(
				"unable to check database '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)

			return false, err
		}
		if err.Error() == "no rows in result set" {
			return false, nil
		}
	}

	return true, nil
}

func (r *databaseRepository) Create(name string) error {

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf("CREATE DATABASE %s", SanitizeString(name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "42P04" {
				log.Infof("database '%s' already exists, skipping creation", name)
				return nil
			}
			log.Errorf(
				"unable to create database '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)
			return err
		}
	}

	return nil
}

func (r *databaseRepository) Delete(name string) error {

	_, err := r.conn.Query(
		context.Background(),
		fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", SanitizeString(name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Errorf(
				"unable to delete database '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)
			return err
		}
	}

	return nil
}

func (r *databaseRepository) AlterOwner(db string, name string) error {

	var currentOwner string
	err := r.conn.QueryRow(
		context.Background(),
		fmt.Sprintf(
			"select r.rolname from pg_roles as r, pg_database as d where r.oid = d.datdba AND d.datname = '%s'",
			db,
		),
	).Scan(&currentOwner)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Errorf(
				"unable to query for database ownership of '%s', failed with code: '%s' and message: '%s'",
				db,
				pgErr.Code,
				pgErr.Message,
			)
			return err
		}
	}

	if name == currentOwner {
		log.Infof(
			"role %s already owns database %s",
			name,
			db,
		)
		return nil
	}

	_, err = r.conn.Query(
		context.Background(),
		fmt.Sprintf(
			"ALTER DATABASE %s OWNER TO %s",
			SanitizeString(db),
			SanitizeString(name),
		),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Errorf(
				"unable to alter database ownership to '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)
			return err
		}
	}

	return nil
}
