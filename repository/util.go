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

func SubstractPrivilegeConjunction(a, b []v1alpha1.Privilege) []v1alpha1.Privilege {
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

	for key, _ := range buffer {
		result = append(result, key)
	}
	return result
}

func PrivilegeSymmetricDifference(a, b []v1alpha1.Privilege) ([]v1alpha1.Privilege, []v1alpha1.Privilege) {
	var aBuffer []v1alpha1.Privilege
	var bBuffer []v1alpha1.Privilege

	aBuffer = SubstractPrivilegeConjunction(a, b)
	bBuffer = SubstractPrivilegeConjunction(b, a)

	return aBuffer, bBuffer
}

func GrantObjectGotSameTarget(a, b *v1alpha1.GrantObject) bool {
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

func GrantObjectIncludesTarget(a, b *v1alpha1.GrantObject) bool {
	if a.Type == "COLUMN" && b.Type == "TABLE" {
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
func RemoveElementFromSlice(a []v1alpha1.GrantObject, index int) {

}

// TODO umbenennung
func GetGrantSymmetricDifference(a, b []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, []v1alpha1.GrantObject) {

	for outerIndex := 0; outerIndex < len(b); outerIndex++ {
		currentGrant := &b[outerIndex]

		for innerIndex := 0; innerIndex < len(a); innerIndex++ {
			desiredGrant := &a[innerIndex]

			if GrantObjectGotSameTarget(desiredGrant, currentGrant) {
				var aBuffer []v1alpha1.Privilege
				var bBuffer []v1alpha1.Privilege

				aBuffer = SubstractPrivilegeConjunction(
					desiredGrant.Privileges,
					currentGrant.Privileges,
				)

				bBuffer = SubstractPrivilegeConjunction(
					currentGrant.Privileges,
					desiredGrant.Privileges,
				)

				desiredGrant.Privileges = aBuffer
				currentGrant.Privileges = bBuffer
			}

			if GrantObjectIncludesTarget(desiredGrant, currentGrant) {
				desiredGrant.Privileges = SubstractPrivilegeConjunction(
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
