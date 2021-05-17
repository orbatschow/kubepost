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

func (r *roleRepository) DoesRoleExist(name string) (bool, error) {

	var exist bool
	err := r.conn.QueryRow(
		context.Background(),
		"SELECT true FROM pg_roles WHERE rolname = $1",
		name,
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

func (r *roleRepository) Grant(role *v1alpha1.Role, grant *v1alpha1.Grant) error {

	// TODO matching of existing grants and revoking unwanted !!!

	for _, grantObject := range grant.Objects {
		query, err := createGrantQuery(
			role.Spec.RoleName,
			&grantObject,
		)

		// if creation of statement failed
		if err != nil {
			log.Errorf(err.Error())
			continue // Continue with next grant-statement
		}

		_, err = r.conn.Exec(
			context.Background(),
			query,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {

				log.Errorf(
					"unable to apply grant to role '%s', failed with code: '%s' and message: '%s'",
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

func createGrantQuery(roleName string, grantTarget *v1alpha1.GrantObject) (string, error) {
	possiblePrivileges := map[string]bool{
		"ALL":        true,
		"INSERT":     true,
		"SELECT":     true,
		"UPDATE":     true,
		"DELETE":     true,
		"TRUNCATE":   true,
		"REFERENCES": true,
		"TRIGGER":    true,
	}

	for _, privilege := range grantTarget.Privileges {
		if !possiblePrivileges[privilege] {
			return "", fmt.Errorf("privilege %s unknown", privilege) // TODO  logging
		}
	}

	var query string
	switch strings.ToUpper(grantTarget.Type) {
	case "TABLE":
		{
			query = fmt.Sprintf(
				"GRANT %s ON TABLE %s TO %s",
				strings.Join(grantTarget.Privileges, ","),
				grantTarget.Identifier,
				roleName,
			)
		}
	case "SCHEMA":
		{
			query = fmt.Sprintf(
				"GRANT %s ON ALL TABLES IN SCHEMA %s TO %s",
				strings.Join(grantTarget.Privileges, ","),
				grantTarget.Identifier,
				roleName,
			)
		}
	case "FUNCTION":
		{
			query = fmt.Sprintf(
				"GRANT %s ON FUNCTION %s TO %s",
				strings.Join(grantTarget.Privileges, ","),
				grantTarget.Identifier,
				roleName,
			)
		}
	case "SEQUENCE":
		{
			query = fmt.Sprintf(
				"GRANT %s ON SEQUENCE %s TO %s",
				strings.Join(grantTarget.Privileges, ","),
				grantTarget.Identifier,
				roleName,
			)
		}
	case "ROLE":
		{
			query = fmt.Sprintf(
				"GRANT %s TO %s",
				grantTarget.Identifier,
				roleName,
			)
			if grantTarget.WithAdminOption {
				query = " WITH ADMIN OPTION"
			}
		}
	default:
		{
			return "", fmt.Errorf("grant type %s unknown", grantTarget.Type) // TODO  logging
		}
	}

	if strings.ToUpper(grantTarget.Type) != "ROLE" && grantTarget.WithGrantOption {
		query += " WITH GRANT OPTION"
	}

	return query, nil
}
