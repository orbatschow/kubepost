package controller

import (
	"errors"
	"fmt"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/repository"
	"github.com/orbatschow/kubepost/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Database v1alpha1.Database

func (database *Database) HandleDatabasePendingState(instances map[string]*Instance, secrets map[string]*v1.Secret) {
	log.Infof("database '%s' in namespace '%s' is in state '%s', reconciling",
		database.Spec.DatabaseName,
		database.Namespace,
		database.Status.Status,
	)

	err := database.createDatabase(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	database.Status.Status = types.Healthy
}

func (database *Database) HandleDatabaseHealthyState(instances map[string]*Instance, secrets map[string]*v1.Secret) {
	log.Infof("database '%s' in namespace '%s' is in state '%s', reconciling",
		database.Spec.DatabaseName,
		database.Namespace,
		database.Status.Status,
	)

	err := database.createDatabase(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	database.Status.Status = types.Healthy
}

func (database *Database) HandleDatabaseUnhealthyState(instances map[string]*Instance, secrets map[string]*v1.Secret) {
	log.Infof("database '%s' in namespace '%s' is in state '%s', reconciling",
		database.Spec.DatabaseName,
		database.Namespace,
		database.Status.Status,
	)

	err := database.createDatabase(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	database.Status.Status = types.Healthy
}

func (database *Database) HandleFinalizeDatabaseState(instances map[string]*Instance, secrets map[string]*v1.Secret)  {

	if database.Spec.PreventDeletion {
		database.Status.Status = types.Deleting
	}

	instance, err := database.getInstanceForDatabase(instances)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	conn, err := instance.GetConnection(secret)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	databaseRepository := repository.NewDatabaseRepository(conn)
	err = databaseRepository.Delete(database.Spec.DatabaseName)
	if err != nil {
		log.Errorf(err.Error())
		database.Status.Status = types.Unhealthy
		return
	}

	database.Status.Status = types.Deleting
}

func (database *Database) HandleDatabaseUnknownState() {
	log.Errorf("instance '%s' in namespace '%s' is in an unkown state, setting state to '%s'", database.Spec.DatabaseName, database.Namespace, types.Pending)
	database.Status.Status = types.Pending
}

func (database *Database) getInstanceForDatabase(instances map[string]*Instance) (*Instance, error) {

	var databaseInstance *Instance

	// if instance ref does not have namespace set, use namespace of database
	if database.Spec.InstanceRef.Namespace == "" {
		database.Spec.InstanceRef.Namespace = database.Namespace
	}

	for _, instance := range instances {
		if database.Spec.InstanceRef.Name == instance.Name && database.Spec.InstanceRef.Namespace == instance.Namespace {
			databaseInstance = instance
		}
	}

	if databaseInstance == nil {
		return nil, errors.New(fmt.Sprintf("could not find instance '%s' in namespace '%s' for database '%s' in namespace '%s', setting database state to '%s'",
			database.Spec.InstanceRef.Name,
			database.Spec.InstanceRef.Name,
			database.Spec.DatabaseName,
			database.Namespace,
			types.Unhealthy),
		)
	}

	return databaseInstance, nil
}

func (database *Database) createDatabase(instances map[string]*Instance, secrets map[string]*v1.Secret) error {
	instance, err := database.getInstanceForDatabase(instances)
	if err != nil {
		database.Status.Status = types.Unhealthy
		return err
	}

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		return err
	}

	conn, err := instance.GetConnection(secret)
	if err != nil {
		return err
	}

	databaseRepository := repository.NewDatabaseRepository(conn)
	err = databaseRepository.Create(database.Spec.DatabaseName)
	if err != nil {
		return err
	}

	err = database.reconcileExtensions(instance, secret)
	if err != nil {
		return err
	}

	return nil
}

func (database *Database) reconcileExtensions(instance *Instance, secret *v1.Secret) error  {
	// switch to the extensions target database, as you can only create
	// extensions from within the database you are connected to
	instance.Spec.Database = database.Spec.DatabaseName
	conn, err := instance.GetConnection(secret)
	if err != nil {
		return err
	}

	extensionRepository := repository.NewExtensionRepository(conn)
	err = extensionRepository.Reconcile(database.Spec.Extensions)
	if err != nil {
		return err
	}

	return nil
}
