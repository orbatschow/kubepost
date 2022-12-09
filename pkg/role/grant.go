package role

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/instance"
	"github.com/orbatschow/kubepost/pkg/postgres"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

func (r *Repository) ReconcileGrants(ctx context.Context, ctrlClient client.Client) error {

	defaultDatabase := r.instance.Spec.Database

	databases, err := r.GetDatabaseNames(ctx)
	if err != nil {
		return err
	}

	log.FromContext(ctx).Info("computed databases for grant", "databases", databases)

	for _, database := range databases {

		// we have to connect to all databases to grant/revoke the privileges
		// therefore we will modify the connection for each database
		r.instance.Spec.Database = database
		r.conn, err = instance.GetConnection(ctx, ctrlClient, r.instance)
		if err != nil {
			return err
		}

		var grantObjects []v1alpha1.GrantObject

		for _, grant := range r.role.Spec.Grants {
			if grant.Database == database {
				grantObjects = append(grantObjects, grant.Objects...)
			}
		}

		// regex
		grantObjects, err = r.regexExpandGrantObjects(ctx, grantObjects)
		if err != nil {
			return err
		}

		currentGrants, err := r.GetCurrentGrants(ctx)
		if err != nil {
			return err
		}

		// get desired and undesired grants by subtracting the intersections of
		// current and desired grants
		desiredGrants, undesiredGrants := getGrantSymmetricDifference(
			grantObjects,
			currentGrants,
		)

		err = r.Grant(ctx, desiredGrants)
		if err != nil {
			return err
		}

		err = r.Revoke(ctx, undesiredGrants)
		if err != nil {
			return err
		}
	}

	// reset the database to the previous database, that was configured within the CRD
	r.instance.Spec.Database = defaultDatabase
	r.conn, err = instance.GetConnection(ctx, ctrlClient, r.instance)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetCurrentGrants(ctx context.Context) ([]v1alpha1.GrantObject, error) {
	var currentGrants []v1alpha1.GrantObject
	var buffer []v1alpha1.GrantObject

	buffer, err := r.getGrantsByType(ctx, postgres.TABLE)
	currentGrants = append(currentGrants, buffer...)
	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
	}

	buffer, err = r.getGrantsByType(ctx, postgres.SCHEMA)
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
	}

	buffer, err = r.getGrantsByType(ctx, postgres.COLUMN)
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
	}

	buffer, err = r.getGrantsByType(ctx, postgres.FUNCTION)
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
	}

	buffer, err = r.getGrantsByType(ctx, postgres.SEQUENCE)
	currentGrants = append(currentGrants, buffer...)

	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
	}

	return currentGrants, nil
}

func (r *Repository) Grant(ctx context.Context, desiredGrants []v1alpha1.GrantObject) error {

	log.FromContext(ctx).Info("reconciling desired grants", "grants", desiredGrants)

	for _, desiredGrant := range desiredGrants {

		query, err := r.createGrantQuery(
			ctx,
			&desiredGrant,
		)

		if err != nil {
			log.FromContext(ctx).Error(err, "could not create grant query")
			continue
		}

		_, err = r.conn.Exec(
			ctx,
			query,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return RepositoryError{
					Role:                 r.role.ObjectMeta.Name,
					Instance:             r.instance.ObjectMeta.Name,
					Namespace:            r.role.ObjectMeta.Namespace,
					Message:              "unable to apply revoke query",
					PostgresErrorCode:    pgErr.Code,
					PostgresErrorMessage: pgErr.Message,
				}
			}
			return RepositoryError{
				Role:      r.role.ObjectMeta.Name,
				Instance:  r.instance.ObjectMeta.Name,
				Namespace: r.role.ObjectMeta.Namespace,
				Message:   fmt.Sprintf("unable to apply Grant query: '%s'", err.Error()),
			}
		}
	}
	return nil
}

func (r *Repository) Revoke(ctx context.Context, undesiredGrants []v1alpha1.GrantObject) error {

	log.FromContext(ctx).Info("reconciling undesired grants", "grants", undesiredGrants)

	for _, undesiredGrant := range undesiredGrants {

		query, err := r.createRevokeQuery(
			ctx,
			&undesiredGrant,
		)

		if err != nil {
			log.FromContext(ctx).Error(err, "could not create revoke query")
			continue // continue with the next grant statement
		}

		_, err = r.conn.Exec(
			ctx,
			query,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return RepositoryError{
					Role:                 r.role.ObjectMeta.Name,
					Instance:             r.instance.ObjectMeta.Name,
					Namespace:            r.role.ObjectMeta.Namespace,
					Message:              "unable to apply Revoke query",
					PostgresErrorCode:    pgErr.Code,
					PostgresErrorMessage: pgErr.Message,
				}
			}
			return RepositoryError{
				Role:      r.role.ObjectMeta.Name,
				Instance:  r.instance.ObjectMeta.Name,
				Namespace: r.role.ObjectMeta.Namespace,
				Message:   fmt.Sprintf("unable to apply Revoke query: '%s'", err.Error()),
			}
		}
	}

	return nil
}

func (r *Repository) regexExpandGrantObjects(ctx context.Context, grantObjects []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, error) {

	var grantObjectsExpanded []v1alpha1.GrantObject
	privileges := map[string][]v1alpha1.Privilege{
		postgres.SCHEMA:   {"USAGE", "CREATE"},
		postgres.TABLE:    {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		postgres.VIEW:     {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		postgres.COLUMN:   {"SELECT", "UPDATE", "INSERT", "REFERENCES"},
		postgres.FUNCTION: {"EXECUTE"},
		postgres.SEQUENCE: {"USAGE", "SELECT", "UPDATE"},
	}

	// In case "ALL" is chosen, replace it with an expanded version
	for index, grant := range grantObjects {
		for _, privilege := range grant.Privileges {
			if privilege == "ALL" {
				grantObjects[index].Privileges = privileges[grant.Type]
			}
		}
	}

	for _, grantObject := range grantObjects {

		var err error
		var rows pgx.Rows

		switch grantObject.Type {
		case postgres.SCHEMA:
			rows, err = r.conn.Query(
				ctx,
				`select nspname from pg_namespace where nspname ~ $1`,
				"^"+grantObject.Identifier+"$",
			)
		case postgres.VIEW:
			rows, err = r.conn.Query(
				ctx,
				`select
					viewname
				from pg_views
				where schemaname ~ $1
				and viewname ~ $2`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Identifier+"$",
			)
		case postgres.TABLE:
			rows, err = r.conn.Query(
				ctx,
				`select
					tablename
				from pg_tables
				where schemaname ~ $1
				and tablename ~ $2`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Identifier+"$",
			)
		case postgres.COLUMN:
			rows, err = r.conn.Query(
				ctx,
				`select
					column_name
				from information_schema.columns
				where table_schema ~ $1
				and table_name ~ $2
				and column_name ~ $3`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Table+"$",
				"^"+grantObject.Identifier+"$",
			)
		case postgres.FUNCTION:
			rows, err = r.conn.Query(
				ctx,
				`select
					routine_name
				from information_schema.routines
				where routine_schema ~ $1
				and routine_name ~ $2`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Identifier+"$",
			)
		case postgres.SEQUENCE:
			rows, err = r.conn.Query(
				ctx,
				`select
					sequence_name
				from information_schema.sequences
				where sequence_schema ~ $1
				and sequence_name ~ $2`,
				"^"+grantObject.Schema+"$",
				"^"+grantObject.Identifier+"$",
			)
		}

		if err != nil {
			return nil, err
		}

		var entries []string
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

			if grantObject.Type == postgres.VIEW {
				grantObject.Type = postgres.TABLE
			}

			grantObjectsExpanded = append(grantObjectsExpanded, grantObject)
		}
	}

	return grantObjectsExpanded, nil
}

func (r *Repository) getGrantsByType(ctx context.Context, grantType string) ([]v1alpha1.GrantObject, error) {

	computePrivileges := func(privileges []string) []v1alpha1.Privilege {
		var buffer []v1alpha1.Privilege
		for _, privilege := range privileges {
			buffer = append(buffer, v1alpha1.Privilege(privilege))
		}
		return buffer
	}

	grantQueries := getGrantQueries()
	var grants []v1alpha1.GrantObject

	rows, err := r.conn.Query(
		ctx,
		grantQueries[grantType],
		r.role.ObjectMeta.Name,
	)

	if err != nil {
		return nil, RepositoryError{
			Role:      r.role.ObjectMeta.Name,
			Instance:  r.instance.ObjectMeta.Name,
			Namespace: r.role.ObjectMeta.Namespace,
			Message:   err.Error(),
		}
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
			return nil, RepositoryError{
				Role:      r.role.ObjectMeta.Name,
				Instance:  r.instance.ObjectMeta.Name,
				Namespace: r.role.ObjectMeta.Namespace,
				Message:   err.Error(),
			}
		}

		grant.Privileges = computePrivileges(privileges)
		grants = append(grants, grant)
	}

	return grants, nil
}

func getGrantSymmetricDifference(a, b []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, []v1alpha1.GrantObject) {

	for outerIndex := 0; outerIndex < len(b); outerIndex++ {
		currentGrant := &b[outerIndex]

		for innerIndex := 0; innerIndex < len(a); innerIndex++ {
			desiredGrant := &a[innerIndex]

			if grantObjectGotSameTarget(desiredGrant, currentGrant) {
				var aBuffer []v1alpha1.Privilege
				var bBuffer []v1alpha1.Privilege

				aBuffer = subtractPrivilegeConjunction(
					desiredGrant.Privileges,
					currentGrant.Privileges,
				)

				bBuffer = subtractPrivilegeConjunction(
					currentGrant.Privileges,
					desiredGrant.Privileges,
				)

				desiredGrant.Privileges = aBuffer
				currentGrant.Privileges = bBuffer
			}

			if grantObjectIncludesTarget(desiredGrant, currentGrant) {
				desiredGrant.Privileges = subtractPrivilegeConjunction(
					desiredGrant.Privileges,
					currentGrant.Privileges,
				)
			}

			// In case there are no privileges left in currentGrant: remove it
			if len(desiredGrant.Privileges) == 0 {
				a[innerIndex] = a[len(a)-1] // Copy last element to index
				a = a[:len(a)-1]            // Truncate slice.
				innerIndex--
			}

			if len(currentGrant.Privileges) == 0 {
				b[outerIndex] = b[len(b)-1] // Copy last element to index
				b = b[:len(b)-1]            // Truncate slice.
				outerIndex--
				break
			}
		}
	}

	return a, b
}

func grantObjectGotSameTarget(a, b *v1alpha1.GrantObject) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Schema != b.Schema {
		return false
	}
	if a.Type == "COLUMN" && a.Table != b.Table {
		return false
	}
	if a.Identifier != b.Identifier {
		return false
	}
	if a.WithGrantOption != b.WithGrantOption {
		return false
	}
	return true
}

func subtractPrivilegeConjunction(a, b []v1alpha1.Privilege) []v1alpha1.Privilege {
	buffer := map[v1alpha1.Privilege]int{}
	var result []v1alpha1.Privilege
	var found bool

	for _, privilegeA := range a {
		found = false
		for _, privilegeB := range b {
			if privilegeA == privilegeB {
				found = true
				break
			}
		}
		if !found {
			buffer[privilegeA] = 1
		}
	}

	for key := range buffer {
		result = append(result, key)
	}
	return result
}

// In case the GrantObjects a and b don't have the same target, there is still
// the possibility for b (table) including a (columns). This function checks for
// this case.
func grantObjectIncludesTarget(a, b *v1alpha1.GrantObject) bool {
	if a.Type == postgres.COLUMN && b.Type == postgres.TABLE {
		if a.Schema != b.Schema {
			return false
		}
		if a.Table != b.Identifier {
			return false
		}
		if a.WithGrantOption != b.WithGrantOption {
			return false
		}
		return true
	}
	return false
}

func getGrantQueries() map[string]string {
	return map[string]string{
		postgres.TABLE: `
        select
		'TABLE' as type,
		table_schema as schema,
		'' as table,
        table_name as identifier,
        array_agg(cast(privilege_type AS text)) as privileges,
        is_grantable::bool as withGrantOption
        from information_schema.role_table_grants
        WHERE grantee=$1
        GROUP BY identifier, schema, withGrantOption`,

		"SCHEMA": `
        SELECT
        type,
		schema,
		'' as table,
        identifier,
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
        GROUP BY identifier, type, schema, withGrantOption`,

		"COLUMN": `
	    Select
			t1.type,
			t1.schema,
			t1.table,
			t1.identifier,
			array_agg(t1.privileges),
			t1.withgrantoption
			from
		(select
		        'COLUMN' as type,
		        table_schema as schema,
		        table_name as table,
		        column_name as identifier,
		        cast(privilege_type AS text) as privileges,
		        is_grantable::bool as withGrantOption
		        from information_schema.column_privileges
		        where grantee=$1
		except
		select
		        'COLUMN' as type,
		        t.table_schema as schema,
		        t.table_name as table,
		        c.column_name as identifier,
		        cast(t.privilege_type AS text) as privileges,
		        t.is_grantable::bool as withGrantOption
		        from information_schema.role_table_grants t
		        LEFT JOIN information_schema.columns c
		        on (t.table_name = c.table_name
		        AND t.table_schema = c.table_schema)
		        WHERE grantee=$1)
		as t1
		group by t1.type, t1.schema, t1.table, t1.identifier, t1.withgrantoption`,

		"FUNCTION": `
		select
        'FUNCTION' as type,
        routine_schema as schema,
        '' as table_name,
        routine_name as identifier,
		array_agg('EXECUTE'::varchar) as privileges,
        is_grantable::bool as withGrantOption
        from information_schema.role_routine_grants
        WHERE grantee=$1
        GROUP BY identifier, type, schema, table_name, withGrantOption`,

		"SEQUENCE": `
        select
			sq.type,
			sq.schema,
			'' as table_name,
			sq.identifier,
			array_agg(sq.privileges) as privileges,
			sq.withGrantOption
		from (
			select
			    'SEQUENCE' as type,
			    nspname as schema,
			    relname as identifier,
			    (aclexplode(relacl)).privilege_type as privileges,
			    (aclexplode(relacl)).is_grantable as withGrantOption,
			    (aclexplode(relacl)).grantee as grantee
			from pg_class cl
			join pg_namespace nsp on (cl.relnamespace = nsp.oid)
			join pg_sequence sq on (cl.oid = sq.seqrelid)
			where relkind='S'
		) as sq
		join pg_authid au on (sq.grantee = au.oid)
		where rolname=$1
		GROUP BY identifier, type, schema, table_name, withGrantOption`,
	}
}

func (r *Repository) createGrantQuery(ctx context.Context, grantTarget *v1alpha1.GrantObject) (string, error) {
	var query string
	switch strings.ToUpper(grantTarget.Type) {
	case postgres.COLUMN:
		query = fmt.Sprintf(
			"GRANT %s (%s) ON TABLE %s.%s TO %s",
			getJoinedPrivileges(ctx, grantTarget),
			postgres.SanitizeString(grantTarget.Identifier),
			postgres.SanitizeString(grantTarget.Schema),
			postgres.SanitizeString(grantTarget.Table),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)
	case postgres.TABLE:
		query = fmt.Sprintf(
			"GRANT %s ON %s.%s TO %s",
			getJoinedPrivileges(ctx, grantTarget),
			postgres.SanitizeString(grantTarget.Schema),
			postgres.SanitizeString(grantTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.SCHEMA:
		query = fmt.Sprintf(
			"GRANT %s ON  SCHEMA %s TO %s",
			getJoinedPrivileges(ctx, grantTarget),
			postgres.SanitizeString(grantTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.FUNCTION:
		query = fmt.Sprintf(
			"GRANT %s ON FUNCTION %s.%s TO %s",
			getJoinedPrivileges(ctx, grantTarget),
			postgres.SanitizeString(grantTarget.Schema),
			postgres.SanitizeString(grantTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.SEQUENCE:
		query = fmt.Sprintf(
			"GRANT %s ON SEQUENCE %s.%s TO %s",
			getJoinedPrivileges(ctx, grantTarget),
			postgres.SanitizeString(grantTarget.Schema),
			postgres.SanitizeString(grantTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	default:
		return "", fmt.Errorf("grant type %s unknown", grantTarget.Type)
	}

	if strings.ToUpper(grantTarget.Type) != "ROLE" && grantTarget.WithGrantOption {
		query += " WITH GRANT OPTION"
	}

	log.FromContext(ctx).Info("computed grant query",
		"query", query)
	return query, nil
}

func (r *Repository) createRevokeQuery(ctx context.Context, revokeTarget *v1alpha1.GrantObject) (string, error) {
	var query string
	switch strings.ToUpper(revokeTarget.Type) {
	case postgres.COLUMN:
		query = fmt.Sprintf(
			"REVOKE %s (%s) ON TABLE %s.%s FROM %s",
			getJoinedPrivileges(ctx, revokeTarget),
			postgres.SanitizeString(revokeTarget.Identifier),
			postgres.SanitizeString(revokeTarget.Schema),
			postgres.SanitizeString(revokeTarget.Table),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)
	case postgres.TABLE:
		query = fmt.Sprintf(
			"REVOKE %s ON %s.%s FROM %s",
			getJoinedPrivileges(ctx, revokeTarget),
			postgres.SanitizeString(revokeTarget.Schema),
			postgres.SanitizeString(revokeTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.SCHEMA:
		query = fmt.Sprintf(
			"REVOKE %s ON SCHEMA %s FROM %s",
			getJoinedPrivileges(ctx, revokeTarget),
			postgres.SanitizeString(revokeTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.FUNCTION:
		query = fmt.Sprintf(
			"REVOKE %s ON FUNCTION %s.%s FROM %s",
			getJoinedPrivileges(ctx, revokeTarget),
			postgres.SanitizeString(revokeTarget.Schema),
			postgres.SanitizeString(revokeTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	case postgres.SEQUENCE:
		query = fmt.Sprintf(
			"REVOKE %s ON SEQUENCE %s.%s FROM %s",
			getJoinedPrivileges(ctx, revokeTarget),
			postgres.SanitizeString(revokeTarget.Schema),
			postgres.SanitizeString(revokeTarget.Identifier),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		)

	default:
		return "", fmt.Errorf("Revoke type %s unknown", revokeTarget.Type)
	}

	log.FromContext(ctx).Info("computed revoke query", "query", query)
	return query, nil
}

func getJoinedPrivileges(ctx context.Context, grantObject *v1alpha1.GrantObject) string {
	privileges := make([]string, len(grantObject.Privileges))

	for index, privilege := range grantObject.Privileges {
		privileges[index] = string(privilege)
	}

	log.FromContext(ctx).Info("computed privileges", "privileges", privileges)

	return strings.Join(privileges, ", ")
}

// TODO: move to database package
func (r *Repository) GetDatabaseNames(ctx context.Context) ([]string, error) {
	var databaseNames []string

	rows, err := r.conn.Query(
		ctx,
		"select datname from pg_database where datistemplate = 'f'",
	)
	if err != nil {
		return nil, err
	}

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
