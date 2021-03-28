module github.com/orbatschow/kubepost

go 1.16

require (
	github.com/jackc/pgconn v1.8.1
	github.com/jackc/pgx/v4 v4.11.0
	github.com/lib/pq v1.10.0
	github.com/ory/dockertest/v3 v3.6.3
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	sigs.k8s.io/controller-runtime v0.8.3
)
