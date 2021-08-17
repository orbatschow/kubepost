package repository

import (
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/thoas/go-funk"
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

func SubtractPrivilegeIntersection(a, b *v1alpha1.GrantObject) int {
	var aBuffer []v1alpha1.Privilege
	var bBuffer []v1alpha1.Privilege
	counter := 0

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

func SubtractGrantIntersection(desiredGrants, currentGrants []v1alpha1.GrantObject) ([]v1alpha1.GrantObject, []v1alpha1.GrantObject) {

	privilegeMap := getPrivilegeMap()
	for outerIndex := 0; outerIndex < len(desiredGrants); outerIndex++ {
		desiredGrant := &desiredGrants[outerIndex]

		// In case "ALL" is choosen as privilege, replace it with an expanded version
		for _, privilege := range desiredGrant.Privileges {
			if privilege == "ALL" {
				desiredGrant.Privileges = privilegeMap[desiredGrant.Type]
			}
		}

		for innerIndex := 0; innerIndex < len(currentGrants); innerIndex++ {
			currentGrant := &currentGrants[innerIndex]

			if desiredGrant.Identifier != currentGrant.Identifier {
				innerIndex++
				continue
			}

			if desiredGrant.Type != currentGrant.Type {
				innerIndex++
				continue
			}

			if desiredGrant.Schema != currentGrant.Schema {
				innerIndex++
				continue
			}

			if desiredGrant.Type != "ROLE" {
				// This function will subtract all intersections between both arrays
				// of privileges in both grantObjects
				SubtractPrivilegeIntersection(desiredGrant, currentGrant)
			}

			// In case there are no privileges left in currentGrant: remove it
			if currentGrant.Privileges == nil {

				currentGrants[innerIndex] = currentGrants[len(currentGrants)-1] // Copy last element to index
				currentGrants = currentGrants[:len(currentGrants)-1]            // Truncate slice.
				innerIndex--
			}

			if desiredGrant.Privileges == nil {

				desiredGrants[outerIndex] = desiredGrants[len(desiredGrants)-1] // Copy last element to index
				desiredGrants = desiredGrants[:len(desiredGrants)-1]            // Truncate slice.
				outerIndex--
			}

		}
	}
	return desiredGrants, currentGrants
}

func expandIdentifier(grantObjects []v1alpha1.GrantObject) []v1alpha1.GrantObject {

	var buffer []v1alpha1.GrantObject

	for _, grantObject := range grantObjects {

		expandedIdentifiers := []string{"query"}

		for _, expandedIdentifier := range expandedIdentifiers {

			// check if buffer already contains the current identifier
			match := funk.IndexOf(buffer, func(b v1alpha1.GrantObject) bool {
				return b.Identifier == expandedIdentifier && b.Type == grantObject.Type
			})

			if match != -1 {
				// combine privileges of existing identifier, with current identifier
				//combinedPrivileges := append(buffer[match].Privileges, grantObject.Privileges...)
				//buffer[match].Privileges = funk.UniqString(combinedPrivileges)
				continue
			}

			// append new identifier to buffer
			buffer = append(buffer, v1alpha1.GrantObject{
				Identifier:      expandedIdentifier,
				Type:            grantObject.Type,
				Privileges:      grantObject.Privileges,
				WithGrantOption: grantObject.WithGrantOption,
				WithAdminOption: grantObject.WithAdminOption,
			})
		}

	}

	return buffer

}

func getPrivilegeMap() map[string][]v1alpha1.Privilege {
	return map[string][]v1alpha1.Privilege{
		"TABLE":  {"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		"SCHEMA": {"USAGE", "CREATE"},
	}
}
