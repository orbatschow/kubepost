package controllers

import (
	"context"
	"github.com/orbatschow/kubepost/pkg/database"
	"github.com/orbatschow/kubepost/pkg/extension"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/orbatschow/kubepost/api/v1alpha1"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var obj v1alpha1.Database
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		log.FromContext(ctx).Error(err, "could not fetch database")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	_, connections, err := database.Reconcile(ctx, r.Client, &obj)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to reconcile database",
			"database", obj.ObjectMeta.Name,
			"namespace", obj.ObjectMeta.Namespace,
		)
		return ctrl.Result{}, err
	}

	if connections == nil {
		return ctrl.Result{}, nil
	}

	// skip everything else, if deletion is scheduled
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	err = extension.Reconcile(ctx, r.Client, connections, &obj)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to reconcile database",
			"database", obj.ObjectMeta.Name,
			"namespace", obj.ObjectMeta.Namespace,
		)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Database{}).
		Complete(r)
}
