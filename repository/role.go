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

func (r *roleRepository) AddGroup(role *v1alpha1.Role, group *v1alpha1.GroupGrantObject) error {

	query := fmt.Sprintf(
		"GRANT %s TO %s",
		SanitizeString(group.Name),
		SanitizeString(role.Name),
	)

	if group.WithAdminOption {
		query += " WITH ADMIN OPTION"
	}

	_, err := r.conn.Exec(
		context.Background(),
		query,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Errorf(
			"unable to add group for role '%s', failed with code: '%s' and message: '%s'",
			role.Name,
			pgErr.Code,
			pgErr.Message,
		)
		return err
	}

	return nil
}

func (r *roleRepository) RemoveGroup(role *v1alpha1.Role, group *v1alpha1.GroupGrantObject) error {

	_, err := r.conn.Exec(
		context.Background(),
		fmt.Sprintf(
			"REVOKE %s FROM %s",
			SanitizeString(group.Name),
			SanitizeString(role.Name),
		),
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Errorf(
			"unable to remove group for role '%s', failed with code: '%s' and message: '%s'",
			role.Name,
			pgErr.Code,
			pgErr.Message,
		)
		return err
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

func (r *roleRepository) GetDatabaseNames() ([]string, error) {
	databaseNames := []string{}

	rows, err := r.conn.Query(
		context.Background(),
		"select datname from pg_database where datistemplate = 'f'",
	)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)

		if err != nil {
			return nil, err
		}

		databaseNames = append(databaseNames, name)
	}
	return databaseNames, nil
}

func (r *roleRepository) GetGroups(role *v1alpha1.Role) ([]v1alpha1.GroupGrantObject, error) {
	groups := []v1alpha1.GroupGrantObject{}

	rows, err := r.conn.Query(
		context.Background(),
		`SELECT 
        u.rolname,
        admin_option as withAdminOption
        FROM pg_catalog.pg_auth_members m
        JOIN pg_catalog.pg_authid u on (m.roleid = u.oid)
        WHERE m.member = (select oid from pg_authid where rolname=$1)`,
		role.Name,
	)

	for rows.Next() {
		var group v1alpha1.GroupGrantObject
		err = rows.Scan(&group.Name, &group.WithAdminOption)

		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}
	return groups, nil
}

func (r *roleRepository) RegexExpandGrantObjects(grantObjects []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, error) {
	grantObjectsExpanded := []v1alpha1.GrantObject{}
	var err error
	var rows pgx.Rows

	for _, grantObject := range grantObjects {
		switch grantObject.Type {
		case "SCHEMA":
			rows, err = r.conn.Query(
				context.Background(),
				`select nspname from pg_namespace where nspname ~ $1`,
				"^"+grantObject.Identifier+"$",
			)
		case "TABLE":
			rows, err = r.conn.Query(
				context.Background(),
				`select tablename from pg_tables where schemaname ~ $1 and tablename ~ $2`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Identifier+"$",
			)
		}
		grantObjectsExpanded = append(grantObjectsExpanded)
		entries := []string{}
		for rows.Next() {
			var entry string
			err = rows.Scan(&entry)

			if err != nil {
				return nil, err
			}

			entries = append(entries, entry)
		}

		for _, entry := range entries {
			grantObject.Identifier = entry
			grantObjectsExpanded = append(grantObjectsExpanded, grantObject)
		}
	}

	return grantObjectsExpanded, nil
}

func (r *roleRepository) getGrantsByType(role *v1alpha1.Role, grantType string) ([]v1alpha1.GrantObject, error) {

	getCurrentGrantsQuerieMap := getGrantQuerieMap()
	var grants []v1alpha1.GrantObject

	rows, err := r.conn.Query(
		context.Background(),
		getCurrentGrantsQuerieMap[grantType],
		role.Spec.RoleName,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var privileges []string
		var grant v1alpha1.GrantObject
		err = rows.Scan(
			&grant.Type,
			&grant.Schema,
			&grant.Table,
			&grant.Identifier,
			&privileges,
			&grant.WithGrantOption,
		)
		if err != nil {
			return nil, err
		}

		grant.Privileges = StringArrayToPrivilegArray(privileges)
		grants = append(grants, grant)
	}

	return grants, nil
}

func (r *roleRepository) GetCurrentGrants(role *v1alpha1.Role) ([]v1alpha1.GrantObject, error) {
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

	return currentGrants, err
}

func (r *roleRepository) Grant(role *v1alpha1.Role, desiredGrants []v1alpha1.GrantObject) error {

	for _, desiredGrant := range desiredGrants {

		query, err := createGrantQuery(
			role.Spec.RoleName,
			&desiredGrant,
		)

		if err != nil {
			log.Errorf(err.Error())
			continue
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

func (r *roleRepository) Revoke(role *v1alpha1.Role, undesiredGrants []v1alpha1.GrantObject) error {

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
	case "COLUMN":
		query = fmt.Sprintf(
			"GRANT %s (%s) ON TABLE %s.%s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(grantTarget.Schema),
			SanitizeString(grantTarget.Table),
			SanitizeString(roleName),
		)
	case "TABLE":
		query = fmt.Sprintf(
			"GRANT %s ON TABLE %s.%s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Schema),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SCHEMA":
		query = fmt.Sprintf(
			"GRANT %s ON  SCHEMA %s.%s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Schema),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

	case "FUNCTION":
		query = fmt.Sprintf(
			"GRANT %s ON FUNCTION %s.%s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Schema),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)
		log.Info(query)

	case "SEQUENCE":
		query = fmt.Sprintf(
			"GRANT %s ON SEQUENCE %s.%s TO %s",
			getJoinedPrivileges(grantTarget),
			SanitizeString(grantTarget.Schema),
			SanitizeString(grantTarget.Identifier),
			SanitizeString(roleName),
		)

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
	case "COLUMN":
		query = fmt.Sprintf(
			"REVOKE %s (%s) ON TABLE %s.%s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(revokeTarget.Schema),
			SanitizeString(revokeTarget.Table),
			SanitizeString(roleName),
		)
	case "TABLE":
		query = fmt.Sprintf(
			"REVOKE %s ON TABLE %s.%s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Schema),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SCHEMA":
		query = fmt.Sprintf(
			"REVOKE %s ON SCHEMA %s.%s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Schema),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "FUNCTION":
		query = fmt.Sprintf(
			"REVOKE %s ON FUNCTION %s.%s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Schema),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	case "SEQUENCE":
		query = fmt.Sprintf(
			"REVOKE %s ON SEQUENCE %s.%s FROM %s",
			getJoinedPrivileges(revokeTarget),
			SanitizeString(revokeTarget.Schema),
			SanitizeString(revokeTarget.Identifier),
			SanitizeString(roleName),
		)

	default:
		return "", fmt.Errorf("revoke type %s unknown", revokeTarget.Type)
	}

	return query, nil
}
