package controller

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"strconv"
)

// TODO: implement deletion finalizer

type Instance v1alpha1.Instance

func (instance *Instance) HandleInstancePendingState(secrets map[string]*v1.Secret) {
	log.Infof("instance '%s' in namespace '%s' is in state '%s', reconciling", instance.Name, instance.Namespace, instance.Status.Status)

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}
	_, err = instance.GetConnection(secret)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}

}

func (instance *Instance) HandleInstanceHealthyState(secrets map[string]*v1.Secret) {
	log.Infof("instance '%s' in namespace '%s' is in state '%s', reconciling",
		instance.Name,
		instance.Namespace,
		instance.Status.Status,
	)

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}

	_, err = instance.GetConnection(secret)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}

	instance.Status.Status = types.Healthy

}

func (instance *Instance) HandleInstanceUnhealthyState(secrets map[string]*v1.Secret) {
	log.Infof("instance '%s' in namespace '%s' is in state '%s', reconciling",
		instance.Name,
		instance.Namespace,
		instance.Status.Status,
	)

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}

	_, err = instance.GetConnection(secret)
	if err != nil {
		log.Errorf(err.Error())
		instance.Status.Status = types.Unhealthy
		return
	}

	instance.Status.Status = types.Healthy

}

func (instance *Instance) HandleUnknownState() {
	log.Errorf("instance '%s' in namespace '%s' is in an unkown state, setting state to '%s'", instance.Name, instance.Namespace, types.Pending)
	instance.Status.Status = types.Pending
}

func (instance *Instance) GetSecret(secrets map[string]*v1.Secret) (*v1.Secret, error) {


	var instanceSecret *v1.Secret
	for _, secret := range secrets {
		if instance.Spec.SecretRef.Name == secret.Name {
			instanceSecret = secret
		}
	}

	if instanceSecret == nil {
		return nil, errors.New(fmt.Sprintf("could not find secret '%s' in namespace '%s' for instance '%s'",
			instance.Spec.SecretRef.Name,
			instance.Namespace,
			instance.Name,
		),
		)
	}

	return instanceSecret, nil

}

func (instance *Instance) GetConnection(secret *v1.Secret) (*pgx.Conn, error) {

	usernameBytes := secret.Data[instance.Spec.SecretRef.UserKey]
	if usernameBytes == nil {
		return nil, errors.New(
			fmt.Sprintf(
				"could not find key '%s' for secret '%s' in namespace '%s' for instance '%s', setting instance state to '%s'",
				instance.Spec.SecretRef.UserKey,
				instance.Spec.SecretRef.Name,
				instance.Namespace,
				instance.Name,
				types.Unhealthy,
			),
		)
	}

	passwordBytes := secret.Data[instance.Spec.SecretRef.PasswordKey]
	if passwordBytes == nil {
		return nil, errors.New(
			fmt.Sprintf(
				"could not find key '%s' for secret '%s' in namespace '%s' for instance '%s', setting instance state to '%s'",
				instance.Spec.SecretRef.UserKey,
				instance.Spec.SecretRef.Name,
				instance.Namespace,
				instance.Name,
				types.Unhealthy,
			),
		)
	}

	p := Postgres{
		Host:     instance.Spec.Host,
		Port:     strconv.Itoa(instance.Spec.Port),
		Username: string(usernameBytes),
		Password: string(passwordBytes),
		Database: instance.Spec.Database,
		SSLMode:  instance.Spec.SSLMode,
	}

	return p.GetConnection()

}
