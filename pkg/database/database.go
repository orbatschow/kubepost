package database

import (
	"context"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/connection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var Finalizer = "finalizer.postgres.kubepost.io/database"

func Reconcile(ctx context.Context, ctrlClient client.Client, db *v1alpha1.Database) (*v1alpha1.Database, []v1alpha1.Connection, error) {

	connections, err := connection.List(ctx, ctrlClient, db.Spec.ConnectionNamespaceSelector, db.Spec.ConnectionSelector)
	if err != nil {
		return nil, nil, err
	}

	for _, postgres := range connections {
		log.FromContext(ctx).Info(
			"reconciling database",
			"connection", types.NamespacedName{
				Namespace: postgres.ObjectMeta.Namespace,
				Name:      postgres.ObjectMeta.Name,
			},
		)

		conn, err := connection.GetConnection(ctx, ctrlClient, &postgres)
		if err != nil {
			log.FromContext(ctx).Error(
				err,
				"failed to establish a connection",
				"connection", types.NamespacedName{
					Namespace: postgres.ObjectMeta.Namespace,
					Name:      postgres.ObjectMeta.Name,
				},
			)
			continue
		}

		repository := Repository{
			database:   db,
			connection: &postgres,
			conn:       conn,
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
				"connection", types.NamespacedName{
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

	return db, connections, nil
}

func (r *Repository) handleFinalizer(ctx context.Context, ctrClient client.Client) error {
	switch {
	// handle deletion and remove finalizer.
	case !r.database.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(r.database, Finalizer):
		log.FromContext(ctx).Info("handling database deletion",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
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
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

	// deletion already handled, don't do anything.
	case !r.database.DeletionTimestamp.IsZero() && !slices.Contains(r.database.ObjectMeta.Finalizers, Finalizer):
		log.FromContext(ctx).Info("deletion pending",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

	// add finalizer to the object.
	case r.database.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(r.database, Finalizer):
		log.FromContext(ctx).Info(
			"updating finalizers",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

		controllerutil.AddFinalizer(r.database, Finalizer)
		if err := ctrClient.Update(ctx, r.database); err != nil {
			return err
		}
	}

	return nil
}
