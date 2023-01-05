package role

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

var Finalizer = "finalizer.postgres.kubepost.io/role"

func Reconcile(ctx context.Context, ctrlClient client.Client, role *v1alpha1.Role) (*v1alpha1.Role, error) {

	connections, err := connection.List(ctx, ctrlClient, role.Spec.ConnectionNamespaceSelector, role.Spec.ConnectionSelector)
	if err != nil {
		return nil, err
	}

	for _, postgres := range connections {
		conn, err := connection.GetConnection(ctx, ctrlClient, &postgres)
		if err != nil {
			log.FromContext(ctx).Error(
				err,
				"failed to establish a connection",
				"connection", postgres.ObjectMeta.Name,
			)
			continue
		}

		repository := Repository{
			conn:       conn,
			connection: &postgres,
			role:       role,
		}

		err = repository.handleFinalizer(ctx, ctrlClient)
		if err != nil {
			return nil, err
		}

		// skip everything else, if deletion is scheduled
		if !repository.role.ObjectMeta.DeletionTimestamp.IsZero() {
			return nil, nil
		}

		var exists bool
		exists, err = repository.Exists(ctx)
		if err != nil {
			return nil, err
		}

		if exists {
			log.FromContext(ctx).Info(
				"role exists, skipping creation",
				"connection", types.NamespacedName{
					Namespace: postgres.ObjectMeta.Namespace,
					Name:      postgres.ObjectMeta.Name,
				},
			)
		} else {
			err = repository.Create(ctx)
			if err != nil {
				return nil, err
			}
		}

		password, err := repository.GetPassword(ctx, ctrlClient)
		if err != nil {
			return nil, err
		}

		err = repository.SetPassword(ctx, password)
		if err != nil {
			return nil, err
		}

		err = repository.Alter(ctx)
		if err != nil {
			return nil, err
		}

		err = repository.ReconcileGroups(ctx)
		if err != nil {
			return nil, err
		}

		err = repository.ReconcileGrants(ctx, ctrlClient)
		if err != nil {
			return nil, err
		}
	}

	return role, nil
}

func (r *Repository) handleFinalizer(ctx context.Context, ctrClient client.Client) error {
	switch {
	// handle deletion and remove finalizer.
	case !r.role.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(r.role, Finalizer):
		log.FromContext(ctx).Info("handling role deletion",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

		if r.role.Spec.Protected != true {
			// delete the role if it is not protected
			log.FromContext(ctx).Info("postgres role will be deleted, protection is turned off",
				"connection", types.NamespacedName{
					Namespace: r.connection.ObjectMeta.Namespace,
					Name:      r.connection.ObjectMeta.Name,
				},
			)
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
		} else {
			log.FromContext(ctx).Info("postgres role will not be deleted, protection is turned on",
				"connection", types.NamespacedName{
					Namespace: r.connection.ObjectMeta.Namespace,
					Name:      r.connection.ObjectMeta.Name,
				},
			)
		}

		// remove the finalizer
		controllerutil.RemoveFinalizer(r.role, Finalizer)
		if err := ctrClient.Update(ctx, r.role); err != nil {
			return err
		}

		log.FromContext(ctx).Info("removing finalizer",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

	// deletion already handled, don't do anything.
	case !r.role.DeletionTimestamp.IsZero() && !slices.Contains(r.role.ObjectMeta.Finalizers, Finalizer):
		log.FromContext(ctx).Info("deletion pending",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

	// add finalizer to the object.
	case r.role.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(r.role, Finalizer):
		log.FromContext(ctx).Info(
			"updating finalizers",
			"connection", types.NamespacedName{
				Namespace: r.connection.ObjectMeta.Namespace,
				Name:      r.connection.ObjectMeta.Name,
			},
		)

		controllerutil.AddFinalizer(r.role, Finalizer)
		if err := ctrClient.Update(ctx, r.role); err != nil {
			return err
		}
	}

	return nil
}
