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

type Role v1alpha1.Role

func (role *Role) HandleRolePendingState(instances map[string]*Instance, secrets map[string]*v1.Secret) {

	log.Infof("role '%s' in namespace '%s' is in state '%s', reconciling",
		role.Spec.RoleName,
		role.Namespace,
		role.Status.Status,
	)

	err := role.createRole(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	role.Status.Status = types.Healthy

}

func (role *Role) HandleRoleHealthyState(instances map[string]*Instance, secrets map[string]*v1.Secret) {

	log.Infof("role '%s' in namespace '%s' is in state '%s', reconciling",
		role.Spec.RoleName,
		role.Namespace,
		role.Status.Status,
	)

	err := role.createRole(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	role.Status.Status = types.Healthy

}

func (role *Role) HandleRoleUnhealthyState(instances map[string]*Instance, secrets map[string]*v1.Secret) {
	log.Infof("role '%s' in namespace '%s' is in state '%s', reconciling",
		role.Spec.RoleName,
		role.Namespace,
		role.Status.Status,
	)

	err := role.createRole(instances, secrets)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	role.Status.Status = types.Healthy

}

func (role *Role) HandleFinalizeRoleState(instances map[string]*Instance, secrets map[string]*v1.Secret) {

	if role.Spec.PreventDeletion {
		role.Status.Status = types.Deleting
	}

	instance, err := role.getInstanceForRole(instances)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	secret, err := instance.GetSecret(secrets)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	conn, err := instance.GetConnection(secret)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	roleRepository := repository.NewRoleRepository(conn)
	err = roleRepository.Delete(role.Spec.RoleName)
	if err != nil {
		log.Errorf(err.Error())
		role.Status.Status = types.Unhealthy
		return
	}

	role.Status.Status = types.Deleting
}

func (role *Role) HandleRoleUnknownState() {
	log.Errorf("instance '%s' in namespace '%s' is in an unkown state, setting state to '%s'", role.Spec.RoleName, role.Namespace, types.Pending)
	role.Status.Status = types.Pending
}

func (role *Role) getInstanceForRole(instances map[string]*Instance) (*Instance, error) {

	var roleInstance *Instance

	// if instance ref does not have namespace set, use namespace of role
	if role.Spec.InstanceRef.Namespace == "" {
		role.Spec.InstanceRef.Namespace = role.Namespace
	}

	for _, instance := range instances {
		if role.Spec.InstanceRef.Name == instance.Name && role.Spec.InstanceRef.Namespace == instance.Namespace {
			roleInstance = instance
		}
	}

	if roleInstance == nil {
		return nil, errors.New(fmt.Sprintf("could not find instance '%s' in namespace '%s' for role '%s' in namespace '%s', setting role state to '%s'",
			role.Spec.InstanceRef.Name,
			role.Spec.InstanceRef.Name,
			role.Spec.RoleName,
			role.Namespace,
			types.Unhealthy),
		)
	}

	return roleInstance, nil
}

func (role *Role) createRole(instances map[string]*Instance, secrets map[string]*v1.Secret) error {
	instance, err := role.getInstanceForRole(instances)
	if err != nil {
		role.Status.Status = types.Unhealthy
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

	roleRepository := repository.NewRoleRepository(conn)
	err = roleRepository.Create(role.Spec.RoleName)
	if err != nil {
		return err
	}

	err = roleRepository.Grant(role.Spec.RoleName, &role.Spec.Grant)
	if err != nil {
		return err
	}

	return nil
}
