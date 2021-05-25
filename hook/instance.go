package hook

import (
    "github.com/gofiber/fiber/v2"
    "github.com/orbatschow/kubepost/api/v1alpha1"
    "github.com/orbatschow/kubepost/controller"
    "github.com/orbatschow/kubepost/types"
    log "github.com/sirupsen/logrus"
    v1 "k8s.io/api/core/v1"
)

type syncInstanceRequest struct {
    Instance controller.Instance `json:"object"`
    Related  struct {
        Secrets   map[string]*v1.Secret           `json:"Secret.v1"`
        Databases map[string]*controller.Database `json:"Database.kubepost.io/v1alpha1"`
        Roles     map[string]*controller.Role     `json:"Role.kubepost.io/v1alpha1"`
    } `json:"related"`
    Finalizing bool `json:"finalizing"`
}

type reconcileInstanceResponse struct {
    Status v1alpha1.InstanceStatus `json:"status"`
}

type finalizeInstanceResponse struct {
    Status    v1alpha1.InstanceStatus `json:"status"`
    Finalized bool                    `json:"finalized"`
}

func RegisterInstanceHandlerGroup(app *fiber.App) {
    instanceRouter := app.Group("/instance")
    instanceRouter.All("/sync", syncInstance)
    instanceRouter.All("/customize", customizeInstance)
}

func syncInstance(c *fiber.Ctx) error {

    request := &syncInstanceRequest{}
    if err := c.BodyParser(request); err != nil {
        log.Errorf("could not parse request: %s", err)
        return fiber.ErrBadRequest
    }

    if request.Finalizing {
        return finalizeInstance(c, request)
    } else {
        return reconcileInstance(c, request)
    }

}

func reconcileInstance(c *fiber.Ctx, request *syncInstanceRequest) error {
    response := &reconcileInstanceResponse{}
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
    return c.JSON(response)
}

func finalizeInstance(c *fiber.Ctx, request *syncInstanceRequest) error {

    response := &finalizeInstanceResponse{}

    instance := request.Instance
    databases := request.Related.Databases
    roles := request.Related.Roles

    instance.HandleFinalizeInstanceState(databases, roles)

    response.Status = instance.Status
    if instance.Status.Status == types.Deleting {
        response.Finalized = true
    } else {
        response.Finalized = false
    }

    return c.JSON(response)
}

func customizeInstance(c *fiber.Ctx) error {

    relatedResources := []RelatedResource{
        {
            ApiVersion: "v1",
            Resource:   "secrets",
        },
        {
            ApiVersion: "kubepost.io/v1alpha1",
            Resource:   "databases",
        },
        {
            ApiVersion: "kubepost.io/v1alpha1",
            Resource:   "roles",
        },
    }

    return c.JSON(CustomizeResponse{
        RelatedResources: relatedResources,
    })
}
