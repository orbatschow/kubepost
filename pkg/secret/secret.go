package secret

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(ctx context.Context, ctrlClient client.Client, namespacedName types.NamespacedName) (*v1.Secret, error) {
	var secret v1.Secret
	err := ctrlClient.Get(ctx, namespacedName, &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}
