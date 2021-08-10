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

