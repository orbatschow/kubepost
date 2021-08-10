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

// This function will compare two grantObjects and will remove priviliges if
// both would grant them. Returning the number of found intersections.
func SubtractPrivilegeIntersection(a, b *v1alpha1.GrantObject) int {
    var aBuffer []v1alpha1.Privilege
    var bBuffer []v1alpha1.Privilege
    counter := 0

    if a.Identifier != b.Identifier {
        return 0
    }

    if a.Type != b.Type {
        return 0
    }

    if a.Schema != b.Schema {
        return 0
    }

    privilegeIntersectionMap := map[v1alpha1.Privilege]int{}

    for _, privilege1 := range a.Privileges {

        for _, privilege2 := range b.Privileges {

            if privilege1 == privilege2 {
                _, contains := privilegeIntersectionMap[privilege1]
                if !contains {
                    privilegeIntersectionMap[privilege1] = 0
                }
                privilegeIntersectionMap[privilege1]++
                counter++
            }
        }
    }

    for _, privilege := range a.Privileges {
        _, contains := privilegeIntersectionMap[privilege]
        if !contains {
            aBuffer = append(aBuffer, privilege)
        }
    }

    a.Privileges = aBuffer

    for _, privilege := range b.Privileges {
        _, contains := privilegeIntersectionMap[privilege]
        if !contains {
            bBuffer = append(bBuffer, privilege)
        }
    }

    b.Privileges = bBuffer

    return counter
}

