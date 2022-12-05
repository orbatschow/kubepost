package database

import (
	"context"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pgk/instance"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var Finalizer = "finalizer.postgres.kubepost.io/database"

func Reconcile(ctx context.Context, ctrlClient client.Client, db *v1alpha1.Database) (*v1alpha1.Database, []v1alpha1.Instance, error) {

	instances, err := instance.List(ctx, ctrlClient, db.Spec.InstanceNamespaceSelector, db.Spec.InstanceSelector)
	if err != nil {
		return nil, nil, err
	}

	for _, postgres := range instances {
		log.FromContext(ctx).Info(
			"reconciling database",
			"instance", types.NamespacedName{
				Namespace: postgres.ObjectMeta.Namespace,
				Name:      postgres.ObjectMeta.Name,
			},
		)

		conn, err := instance.GetConnection(ctx, ctrlClient, &postgres)
		if err != nil {
			log.FromContext(ctx).Error(
				err,
				"failed to establish a connection",
				"instance", types.NamespacedName{
					Namespace: postgres.ObjectMeta.Namespace,
					Name:      postgres.ObjectMeta.Name,
				},
			)
			continue
		}

		repository := Repository{
			database: db,
			instance: &postgres,
			conn:     conn,
		}

		err = repository.handleFinalizer(ctx, ctrlClient)
		if err != nil {
			return nil, nil, err
		}

		// skip everything else, if deletion is scheduled
		if !repository.database.ObjectMeta.DeletionTimestamp.IsZero() {
			return nil, nil, nil
		}

		exists, err := repository.Exists(ctx)
		if err != nil {
			return nil, nil, err
		}

		if exists == true {
			log.FromContext(ctx).Info(
				"database exists, skipping creation",
				"instance", types.NamespacedName{
					Namespace: postgres.ObjectMeta.Namespace,
					Name:      postgres.ObjectMeta.Name,
				},
			)
		} else {
			err = repository.Create(ctx)
			if err != nil {
				return nil, nil, err
			}
		}

		err = repository.AlterOwner(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	return db, instances, nil

}

func (r *Repository) handleFinalizer(ctx context.Context, ctrClient client.Client) error {
	switch {
	// handle deletion and remove finalizer.
	case !r.database.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(r.database, Finalizer):
		log.FromContext(ctx).Info("handling database deletion",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

		// delete the database
		exists, err := r.Exists(ctx)
		if err != nil {
			return err
		}

		if exists {
			err := r.Delete(ctx)
			if err != nil {
				return err
			}
		}

		// remove the finalizer
		controllerutil.RemoveFinalizer(r.database, Finalizer)
		if err := ctrClient.Update(ctx, r.database); err != nil {
			return err
		}

		log.FromContext(ctx).Info("removing finalizer",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

	// deletion already handled, don't do anything.
	case !r.database.DeletionTimestamp.IsZero() && !slices.Contains(r.database.ObjectMeta.Finalizers, Finalizer):
		log.FromContext(ctx).Info("deletion pending",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

	// add finalizer to the object.
	case r.database.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(r.database, Finalizer):
		log.FromContext(ctx).Info(
			"updating finalizers",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

		controllerutil.AddFinalizer(r.database, Finalizer)
		if err := ctrClient.Update(ctx, r.database); err != nil {
			return err
		}
	}

	return nil
}
