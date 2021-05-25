package hook

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
)

type CustomizeResponse struct {
    RelatedResources []RelatedResource `json:"relatedResources"`
}

type RelatedResource struct {
    ApiVersion string `json:"apiVersion"`
    Resource   string `json:"resource"`
}

func CustomizeHandler(w http.ResponseWriter, r *http.Request, c *CustomizeResponse) {
    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    body, err = json.Marshal(&c)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(body)
}
