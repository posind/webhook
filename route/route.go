package route

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return // If it's a preflight request, return early.
	}
	// config.SetEnv()

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "POST" && helper.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/telebot/:nomorwa"):
		controller.TelebotWebhook(w, r)
	case method == "POST" && path == "/register":
		controller.Register(w, r)
	case method == "POST" && path == "/login":
		controller.Login(w, r)

	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	//jalan setiap jam 3 pagi
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)

	case method == "GET" && path == "/prohibited-items/en":
		controller.GetDataProhibitedItemEn(w, r)
	case method == "POST" && path == "/prohibited-items/en":
		controller.CreateProhibitedItemEn(w, r)
	case method == "PUT" && path == "/prohibited-items/en":
		controller.UpdateProhibitedItemEn(w, r)
	case method == "DELETE" && path == "/prohibited-items/en":
		controller.DeleteProhibitedItemEn(w, r)

	// ProhibitedItem Routes (Indonesian)
	case method == "GET" && path == "/prohibited-items/id":
		controller.GetDataProhibitedItemId(w, r)
	case method == "POST" && path == "/prohibited-items/id":
		controller.CreateProhibitedItemId(w, r)
	case method == "PUT" && path == "/prohibited-items/id":
		controller.UpdateProhibitedItemId(w, r)
	case method == "DELETE" && path == "/prohibited-items/id":
		controller.DeleteProhibitedItemId(w, r)
	default:
		controller.NotFound(w, r)
	}
}
