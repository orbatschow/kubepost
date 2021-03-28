package hook

import (
	"encoding/json"
	"github.com/orbatschow/kubepost/api/v1alpha1"
	"github.com/orbatschow/kubepost/controller"
	"github.com/orbatschow/kubepost/types"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"net/http"
)

type syncRoleRequest struct {
	Role    controller.Role `json:"object"`
	Related struct {
		Instances map[string]*controller.Instance `json:"Instance.kubepost.io/v1alpha1"`
		Secrets   map[string]*v1.Secret           `json:"Secret.v1"`
	} `json:"related"`
	Finalizing bool `json:"finalizing"`
}

type syncRoleResponse struct {
	Status v1alpha1.RoleStatus `json:"status"`
}

type finalizeRoleResponse struct {
	Status v1alpha1.RoleStatus `json:"status"`
	Finalized bool `json:"finalized"`
}

func syncRole(request *syncRoleRequest) (*syncRoleResponse, error) {
	response := &syncRoleResponse{}

	role := request.Role
	instances := request.Related.Instances
	secrets := request.Related.Secrets

	switch role.Status.Status {
	case types.Pending:
		role.HandleRolePendingState(instances, secrets)
	case types.Healthy:
		role.HandleRoleHealthyState(instances, secrets)
	case types.Unhealthy:
		role.HandleRoleUnhealthyState(instances, secrets)
	default:
		role.HandleRoleUnknownState()
	}

	response.Status = role.Status
	return response, nil
}

func finalizeRole(request *syncRoleRequest) (*finalizeRoleResponse, error) {
	response := &finalizeRoleResponse{}

	role := request.Role
	instances := request.Related.Instances
	secrets := request.Related.Secrets

	role.HandleFinalizeRoleState(instances, secrets)

	response.Status = role.Status
	if role.Status.Status == types.Deleting {
		response.Finalized = true
	} else {
		response.Finalized = false
	}

	return response, nil
}

func RoleSyncHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := &syncRoleRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response interface{}
	if !request.Finalizing {
		response, err = syncRole(request)
	} else {
		response, err = finalizeRole(request)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err = json.Marshal(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func RoleCustomizeHandler(w http.ResponseWriter, r *http.Request) {

	relatedResources := []RelatedResource{
		{
			ApiVersion: "kubepost.io/v1alpha1",
			Resource:   "instances",
		},
		{
			ApiVersion: "v1",
			Resource:   "secrets",
		},
	}

	response := CustomizeResponse{RelatedResources: relatedResources}

	CustomizeHandler(w, r, &response)

}
