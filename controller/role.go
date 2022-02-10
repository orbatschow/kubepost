package controller

import (
	"context"
	"encoding/base64"
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

	err := role.reconcileRole(instances, secrets)
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

	err := role.reconcileRole(instances, secrets)
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

	err := role.reconcileRole(instances, secrets)
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
		return
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

	for _, instance := range instances {
		if role.Spec.InstanceRef.Name == instance.Name {
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

func (role *Role) reconcileRole(instances map[string]*Instance, secrets map[string]*v1.Secret) error {
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

	var exists bool
	exists, err = roleRepository.DoesRoleExist(role.Spec.RoleName)
	if err != nil {
		return err
	}

	if exists {
		log.Infof(
			"role '%s' in namespace '%s' already exists, skipping creation",
			role.Spec.RoleName,
			role.Namespace,
		)
	} else {
		err = roleRepository.Create(role.Spec.RoleName)
		if err != nil {
			return err
		}
	}

	password, err := role.getRolePassword(secrets)
	if err != nil {
		return err
	}

	err = roleRepository.SetPassword(role.Spec.RoleName, password)
	if err != nil {
		return err
	}

	err = roleRepository.Alter((*v1alpha1.Role)(role))
	if err != nil {
		return err
	}

	err = role.reconcileGroups(instance, secret)
	if err != nil {
		return err
	}

	err = role.reconcileGrants(instance, secret)
	if err != nil {
		return err
	}

	return nil
}

func (role *Role) reconcileGroups(instance *Instance, secret *v1.Secret) error {

	// collect databases to connect to
	instance.Spec.Database = "postgres"
	conn, err := instance.GetConnection(secret)
	if err != nil {
		return err
	}

	roleRepository := repository.NewRoleRepository(conn)

	currentGroups, err := roleRepository.GetGroups((*v1alpha1.Role)(role))

	if err != nil {
		return err
	}

	desiredGroups := role.Spec.Groups

	desiredGroups, currentGroups = roleRepository.GetGroupGrantObjectSymmetricDifference(
		desiredGroups,
		currentGroups,
	)

	for _, undesiredGroup := range currentGroups {
		err := roleRepository.RemoveGroup((*v1alpha1.Role)(role), &undesiredGroup)
		if err != nil {
			return err
		}
	}

	for _, desiredGroup := range desiredGroups {
		err := roleRepository.AddGroup((*v1alpha1.Role)(role), &desiredGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func (role *Role) reconcileGrants(instance *Instance, secret *v1.Secret) error {

	// collect databases to connect to
	instance.Spec.Database = "postgres"
	conn, err := instance.GetConnection(secret)
	if err != nil {
		return err
	}

	roleRepository := repository.NewRoleRepository(conn)

	var databases []string
	databases, err = roleRepository.GetDatabaseNames()
	if err != nil {
		return err
	}
	conn.Close(context.Background())

	for _, database := range databases {
		grantObjects := []v1alpha1.GrantObject{}

		for _, grant := range role.Spec.Grants {
			if grant.Database == database {
				grantObjects = append(grantObjects, grant.Objects...)
			}
		}

		instance.Spec.Database = database
		conn, err := instance.GetConnection(secret)
		if err != nil {
			return err
		}

		roleRepository := repository.NewRoleRepository(conn)

		// regex
		grantObjects, _ = roleRepository.RegexExpandGrantObjects(grantObjects)

		currentGrants, err := roleRepository.GetCurrentGrants((*v1alpha1.Role)(role))

		if err != nil {
			return err
		}

		// get desired and undesired grants by subtracting the intersections of
		// current and desired grants
		desiredGrants, undesiredGrants := repository.GetGrantSymmetricDifference(
			grantObjects,
			currentGrants,
		)

		err = roleRepository.Grant((*v1alpha1.Role)(role), desiredGrants)
		if err != nil {
			return err
		}

		err = roleRepository.Revoke((*v1alpha1.Role)(role), undesiredGrants)
		if err != nil {
			return err
		}

		conn.Close(context.Background())
	}
	return nil
}

func (role *Role) getRolePassword(secrets map[string]*v1.Secret) (string, error) {

	// if the password is set via the `password` option, just return the base64 decoded value
	if len(role.Spec.Password) > 0 {
		data, err := base64.StdEncoding.DecodeString(role.Spec.Password)
		if err != nil {
			return "", errors.New(fmt.Sprintf("could not decode password for role '%s' in namespace '%s' - (should be base64 formatted)",
				role.Spec.PasswordRef.Name,
				role.Namespace,
			))
		}

		return string(data), nil
	}

	// if neither password, nor passwordRef are set, set the password to `NULL`
	if (v1alpha1.PasswordRef{} == role.Spec.PasswordRef) {
		return "NULL", nil
	}

	var rolePasswordSecret *v1.Secret
	for _, secret := range secrets {
		if role.Spec.PasswordRef.Name == secret.Name {
			rolePasswordSecret = secret
		}
	}

	if rolePasswordSecret == nil {
		return "", errors.New(fmt.Sprintf("could not find secret '%s' in namespace '%s' for role '%s'",
			role.Spec.PasswordRef.Name,
			role.Namespace,
			role.Spec.RoleName,
		),
		)
	}

	// extract the password
	passwordBytes := rolePasswordSecret.Data[role.Spec.PasswordRef.PasswordKey]
	if passwordBytes == nil {
		return "", errors.New(
			fmt.Sprintf(
				"could not find key '%s' for secret '%s' in namespace '%s' for role '%s', setting role state to '%s'",
				role.Spec.PasswordRef.PasswordKey,
				role.Spec.PasswordRef.Name,
				role.Namespace,
				role.Name,
				types.Unhealthy,
			),
		)
	}

	return string(passwordBytes), nil
}
