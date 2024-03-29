package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/postgres"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Repository struct {
	database   *v1alpha1.Database
	connection *v1alpha1.Connection
	conn       *pgx.Conn
}

type RepositoryError struct {
	Database   string
	Connection string
	Namespace  string
	Message    string

	PostgresErrorCode    string
	PostgresErrorMessage string
}

func (e RepositoryError) Error() string {
	return e.Message
}

func (r *Repository) Exists(ctx context.Context) (bool, error) {
	log.FromContext(ctx).Info(
		"checking if database already exists",
		"database", r.database.ObjectMeta.Name,
		"connection", r.connection.ObjectMeta.Name,
		"namespace", r.connection.ObjectMeta.Namespace,
	)

	var exists bool
	err := r.conn.QueryRow(
		ctx,
		"SELECT true FROM pg_database WHERE datname = $1",
		r.database.ObjectMeta.Name,
	).Scan(&exists)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}

		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return false, &RepositoryError{
			Database:             r.database.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.database.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	return true, nil
}

func (r *Repository) Create(ctx context.Context) error {

	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s", postgres.SanitizeString(r.database.ObjectMeta.Name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return &RepositoryError{
			Database:             r.database.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.database.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	log.FromContext(ctx).Info("created database",
		"connection", types.NamespacedName{
			Namespace: r.connection.ObjectMeta.Namespace,
			Name:      r.connection.ObjectMeta.Name,
		},
	)

	return nil
}

func (r *Repository) Delete(ctx context.Context) *RepositoryError {
	_, err := r.conn.Query(
		ctx,
		fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", postgres.SanitizeString(r.database.ObjectMeta.Name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return &RepositoryError{
			Database:             r.database.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.database.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	log.FromContext(ctx).Info("deleted database",
		"connection", types.NamespacedName{
			Namespace: r.connection.ObjectMeta.Namespace,
			Name:      r.connection.ObjectMeta.Name,
		},
	)

	return nil
}

func (r *Repository) AlterOwner(ctx context.Context) error {

	var currentOwner string
	err := r.conn.QueryRow(
		context.Background(),
		fmt.Sprintf(
			"select r.rolname from pg_roles as r, pg_database as d where r.oid = d.datdba AND d.datname = '%s'",
			r.database.ObjectMeta.Name,
		),
	).Scan(&currentOwner)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return &RepositoryError{
			Database:             r.database.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.database.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	if r.database.Spec.Owner == currentOwner {
		log.FromContext(ctx).Info(
			"skipping ownership change",
			"owner", r.database.Spec.Owner,
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)
		return nil
	}

	_, err = r.conn.Query(
		ctx,
		fmt.Sprintf(
			"ALTER DATABASE %s OWNER TO %s",
			postgres.SanitizeString(r.database.ObjectMeta.Name),
			postgres.SanitizeString(r.database.Spec.Owner),
		),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return &RepositoryError{
			Database:             r.database.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.database.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	log.FromContext(ctx).Info("changed ownership",
		"connection", types.NamespacedName{
			Namespace: r.connection.ObjectMeta.Namespace,
			Name:      r.connection.ObjectMeta.Name,
		},
	)

	return nil
}
