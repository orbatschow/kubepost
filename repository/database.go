package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
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

func (r *databaseRepository) ReconcileExtensions(extensions []v1alpha1.Extension) error {

	type pgExtension struct {
		Name    string `json:"extname"`
		Version string `json:"extversion"`
	}
	var pgExtensions []*pgExtension

	err := pgxscan.Select(context.Background(), r.conn, &pgExtensions, "SELECT extname AS name, extversion AS version FROM pg_extension")

	for _, match := range pgExtensions {
		
	}

	if err != nil {
		log.Error(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			//log.Errorf("unable to delete database '%s', failed with code: '%s' and message: '%s'", pgErr.Code, pgErr.Message)
			log.Error(err)
			return err
		}
	}

	return nil
}
