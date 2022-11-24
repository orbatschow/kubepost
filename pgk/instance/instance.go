package instance

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pgk/namespace"
	"github.com/orbatschow/kubepost/pgk/postgres"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func List(ctx context.Context, ctrlClient client.Client, instanceNamespaceSelector metav1.LabelSelector, instanceSelector metav1.LabelSelector) ([]v1alpha1.Instance, error) {

	namespaces, err := namespace.List(ctx, ctrlClient, instanceNamespaceSelector)
	if err != nil {
		return nil, err
	}

	var instances []v1alpha1.Instance
	selector, err := metav1.LabelSelectorAsSelector(&instanceSelector)
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces {
		var buffer v1alpha1.InstanceList
		err = ctrlClient.List(ctx, &buffer, client.InNamespace(ns.Name), client.MatchingLabelsSelector{
			Selector: selector,
		})

		instances = append(instances, buffer.Items...)
		if err != nil {
			return nil, err
		}
	}

	return instances, nil
}

func GetConnection(ctx context.Context, client client.Client, instance *v1alpha1.Instance) (*pgx.Conn, error) {

	usernameRef := types.NamespacedName{
		Namespace: instance.ObjectMeta.Namespace,
		Name:      instance.Spec.Username.Name,
	}

	passwordRef := types.NamespacedName{
		Namespace: instance.ObjectMeta.Namespace,
		Name:      instance.Spec.Username.Name,
	}

	var usernameSecret v1.Secret
	if err := client.Get(ctx, usernameRef, &usernameSecret); err != nil {
		return nil, errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, usernameSecret.Name)
	}

	var passwordSecret v1.Secret
	if err := client.Get(ctx, passwordRef, &passwordSecret); err != nil {
		return nil, errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, passwordSecret.Name)
	}

	usernameBytes := usernameSecret.Data[instance.Spec.Username.Key]
	if usernameBytes == nil {
		return nil, fmt.Errorf("could not parse username for instance '%s/%s' from secret '%s/%s", instance.ObjectMeta.Namespace, instance.ObjectMeta.Name, instance.ObjectMeta.Namespace, usernameSecret.ObjectMeta.Name)
	}

	passwordBytes := passwordSecret.Data[instance.Spec.Password.Key]
	if passwordBytes == nil {
		return nil, fmt.Errorf("could not parse password for instance '%s/%s' from secret '%s/%s", instance.ObjectMeta.Namespace, instance.ObjectMeta.Name, instance.ObjectMeta.Namespace, usernameSecret.ObjectMeta.Name)
	}

	p := postgres.Postgres{
		Host:     instance.Spec.Host,
		Port:     strconv.Itoa(instance.Spec.Port),
		Username: string(usernameBytes),
		Password: string(passwordBytes),
		Database: instance.Spec.Database,
		SSLMode:  instance.Spec.SSLMode,
	}

	conn, err := pgx.Connect(context.Background(), fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=kubepost",
		p.Username,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
	),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: '%s' on host '%s' with user '%s' : '%s'", p.Database, p.Host, p.Username, err)
	}

	return conn, nil
}
