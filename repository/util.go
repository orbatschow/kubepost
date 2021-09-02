package repository

import (
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
)

func SanitizeString(input string) string {
	var ids pgx.Identifier
	ids = append(ids, input)
	return ids.Sanitize()
}

func StringArrayToPrivilegArray(sa []string) []v1alpha1.Privilege {
	var buffer []v1alpha1.Privilege
	for _, s := range sa {
		buffer = append(buffer, v1alpha1.Privilege(s))
	}
	return buffer
}

func PrivilegeSymmetricDifference(a, b []v1alpha1.Privilege) ([]v1alpha1.Privilege, []v1alpha1.Privilege) {
	var aBuffer []v1alpha1.Privilege
	var bBuffer []v1alpha1.Privilege
	privilegeIntersectionMap := map[v1alpha1.Privilege]int{}

	for _, privilege1 := range a {

		for _, privilege2 := range b {

			if privilege1 == privilege2 {
				_, contains := privilegeIntersectionMap[privilege1]
				if !contains {
					privilegeIntersectionMap[privilege1] = 0
				}
				privilegeIntersectionMap[privilege1]++
			}
		}
	}

	for _, privilege := range a {
		_, contains := privilegeIntersectionMap[privilege]
		if !contains {
			aBuffer = append(aBuffer, privilege)
		}
	}

	for _, privilege := range b {
		_, contains := privilegeIntersectionMap[privilege]
		if !contains {
			bBuffer = append(bBuffer, privilege)
		}
	}

	return aBuffer, bBuffer
}

// TODO umbenennung
func GrantSymmetricDifference(a, b []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, []v1alpha1.GrantObject) {

	privilegeMap := getPrivilegeMap()

	for outerIndex := 0; outerIndex < len(a); outerIndex++ {
		desiredGrant := &a[outerIndex]

		// In case "ALL" is choosen as privilege, replace it with an expanded version
		for _, privilege := range desiredGrant.Privileges {
			if privilege == "ALL" {
				desiredGrant.Privileges = privilegeMap[desiredGrant.Type]
			}
		}

		for innerIndex := 0; innerIndex < len(b); innerIndex++ {
			currentGrant := &b[innerIndex]

			if desiredGrant.Identifier != currentGrant.Identifier {
				continue
			}

			if desiredGrant.Type != currentGrant.Type {
				continue
			}

			if desiredGrant.Schema != currentGrant.Schema {
				continue
			}

			desiredGrant.Privileges, currentGrant.Privileges = PrivilegeSymmetricDifference(
				desiredGrant.Privileges,
				currentGrant.Privileges,
			)

			// In case there are no privileges left in currentGrant: remove it
			if currentGrant.Privileges == nil {

				b[innerIndex] = b[len(b)-1] // Copy last element to index
				b = b[:len(b)-1]            // Truncate slice.
				innerIndex--
			}

			if desiredGrant.Privileges == nil {

				a[outerIndex] = a[len(a)-1] // Copy last element to index
				a = a[:len(a)-1]            // Truncate slice.
				outerIndex--
				break
			}

		}
	}
	return a, b
}

func getPrivilegeMap() map[string][]v1alpha1.Privilege {
	return map[string][]v1alpha1.Privilege{
		"SCHEMA":   {"USAGE", "CREATE"},
		"TABLE":    {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		"COLUMN":   {"SELECT", "UPDATE", "INSERT", "REFERENCES"},
		"FUNCTION": {"EXECUTE"},
		"SEQUENCE": {"USAGE", "SELECT", "UPDATE"},
	}
}

func getGrantQuerieMap() map[string]string {
	return map[string]string{
		"TABLE": `
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
	    select
		'COLUMN' as type,
	    table_schema as schema,
	    table_name as table,
		column_name as identifier,
	    array_agg(cast(privilege_type AS text)) as privileges,
	    is_grantable::bool as withGrantOption
		from information_schema.column_privileges
		where grantee = $1
		GROUP BY type, schema, table_name, identifier, withGrantOption`,

		"FUNCTION": `
		select
        'FUNCTION' as type,
        routine_schema as schema,
        '' as table,
        routine_name as identifier,
        array_agg(privilege_type) as priviliges,
        is_grantable::bool as withGrantOption
        from information_schema.role_routine_grants 
        WHERE grantee=$1
        GROUP BY identifier, type, schema, table, withGrantOption`,

		"SEQUENCE": `
        select
        'FUNCTION' as type,
        squence_schema as schema,
        '' as table_name,
        squence_name as identifier,
        array_agg(privilege_type) as priviliges,
        is_grantable::bool as withGrantOption
        from information_schema.role_routine_grants 
        WHERE grantee=$1
        GROUP BY identifier, type, schema, table_name, withGrantOption`,
	}
}

func getRegexQueryByType() map[string]string {
	return map[string]string{
		"SCHEMA":   "select nspname from pg_namespace where nspname ~ '$1'",
		"TABLE":    "select tablename from pg_tables where schemaname ~ '$1' and tablename ~ '$2';",
		"COLUMN":   "select column_name from information_schema.columns where table_schema ~ '$1' and table_name ~ '$2' and column_name ~ '$3'",
		"ROLE":     "select rolname from pg_authid where rolname ~ '1$'",
		"FUNCTION": "select routine_name from  information_schema.routines where routine_schema ~ '$1' and routine_name ~ '$2'",
		"SEQUENCE": "",
	}
}
