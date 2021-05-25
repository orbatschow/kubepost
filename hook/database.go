package hook

import (
    "github.com/gofiber/fiber/v2"
    "github.com/orbatschow/kubepost/api/v1alpha1"
    "github.com/orbatschow/kubepost/controller"
    "github.com/orbatschow/kubepost/types"
    log "github.com/sirupsen/logrus"
    v1 "k8s.io/api/core/v1"
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

func RegisterDatabaseHandlerGroup(app *fiber.App) {
    instanceRouter := app.Group("/database")
    instanceRouter.All("/sync", syncDatabase)
    instanceRouter.All("/customize", customizeDatabase)
}

func syncDatabase(c *fiber.Ctx) error {

    request := &syncDatabaseRequest{}
    if err := c.BodyParser(request); err != nil {
        log.Errorf("could not parse request: %s", err)
        return fiber.ErrBadRequest
    }

    if request.Finalizing {
        return finalizeDatabase(c, request)
    } else {
        return reconcileDatabase(c, request)
    }

}

func reconcileDatabase(c *fiber.Ctx, request *syncDatabaseRequest) error {
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
    return c.JSON(response)
}

func finalizeDatabase(c *fiber.Ctx, request *syncDatabaseRequest) error {
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

    return c.JSON(response)
}

func customizeDatabase(c *fiber.Ctx) error {

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

    return c.JSON(CustomizeResponse{
        RelatedResources: relatedResources,
    })

}
