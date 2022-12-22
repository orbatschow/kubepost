package connection

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/pkg/namespace"
	"github.com/orbatschow/kubepost/pkg/postgres"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func List(ctx context.Context, ctrlClient client.Client, connectionNamespaceSelector metav1.LabelSelector, connectionSelector metav1.LabelSelector) ([]v1alpha1.Connection, error) {
	namespaces, err := namespace.List(ctx, ctrlClient, connectionNamespaceSelector)
	if err != nil {
		return nil, err
	}

	var connections []v1alpha1.Connection
	selector, err := metav1.LabelSelectorAsSelector(&connectionSelector)
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces {
		var buffer v1alpha1.ConnectionList
		err = ctrlClient.List(ctx, &buffer, client.InNamespace(ns.Name), client.MatchingLabelsSelector{
			Selector: selector,
		})

		connections = append(connections, buffer.Items...)
		if err != nil {
			return nil, err
		}
	}

	return connections, nil
}

func GetConnection(ctx context.Context, client client.Client, connection *v1alpha1.Connection) (*pgx.Conn, error) {
	usernameRef := types.NamespacedName{
		Namespace: connection.ObjectMeta.Namespace,
		Name:      connection.Spec.Username.Name,
	}

	passwordRef := types.NamespacedName{
		Namespace: connection.ObjectMeta.Namespace,
		Name:      connection.Spec.Username.Name,
	}

	var usernameSecret v1.Secret
	if err := client.Get(ctx, usernameRef, &usernameSecret); err != nil {
		return nil, errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, usernameSecret.Name)
	}

	var passwordSecret v1.Secret
	if err := client.Get(ctx, passwordRef, &passwordSecret); err != nil {
		return nil, errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, passwordSecret.Name)
	}

	usernameBytes := usernameSecret.Data[connection.Spec.Username.Key]
	if usernameBytes == nil {
		return nil, fmt.Errorf("could not parse username for connection '%s/%s' from secret '%s/%s", connection.ObjectMeta.Namespace, connection.ObjectMeta.Name, connection.ObjectMeta.Namespace, usernameSecret.ObjectMeta.Name)
	}

	passwordBytes := passwordSecret.Data[connection.Spec.Password.Key]
	if passwordBytes == nil {
		return nil, fmt.Errorf("could not parse password for connection '%s/%s' from secret '%s/%s", connection.ObjectMeta.Namespace, connection.ObjectMeta.Name, connection.ObjectMeta.Namespace, usernameSecret.ObjectMeta.Name)
	}

	p := postgres.Postgres{
		Host:     connection.Spec.Host,
		Port:     strconv.Itoa(connection.Spec.Port),
		Username: string(usernameBytes),
		Password: string(passwordBytes),
		Database: connection.Spec.Database,
		SSLMode:  connection.Spec.SSLMode,
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
