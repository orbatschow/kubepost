package role

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/postgres"
	"github.com/orbatschow/kubepost/pkg/secret"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

type Repository struct {
	role       *v1alpha1.Role
	connection *v1alpha1.Connection
	conn       *pgx.Conn
}

type RepositoryError struct {
	Role       string
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

	var exist bool
	err := r.conn.QueryRow(
		ctx,
		"SELECT true FROM pg_roles WHERE rolname = $1",
		r.role.ObjectMeta.Name,
	).Scan(&exist)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *Repository) Create(ctx context.Context) error {

	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf("CREATE ROLE %s", postgres.SanitizeString(r.role.ObjectMeta.Name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "42710" || pgErr.Code == "23505" {
				log.FromContext(ctx).Info("postgres role already exists, skipping creation")
				return nil
			}

			return err
		}
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context) error {

	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf("DROP ROLE %s", postgres.SanitizeString(r.role.ObjectMeta.Name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "42704" {
				// TODO Withconte("postgres role does not exist, skipping deletion", ctx)
				return nil
			}

			return err
		}
	}

	return nil
}

func (r *Repository) SetPassword(ctx context.Context, password string) error {

	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf(
			"ALTER ROLE %s WITH PASSWORD '%s';",
			postgres.SanitizeString(r.role.ObjectMeta.Name),
			password,
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

		return RepositoryError{
			Role:                 r.role.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.role.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	return nil
}

func (r *Repository) GetPassword(ctx context.Context, ctrlClient client.Client) (string, error) {

	// if neither password isn't set, set it to NULL
	if r.role.Spec.Password == nil {
		return "NULL", nil
	}

	namespacedName := types.NamespacedName{
		Name:      r.role.Spec.Password.Name,
		Namespace: r.role.ObjectMeta.Namespace,
	}

	passwordSecret, err := secret.Get(ctx, ctrlClient, namespacedName)
	if err != nil {
		return "", err
	}

	// extract the password
	buffer := passwordSecret.Data[r.role.Spec.Password.Key]
	if buffer == nil {
		return "",
			fmt.Errorf(
				"could not find key '%s' for secret '%s' in namespace '%s' for role '%s'",
				r.role.Spec.Password.Key,
				r.role.Spec.Password.Name,
				r.role.Namespace,
				r.role.Name,
			)
	}

	return string(buffer), nil
}

func (r *Repository) Alter(ctx context.Context) error {

	// if no options were given, return without effect
	if len(r.role.Spec.Options) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"ALTER ROLE %s WITH %s;",
		postgres.SanitizeString(r.role.ObjectMeta.Name),
		strings.Join(r.role.Spec.Options[:], " "),
	)

	log.FromContext(ctx).Info("computed alter role query", "query", query)

	_, err := r.conn.Exec(
		ctx,
		query,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return RepositoryError{
			Role:                 r.role.ObjectMeta.Name,
			Connection:           r.connection.ObjectMeta.Name,
			Namespace:            r.role.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	return nil
}
