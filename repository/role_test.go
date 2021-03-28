package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/controller"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

var conn *pgx.Conn


func TestGetNotExistingRole(t *testing.T) {
	roleName := "kubepost"

	roleRepository := NewRoleRepository(conn)
	exist, err := roleRepository.DoesRoleExist(roleName)
	assert.NoError(t, err)
	assert.False(t, exist)
}

func TestCreateRole(t *testing.T) {
	roleName := "kubepost"
	roleRepository := NewRoleRepository(conn)
	err := roleRepository.Create(roleName)

	assert.NoError(t, err)

	exist, err := roleRepository.DoesRoleExist(roleName)
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestCreateRoleAlreadyExists(t *testing.T) {
	roleName := "kubepost"
	_, err := conn.Exec(context.Background(), fmt.Sprintf("CREATE ROLE %s", roleName))
	assert.NoError(t, err)

	roleRepository := NewRoleRepository(conn)
	err = roleRepository.Create(roleName)

	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	password := "root"

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "13.2", []string{fmt.Sprintf("POSTGRES_PASSWORD=%s", password)})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		p := controller.Postgres{
			Host:     "localhost",
			Port:     resource.GetPort("5432/tcp"),
			Username: "postgres",
			Password: password,
			Database: "postgres",
		}
		conn, err = p.GetConnection()
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
