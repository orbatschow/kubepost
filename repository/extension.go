package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Instance v1alpha1.Instance

type extensionRepository struct {
	conn *pgx.Conn
}

func NewExtensionRepository(conn *pgx.Conn) extensionRepository {
	return extensionRepository{
		conn: conn,
	}
}

func (r *extensionRepository) Reconcile(desiredExtensions []v1alpha1.Extension, instance *v1alpha1.Instance) error {

	var existingExtensions []*v1alpha1.Extension

	err := pgxscan.Select(
		context.Background(),
		r.conn,
		&existingExtensions,
		"SELECT extname AS name, extversion AS version FROM pg_extension",
	)

	if err != nil {
		return err
	}

	for _, desiredExtension := range desiredExtensions {
		// check if desired extension already exists
		var exists bool
		for _, existingExtension := range existingExtensions {
			if desiredExtension.Name == existingExtension.Name {
				exists = true
			}
		}

		// create desired extension if it does not already exist, otherwise update
		if exists != true {
			err = createExtension(r.conn, &desiredExtension)
			if err != nil {
				return err
			}
		} else {
			err = updateExtension(r.conn, &desiredExtension)
			if err != nil {
				return err
			}
		}

	}

	// check if existing extension is desired
	for _, existingExtension := range existingExtensions {
		var desired bool

		for _, desiredExtension := range desiredExtensions {
			if existingExtension.Name == desiredExtension.Name {
				desired = true
			}
		}

		// delete existing extension if it is not desired
		if desired != true {

			// check if existingExtension is dependencie of other extension
			var dependencies []string
			err, dependencies = r.GetDependencies(existingExtension)

			if err != nil {
				return err
			}

			if len(dependencies) > 0 {
				log.Infof(
					"extension '%s' in database '%s' in namespace '%s' is an dependency for extensions:%s, skipping deletion",
					existingExtension.Name,
					instance.Spec.Database,
					instance.Namespace,
					dependencies,
				)
			} else {
				err = deleteExtension(r.conn, existingExtension)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createExtension(conn *pgx.Conn, extension *v1alpha1.Extension) error {
	if extension.Version == "latest" || extension.Version == "" {
		_, err := conn.Exec(
			context.Background(),
			fmt.Sprintf("CREATE EXTENSION %s CASCADE", SanitizeString(extension.Name)),
		)

		return err
	} else {
		_, err := conn.Exec(
			context.Background(),
			fmt.Sprintf(
				"CREATE EXTENSION %s WITH VERSION %s CASCADE",
				SanitizeString(extension.Name),
				SanitizeString(extension.Version),
			),
		)

		return err
	}
}

func updateExtension(conn *pgx.Conn, extension *v1alpha1.Extension) error {
	if extension.Version == "latest" {

		_, err := conn.Exec(
			context.Background(),
			fmt.Sprintf(
				"ALTER EXTENSION %s UPDATE",
				SanitizeString(extension.Name)),
		)

		return err
	} else {
		_, err := conn.Exec(
			context.Background(),
			fmt.Sprintf(
				"ALTER EXTENSION %s UPDATE TO %s",
				SanitizeString(extension.Name),
				SanitizeString(extension.Version),
			),
		)
		return err
	}
}

func deleteExtension(conn *pgx.Conn, extension *v1alpha1.Extension) error {
	_, err := conn.Exec(
		context.Background(),
		fmt.Sprintf(
			"DROP EXTENSION %s",
			SanitizeString(extension.Name),
		),
	)
	return err
}

func (r *extensionRepository) GetDependencies(extension *v1alpha1.Extension) (error, []string) {
	var dependencies []string

	err := pgxscan.Select(
		context.Background(),
		r.conn,
		&dependencies,
		"select extname from pg_depend join pg_extension on objid = oid where refobjid=(select oid from pg_extension where extname = $1)",
		extension.Name,
	)

	if err != nil {
		return err, nil
	}
	return nil, dependencies
}
