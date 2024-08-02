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
	case method == "POST" && path == "/en":
		controller.CreateItemEn(w, r)
	case method == "POST" && path == "/id":
		controller.CreateItemId(w, r)

	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	//jalan setiap jam 3 pagi
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)
	case method == "GET" && path == "/eng":
		controller.GetItemsEn(w, r)
	case method == "GET" && path == "/ind":
		controller.GetItemsId(w, r)

	case method == "PUT" && path == "/webhook/crud/item/en":
		controller.UpdateItemEn(w, r)
	case method == "PUT" && path == "/webhook/crud/item/id":
		controller.UpdateItemId(w, r)

	case method == "DELETE" && path == "/webhook/crud/item/en":
		controller.DeleteItemEn(w, r)
	case method == "DELETE" && path == "/webhook/crud/item/id":
		controller.DeleteItemId(w, r)

	default:
		controller.NotFound(w, r)
	}
}
