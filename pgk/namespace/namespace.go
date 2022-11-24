package namespace

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(ctx context.Context, ctrlClient client.Client, namespacesSelector metav1.LabelSelector) ([]v1.Namespace, error) {

	var namespaces v1.NamespaceList

	selector, err := metav1.LabelSelectorAsSelector(&namespacesSelector)
	if err != nil {
		return nil, err
	}

	err = ctrlClient.List(ctx, &namespaces, client.MatchingLabelsSelector{
		Selector: selector,
	})

	if err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}
