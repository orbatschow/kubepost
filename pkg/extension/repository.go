package extension

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/postgres"
)

const latest = "latest"

type Repository struct {
	database *v1alpha1.Database
	instance *v1alpha1.Instance
	conn     *pgx.Conn
}

type RepositoryError struct {
	Database  string
	Instance  string
	Namespace string
	Message   string

	PostgresErrorCode    string
	PostgresErrorMessage string
}

func (e RepositoryError) Error() string {
	return e.Message
}

func (r *Repository) List(ctx context.Context) ([]v1alpha1.Extension, error) {

	var extensions []v1alpha1.Extension

	err := pgxscan.Select(
		ctx,
		r.conn,
		&extensions,
		"SELECT extname AS name, extversion AS version FROM pg_extension",
	)

	if err != nil {
		return nil, err
	}

	return extensions, nil
}

func (r *Repository) Create(ctx context.Context, extension *v1alpha1.Extension) error {
	if extension.Version == latest || extension.Version == "" {
		_, err := r.conn.Exec(
			ctx,
			fmt.Sprintf("CREATE EXTENSION %s CASCADE", postgres.SanitizeString(extension.Name)),
		)

		return err
	}
	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf(
			"CREATE EXTENSION %s WITH VERSION %s CASCADE",
			postgres.SanitizeString(extension.Name),
			postgres.SanitizeString(extension.Version),
		),
	)

	return err
}

func (r *Repository) Update(ctx context.Context, extension *v1alpha1.Extension) error {
	if extension.Version == latest {

		_, err := r.conn.Exec(
			ctx,
			fmt.Sprintf(
				"ALTER EXTENSION %s UPDATE",
				postgres.SanitizeString(extension.Name)),
		)

		return err
	}
	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf(
			"ALTER EXTENSION %s UPDATE TO %s",
			postgres.SanitizeString(extension.Name),
			postgres.SanitizeString(extension.Version),
		),
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, extension *v1alpha1.Extension) error {
	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf(
			"DROP EXTENSION %s",
			postgres.SanitizeString(extension.Name),
		),
	)
	return err
}

func (r *Repository) GetChildExtensions(ctx context.Context, extension *v1alpha1.Extension) ([]string, error) {
	var parentExtension []string

	err := pgxscan.Select(
		ctx,
		r.conn,
		&parentExtension,
		"select extname from pg_depend join pg_extension on objid = oid where refobjid=(select oid from pg_extension where extname = $1)",
		extension.Name,
	)

	if err != nil {
		return nil, err
	}
	return parentExtension, err
}
