package role

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/postgres"
)

func (r *Repository) ReconcileGroups(ctx context.Context) error {

	currentGroups, err := r.GetGroups(ctx)
	if err != nil {
		return err
	}

	desiredGroups := r.role.Spec.Groups

	desiredGroups, currentGroups = getGroupGrantObjectSymmetricDifference(
		desiredGroups,
		currentGroups,
	)

	for _, undesiredGroup := range currentGroups {
		err := r.RemoveGroup(ctx, &undesiredGroup)
		if err != nil {
			return err
		}
	}

	for _, desiredGroup := range desiredGroups {
		err := r.AddGroup(ctx, &desiredGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) GetGroups(ctx context.Context) ([]v1alpha1.GroupGrantObject, error) {
	var groups []v1alpha1.GroupGrantObject

	rows, err := r.conn.Query(
		ctx,
		`SELECT
        u.rolname,
        admin_option as withAdminOption
        FROM pg_catalog.pg_auth_members m
        JOIN pg_catalog.pg_authid u on (m.roleid = u.oid)
        WHERE m.member = (select oid from pg_authid where rolname=$1)`,
		r.role.ObjectMeta.Name,
	)
	if err != nil {
		return nil, err
	}

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

func (r *Repository) AddGroup(ctx context.Context, group *v1alpha1.GroupGrantObject) error {

	query := fmt.Sprintf(
		"GRANT %s TO %s",
		postgres.SanitizeString(group.Name),
		postgres.SanitizeString(r.role.ObjectMeta.Name),
	)

	if group.WithAdminOption {
		query += " WITH ADMIN OPTION"
	}

	_, err := r.conn.Exec(
		ctx,
		query,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return RepositoryError{
			Role:                 r.role.ObjectMeta.Name,
			Instance:             r.instance.ObjectMeta.Name,
			Namespace:            r.role.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}

	return nil
}

func (r *Repository) RemoveGroup(ctx context.Context, group *v1alpha1.GroupGrantObject) error {

	_, err := r.conn.Exec(
		ctx,
		fmt.Sprintf(
			"REVOKE %s FROM %s",
			postgres.SanitizeString(group.Name),
			postgres.SanitizeString(r.role.ObjectMeta.Name),
		),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		errorCode := ""
		errorMessage := ""
		if errors.As(err, &pgErr) {
			errorCode = pgErr.Code
			errorMessage = pgErr.Message
		}

		return RepositoryError{
			Role:                 r.role.ObjectMeta.Name,
			Instance:             r.instance.ObjectMeta.Name,
			Namespace:            r.role.ObjectMeta.Namespace,
			Message:              err.Error(),
			PostgresErrorCode:    errorCode,
			PostgresErrorMessage: errorMessage,
		}
	}
	return nil
}

func getGroupGrantObjectSymmetricDifference(desiredGroups, currentGroups []v1alpha1.GroupGrantObject) ([]v1alpha1.GroupGrantObject, []v1alpha1.GroupGrantObject) {
	for outerIndex := 0; outerIndex < len(desiredGroups); outerIndex++ {
		desiredGroup := &desiredGroups[outerIndex]

		for innerIndex := 0; innerIndex < len(currentGroups); innerIndex++ {
			currentGroup := &currentGroups[innerIndex]

			if desiredGroup.Name != currentGroup.Name {
				continue
			}

			if desiredGroup.WithAdminOption != currentGroup.WithAdminOption {
				continue
			}

			currentGroups[innerIndex] = currentGroups[len(currentGroups)-1] // Copy last element to index
			currentGroups = currentGroups[:len(currentGroups)-1]            // Truncate slice.
			innerIndex--

			desiredGroups[outerIndex] = desiredGroups[len(desiredGroups)-1] // Copy last element to index
			desiredGroups = desiredGroups[:len(desiredGroups)-1]            // Truncate slice.
			outerIndex--

			break
		}
	}

	return desiredGroups, currentGroups
}
