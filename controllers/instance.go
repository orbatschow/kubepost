package controllers

import (
	"context"
	postgresv1alpha1 "github.com/orbatschow/kubepost/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=postgres.kubepost.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=postgres.kubepost.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=postgres.kubepost.io,resources=instances/finalizers,verbs=update

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var obj postgresv1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO
	//_, err := instance.GetConnection(ctx, r.Client, obj)
	//if err != nil {
	//	return ctrl.Result{}, err
	//}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&postgresv1alpha1.Instance{}).
		Complete(r)
}
