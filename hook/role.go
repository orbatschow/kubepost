package hook

import (
    "github.com/gofiber/fiber/v2"
    "github.com/orbatschow/kubepost/api/v1alpha1"
    "github.com/orbatschow/kubepost/controller"
    "github.com/orbatschow/kubepost/types"
    log "github.com/sirupsen/logrus"
    v1 "k8s.io/api/core/v1"
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
    Status    v1alpha1.RoleStatus `json:"status"`
    Finalized bool                `json:"finalized"`
}

func RegisterRoleHandlerGroup(app *fiber.App) {
    instanceRouter := app.Group("/role")
    instanceRouter.All("/sync", syncRole)
    instanceRouter.All("/customize", customizeRole)
}

func syncRole(c *fiber.Ctx) error {

    request := &syncRoleRequest{}
    if err := c.BodyParser(request); err != nil {
        log.Errorf("could not parse request: %s", err)
        return fiber.ErrBadRequest
    }

    if request.Finalizing {
        return finalizeRole(c, request)
    } else {
        return reconcileRole(c, request)
    }

}

func reconcileRole(c *fiber.Ctx, request *syncRoleRequest) error {
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
    return c.JSON(response)
}

func finalizeRole(c *fiber.Ctx, request *syncRoleRequest) error {
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

    return c.JSON(response)
}

func customizeRole(c *fiber.Ctx) error {

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
