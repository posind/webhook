package gocroot

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/route"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("WebHook", func(w http.ResponseWriter, r *http.Request) {
		if config.SetAccessControlHeaders(w, r) && r.Method == http.MethodOptions {
			return 
		}
		route.URL(w, r)
	})
}