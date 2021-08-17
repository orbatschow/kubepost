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

func (r *roleRepository) getGrantsByType(role *v1alpha1.Role, grantType string) ([]v1alpha1.GrantObject, error) {
	var grantQuery string

	switch strings.ToUpper(grantType) {
	case "TABLE":
		grantQuery = `
        select
        table_name as identifier,
        'TABLE' as type,
        table_schema as schema,
        array_agg(cast(privilege_type AS text)) as privileges,
        is_grantable::bool as withGrantOption
        from information_schema.role_table_grants
        WHERE grantee=$1
        GROUP BY identifier, schema, withGrantOption
        `

	case "SCHEMA":
		grantQuery = `
        SELECT
        identifier,
        type,
        schema,
        array_agg(privileges),
        withGrantOption
        FROM
        (SELECT
        nspname as identifier,
        'SCHEMA' as type,
        'public' as schema,
        (aclexplode(nspacl)).privilege_type as privileges,
        (aclexplode(nspacl)).is_grantable as withGrantOption,
        (aclexplode(nspacl)).grantee as grantee
        FROM pg_catalog.pg_namespace) sub
        WHERE grantee = (SELECT oid FROM pg_catalog.pg_roles where rolname=$1)
        GROUP BY identifier, type, schema, withGrantOption`

	case "FUNCTION":
		grantQuery = `
        select
        table_name as identifier,
        'FUNCTION' as type,
        table_schema as schema,
        array_agg(privilege_type) as priviliges,
        is_grantable::bool as withGrantOption
        from information_schema.role_routine_grants 
        WHERE grantee=$1
        GROUP BY identifier, type, schema, withGrantOption`

	case "SEQUENCE":
		grantQuery = `
        select
        table_name as identifier,
        'FUNCTION' as type,
        table_schema as schema,
        array_agg(privilege_type) as priviliges,
        is_grantable::bool as withGrantOption
        from information_schema.role_routine_grants 
        WHERE grantee=$1
        GROUP BY identifier, type, schema, withGrantOption`

	case "ROLE":
		grantQuery = `
        SELECT 
        u.rolname,
		'ROLE' as type,
        'public' as schema,
        NULL as privileges,
        admin_option as withGrantOption
        FROM pg_catalog.pg_auth_members m
        JOIN pg_catalog.pg_authid u on (m.roleid = u.oid)
        WHERE m.member = (select oid from pg_authid where rolname=$1)`

	default:
		return nil, fmt.Errorf("grant type %s unknown or unsupported", grantType)
	}

	var grants []v1alpha1.GrantObject

	rows, err := r.conn.Query(
		context.Background(),
		grantQuery,
		role.Spec.RoleName,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var privileges []string
		var grant v1alpha1.GrantObject
		err = rows.Scan(
			&grant.Identifier,
			&grant.Type,
			&grant.Schema,
			&privileges,
			&grant.WithGrantOption,
		)
		if err != nil {
			return nil, err
		}

		grant.Privileges = StringArrayToPrivilegArray(privileges)

		//grant.Privileges = []v1alpha1.Privilege{v1alpha1.Privilege(privileges[0])}
		grants = append(grants, grant)
	}

	return grants, nil
}

func (r *roleRepository) getAllCurrentGrants(role *v1alpha1.Role) ([]v1alpha1.GrantObject, error) {
	var currentGrants []v1alpha1.GrantObject
	var buffer []v1alpha1.GrantObject

	// get table grants
	buffer, err := r.getGrantsByType(role, "TABLE")
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, err
	}

	// get schema grants
	buffer, err = r.getGrantsByType(role, "SCHEMA")
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, err
	}

	// get role grants
	buffer, err = r.getGrantsByType(role, "ROLE")
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, err
	}

	return currentGrants, err
}

func (r *roleRepository) Grant(role *v1alpha1.Role, grant *v1alpha1.Grant) error {

	currentGrants, err := r.getAllCurrentGrants(role)

	if err != nil {
		return err
	}

	// get desired and undesired grants by subtracting the intersections of
	// current and desired grants
	desiredGrants, undesiredGrants := SubtractGrantIntersection(grant.Objects, currentGrants)

	for _, desiredGrant := range desiredGrants {

		query, err := createGrantQuery(
			role.Spec.RoleName,
			&desiredGrant,
		)

		if err != nil {
			log.Errorf(err.Error())
			continue // continue with next grant-statement
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

	for _, undesiredGrant := range undesiredGrants {

		query, err := createRevokeQuery(
			role.Spec.RoleName,
			&undesiredGrant,
		)

		if err != nil {
			log.Errorf(err.Error())
			continue // continue with next grant-statement
		}

		_, err = r.conn.Exec(
			context.Background(),
			query,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {

				log.Errorf(
					"unable to apply revoke to role '%s', failed with code: '%s' and message: '%s'",
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

func getJoinedPrivileges(grantObject *v1alpha1.GrantObject) string {
	privileges := make([]string, len(grantObject.Privileges))

	for index, privilege := range grantObject.Privileges {
		privileges[index] = string(privilege)
	}

	return strings.Join(privileges, ", ")
}

func createGrantQuery(roleName string, grantTarget *v1alpha1.GrantObject) (string, error) {
	var query string
	switch strings.ToUpper(grantTarget.Type) {
	case "TABLE":
		query = fmt.Sprintf(
			"GRANT %s ON TABLE %s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SCHEMA":
		query = fmt.Sprintf(
			"GRANT %s ON  SCHEMA %s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "FUNCTION":
		query = fmt.Sprintf(
			"GRANT %s ON FUNCTION %s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SEQUENCE":
		query = fmt.Sprintf(
			"GRANT %s ON SEQUENCE %s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "ROLE":
		query = fmt.Sprintf(
			"GRANT %s TO %s",
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)
		if grantTarget.WithAdminOption {
			query += " WITH ADMIN OPTION"
		}

	default:
		return "", fmt.Errorf("grant type %s unknown", grantTarget.Type)
	}

	if strings.ToUpper(grantTarget.Type) != "ROLE" && grantTarget.WithGrantOption {
		query += " WITH GRANT OPTION"
	}
	return query, nil
}

func createRevokeQuery(roleName string, revokeTarget *v1alpha1.GrantObject) (string, error) {
	var query string
	switch strings.ToUpper(revokeTarget.Type) {
	case "TABLE":
		query = fmt.Sprintf(
			"REVOKE %s ON TABLE %s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SCHEMA":
		query = fmt.Sprintf(
			"REVOKE %s ON SCHEMA %s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "FUNCTION": //TODO
		query = fmt.Sprintf(
			"GRANT %s ON FUNCTION %s TO %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SEQUENCE": //TODO
		query = fmt.Sprintf(
			"GRANT %s ON SEQUENCE %s TO %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "ROLE":
		query = fmt.Sprintf(
			"REVOKE %s FROM %s",
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	default:
		return "", fmt.Errorf("grant type %s unknown", revokeTarget.Type)
	}

	return query, nil
}
