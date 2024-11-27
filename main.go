package gocroot

import (
	"github.com/gocroot/route"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("WebHook", route.URL)
}