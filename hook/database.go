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

type syncDatabaseRequest struct {
	Database controller.Database `json:"object"`
	Related  struct {
		Instances map[string]*controller.Instance `json:"Instance.kubepost.io/v1alpha1"`
		Secrets   map[string]*v1.Secret           `json:"Secret.v1"`
	} `json:"related"`
	Finalizing bool `json:"finalizing"`
}

type syncDatabaseResponse struct {
	Status v1alpha1.DatabaseStatus `json:"status"`
}

type finalizeDatabaseResponse struct {
	Status    v1alpha1.DatabaseStatus `json:"status"`
	Finalized bool                    `json:"finalized"`
}

func syncDatabase(request *syncDatabaseRequest) (*syncDatabaseResponse, error) {
	response := &syncDatabaseResponse{}

	database := request.Database
	instances := request.Related.Instances
	secrets := request.Related.Secrets

	switch database.Status.Status {
	case types.Pending:
		database.HandleDatabasePendingState(instances, secrets)
	case types.Healthy:
		database.HandleDatabaseHealthyState(instances, secrets)
	case types.Unhealthy:
		database.HandleDatabaseUnhealthyState(instances, secrets)
	default:
		database.HandleDatabaseUnknownState()
	}

	response.Status = database.Status
	return response, nil
}

func finalizeDatabase(request *syncDatabaseRequest) (*finalizeDatabaseResponse, error) {
	response := &finalizeDatabaseResponse{}

	database := request.Database
	instances := request.Related.Instances
	secrets := request.Related.Secrets

	database.HandleFinalizeDatabaseState(instances, secrets)

	response.Status = database.Status
	if database.Status.Status == types.Deleting {
		response.Finalized = true
	} else {
		response.Finalized = false
	}

	return response, nil
}

func DatabaseSyncHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := &syncDatabaseRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response interface{}
	if !request.Finalizing {
		response, err = syncDatabase(request)
	} else {
		response, err = finalizeDatabase(request)
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

func DatabaseCustomizeHandler(w http.ResponseWriter, r *http.Request) {

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
