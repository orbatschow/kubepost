package utils

import (
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
)

func SanitizeString(input string) string {
	var ids pgx.Identifier
	ids = append(ids, input)
	return ids.Sanitize()
}

func getPrivilegeMap() map[string][]v1alpha1.Privilege {
	return map[string][]v1alpha1.Privilege{
		"SCHEMA":   {"USAGE", "CREATE"},
		"TABLE":    {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		"VIEW":     {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
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
		array_agg('EXECUTE'::varchar) as priviliges,
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
			array_agg(sq.priviliges) as priviliges,
			sq.withGrantOption
		from (
			select
			    'SEQUENCE' as type,
			    nspname as schema,
			    relname as identifier,
			    (aclexplode(relacl)).privilege_type as priviliges,
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
