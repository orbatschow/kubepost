package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type roleRepository struct {
	conn *pgx.Conn
}

func NewRoleRepository(conn *pgx.Conn) roleRepository {
	return roleRepository{
		conn: conn,
	}
}

func (r roleRepository) DoesRoleExist(name string) (bool, error) {

	var exist bool
	err := r.conn.QueryRow(
		context.Background(),
		fmt.Sprintf(
			"SELECT true FROM pg_roles WHERE rolname = '%s'",
			SanitizeString(name),
		),
	).Scan(&exist)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			log.Errorf(
				"unable to get role '%s', failed with code: '%s' and message: '%s'",
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

func (r *roleRepository) Create(name string) error {

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf("CREATE ROLE %s", SanitizeString(name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "42710" || pgErr.Code == "23505" {
				log.Infof("role '%s' already exists, skipping creation", name)
				return nil
			}

			log.Errorf(
				"unable to create role '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)

			return err
		}
	}

	return nil
}

func (r *roleRepository) Delete(name string) error {

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf("DROP ROLE %s", SanitizeString(name)),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "42704" {
				log.Infof("role '%s' does not exist, skipping deletion", name)
				return nil
			}

			log.Errorf(
				"unable to delete role '%s', failed with code: '%s' and message: '%s'",
				name,
				pgErr.Code,
				pgErr.Message,
			)

			return err
		}
	}

	return nil
}

func (r *roleRepository) SetPassword(name string, password string) error {

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf(
			"ALTER ROLE %s WITH PASSWORD '%s';",
			SanitizeString(name),
			password,
		),
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Errorf(
			"unable to set password for role '%s', failed with code: '%s' and message: '%s'",
			name,
			pgErr.Code,
			pgErr.Message,
		)
		return err
	}

	return nil
}

func (r *roleRepository) Alter(role *v1alpha1.Role) error {

	// if no Options were given, return without effect
	if len(role.Spec.Options) == 0 {
		return nil
	}

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf(
			"ALTER ROLE %s WITH %s;",
			SanitizeString(role.Spec.RoleName),
			strings.Join(role.Spec.Options[:], " "),
		),
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {

		log.Errorf(
			"unable to alter permissions to role '%s', failed with code: '%s' and message: '%s'",
			role.Spec.RoleName,
			pgErr.Code,
			pgErr.Message,
		)

		return err
	}
	return nil
}

func (r *roleRepository) Grant(role *v1alpha1.Role) error {

	// grant/revoke all grants
	for _, grant := range role.Spec.Grants {

		if grant.Database == "" && grant.Schema == "" {
			log.Error("either schema or database has to be defined within a grant")
			return errors.New("either schema or database has to be defined within a grant")
		}

		// revoke permissions
		_, err := r.conn.Exec(
			context.Background(),
			createRevokeQuery(role.Spec.RoleName, &grant),
		)

		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {

				log.Errorf(
					"unable to revoke grants from role '%s', failed with code: '%s' and message: '%s'",
					role.Spec.RoleName,
					pgErr.Code,
					pgErr.Message,
				)

				return err
			}
		}

		// grant permissions
		_, err = r.conn.Exec(
			context.Background(),
			createGrantQuery(role.Spec.RoleName, &grant),
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {

				log.Errorf(
					"unable to apply grants to role '%s', failed with code: '%s' and message: '%s'",
					r,
					pgErr.Code,
					pgErr.Message,
				)

				return err
			}
		}
	}

	return nil
}

func createGrantQuery(roleName string, grant *v1alpha1.Grant) string {
	var query string

	switch strings.ToUpper(grant.ObjectType) {
	case "DATABASE":
		query = fmt.Sprintf(
			"GRANT %s ON DATABASE %s TO %s",
			strings.Join(grant.Privileges, ","),
			SanitizeString(grant.Database),
			SanitizeString(roleName),
		)
	case "SCHEMA":
		query = fmt.Sprintf(
			"GRANT %s ON SCHEMA %s TO %s",
			strings.Join(grant.Privileges, ","),
			SanitizeString(grant.Schema),
			SanitizeString(roleName),
		)
	case "TABLE", "SEQUENCE", "FUNCTION":
		query = fmt.Sprintf(
			"GRANT %s ON ALL %sS IN SCHEMA %s TO %s",
			strings.Join(grant.Privileges, ","),
			SanitizeString(grant.ObjectType),
			SanitizeString(grant.Schema),
			SanitizeString(roleName),
		)
	}

	// TODO ASO
	/*
		if d.DoesRoleExist("with_grant_option").(bool) == true {
			query = query + " WITH GRANT OPTION"
		}
	*/

	return query
}

func createRevokeQuery(roleName string, grant *v1alpha1.Grant) string {
	var query string

	switch strings.ToUpper(grant.ObjectType) {
	case "DATABASE":
		query = fmt.Sprintf(
			"REVOKE ALL PRIVILEGES ON DATABASE %s FROM %s",
			SanitizeString(grant.Database),
			SanitizeString(roleName),
		)
	case "SCHEMA":
		query = fmt.Sprintf(
			"REVOKE ALL PRIVILEGES ON SCHEMA %s FROM %s",
			SanitizeString(grant.Schema),
			SanitizeString(roleName),
		)
	case "TABLE", "SEQUENCE", "FUNCTION":
		query = fmt.Sprintf(
			"REVOKE ALL PRIVILEGES ON ALL %sS IN SCHEMA %s FROM %s",
			SanitizeString(grant.ObjectType),
			SanitizeString(grant.Schema),
			SanitizeString(roleName),
		)
	}

	return query
}
