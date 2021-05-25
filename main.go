package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/orbatschow/kubepost/hook"
    log "github.com/sirupsen/logrus"
)

func main() {
    app := fiber.New()

    hook.RegisterInstanceHandlerGroup(app)
    hook.RegisterRoleHandlerGroup(app)
    hook.RegisterDatabaseHandlerGroup(app)

    log.Fatal(app.Listen(":8080"))

}
