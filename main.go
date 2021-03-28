package main

import (
	"github.com/orbatschow/kubepost/hook"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	http.HandleFunc("/instance/sync", hook.InstanceSyncHandler)
	http.HandleFunc("/instance/customize", hook.InstanceCustomizeHandler)

	http.HandleFunc("/role/sync", hook.RoleSyncHandler)
	http.HandleFunc("/role/customize", hook.RoleCustomizeHandler)


	http.HandleFunc("/database/sync", hook.DatabaseSyncHandler)
	http.HandleFunc("/database/customize", hook.DatabaseCustomizeHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
