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

type syncInstanceRequest struct {
	Instance controller.Instance `json:"object"`
	Related  struct {
		Secrets map[string]*v1.Secret `json:"Secret.v1"`
	} `json:"related"`
}

type syncInstanceResponse struct {
	Status v1alpha1.InstanceStatus `json:"status"`
}

func syncInstance(request *syncInstanceRequest) (*syncInstanceResponse, error) {
	response := &syncInstanceResponse{}
	instance := request.Instance

	switch instance.Status.Status {
	case types.Pending:
		instance.HandleInstancePendingState(request.Related.Secrets)
	case types.Healthy:
		instance.HandleInstanceHealthyState(request.Related.Secrets)
	case types.Unhealthy:
		instance.HandleInstanceUnhealthyState(request.Related.Secrets)
	default:
		instance.HandleUnknownState()
	}

	response.Status = instance.Status
	return response, nil
}

func InstanceSyncHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := &syncInstanceRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := syncInstance(request)
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


func InstanceCustomizeHandler(w http.ResponseWriter, r *http.Request) {

	relatedResources := []RelatedResource{
		{
			ApiVersion: "v1",
			Resource:   "secrets",
		},
	}

	response := CustomizeResponse{RelatedResources: relatedResources}

	CustomizeHandler(w, r, &response)

}
