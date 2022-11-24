package extension

import (
	"context"
	v1alpha1 "github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pgk/instance"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Reconcile(ctx context.Context, ctrlClient client.Client, instances []v1alpha1.Instance, db *v1alpha1.Database) error {

	for _, postgres := range instances {

		// we have to connect to the desired database, so a switch from the instance database is performed here
		postgres.Spec.Database = db.ObjectMeta.Name

		conn, err := instance.GetConnection(ctx, ctrlClient, &postgres)
		if err != nil {
			log.FromContext(ctx).Error(
				err,
				"failed to establish a connection",
				"instance", postgres.ObjectMeta.Name,
			)
		}

		repository := Repository{
			conn:     conn,
			instance: &postgres,
			database: db,
		}

		existingExtensions, err := repository.List(ctx)

		// create missing extensions, update existing ones
		// only applies to configured extensions, all other extensions won't be touched
		for _, desiredExtension := range repository.database.Spec.Extensions {
			// check if desired extension already exists
			var exists bool
			for _, existingExtension := range existingExtensions {
				if desiredExtension.Name == existingExtension.Name {
					exists = true
				}
			}

			// create desired extension if it does not already exist, otherwise update
			if exists != true {
				err = repository.Create(ctx, &desiredExtension)
				if err != nil {
					return err
				}
			} else {
				err = repository.Update(ctx, &desiredExtension)
				if err != nil {
					return err
				}
			}

		}

		// check if existing extension is still desired
		// if the extension is not desired anymore, delete is

		// if there are extensions, that rely on a extension, which is scheduled for deletion,
		// this extension won't be touched
		for _, existingExtension := range existingExtensions {
			var desired bool

			for _, desiredExtension := range repository.database.Spec.Extensions {
				if existingExtension.Name == desiredExtension.Name {
					desired = true
				}
			}

			// delete existing extension if it is not desired
			if desired != true {

				// check if existingExtension is dependency of other extension
				err, childExtensions := repository.GetChildExtensions(ctx, &existingExtension)

				if err != nil {
					return err
				}

				if len(childExtensions) > 0 {
					log.FromContext(ctx).V(4).Info(
						"skipping deletion for extension, unresolved dependencies",
						"extension", existingExtension.Name,
						"children", childExtensions,
					)
				} else {
					err = repository.Delete(ctx, &existingExtension)
					if err != nil {
						return err
					}
				}
			}
		}

	}
	return nil
}
