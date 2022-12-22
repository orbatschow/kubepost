package controllers

import (
	"context"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/role"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RoleReconciler reconciles a Role object
type RoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=roles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgres.kubepost.io,resources=roles/finalizers,verbs=update

func (r *RoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var obj v1alpha1.Role
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		log.FromContext(ctx).Error(err, "could not fetch role")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	_, err := role.Reconcile(ctx, r.Client, &obj)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to reconcile role",
			"database", obj.ObjectMeta.Name,
			"namespace", obj.ObjectMeta.Namespace,
		)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Role{}).
		Complete(r)
}
