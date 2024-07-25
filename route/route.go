package route

import (
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.ErrorMongoconn != nil {
		log.Println(config.ErrorMongoconn.Error())
	}
	
	if config.SetAccessControlHeaders(w, r) && r.Method == http.MethodOptions {
		return // preflight request selesai
	}

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/telebot/:nomorwa"):
		controller.TelebotWebhook(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/endpoint_user/userLogin"):
		controller.Login(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/endpoint_user/userRegister"):
		controller.Register(w, r)

	// CRUD for English version
	case method == "POST" && helper.URLParam(path, "/webhook/crud/item/en"):
		controller.CreateItemEn(w, r)
	case method == "PUT" && helper.URLParam(path, "/webhook/crud/item/en"):
		controller.UpdateItemEn(w, r)
	case method == "GET" && helper.URLParam(path, "/webhook/crud/items/en"):
		controller.GetItemsEn(w, r)
	case method == "DELETE" && helper.URLParam(path, "/webhook/crud/item/en"):
		controller.DeleteItemEn(w, r)

	// CRUD for Indonesian version
	case method == "POST" && helper.URLParam(path, "/webhook/crud/item/id"):
		controller.CreateItemId(w, r)
	case method == "PUT" && helper.URLParam(path, "/webhook/crud/item/id"):
		controller.UpdateItemId(w, r)
	case method == "GET" && helper.URLParam(path, "/webhook/crud/items/id"):
		controller.GetItemsId(w, r)
	case method == "DELETE" && helper.URLParam(path, "/webhook/crud/item/id"):
		controller.DeleteItemId(w, r)

	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
