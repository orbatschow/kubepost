package role

import (
	"context"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/instance"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var Finalizer = "finalizer.postgres.kubepost.io/role"

func Reconcile(ctx context.Context, ctrlClient client.Client, role *v1alpha1.Role) (*v1alpha1.Role, error) {

	instances, err := instance.List(ctx, ctrlClient, role.Spec.InstanceNamespaceSelector, role.Spec.InstanceSelector)
	if err != nil {
		return nil, err
	}

	for _, postgres := range instances {
		conn, err := instance.GetConnection(ctx, ctrlClient, &postgres)
		if err != nil {
			log.FromContext(ctx).Error(
				err,
				"failed to establish a connection",
				"instance", postgres.ObjectMeta.Name,
			)
			continue
		}

		repository := Repository{
			conn:     conn,
			instance: &postgres,
			role:     role,
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
				"instance", types.NamespacedName{
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
		controllerutil.RemoveFinalizer(r.role, Finalizer)
		if err := ctrClient.Update(ctx, r.role); err != nil {
			return err
		}

		log.FromContext(ctx).Info("removing finalizer",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

	// deletion already handled, don't do anything.
	case !r.role.DeletionTimestamp.IsZero() && !slices.Contains(r.role.ObjectMeta.Finalizers, Finalizer):
		log.FromContext(ctx).Info("deletion pending",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

	// add finalizer to the object.
	case r.role.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(r.role, Finalizer):
		log.FromContext(ctx).Info(
			"updating finalizers",
			"instance", types.NamespacedName{
				Namespace: r.instance.ObjectMeta.Namespace,
				Name:      r.instance.ObjectMeta.Name,
			},
		)

		controllerutil.AddFinalizer(r.role, Finalizer)
		if err := ctrClient.Update(ctx, r.role); err != nil {
			return err
		}
	}

	return nil
}
