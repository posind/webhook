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

	switch r.Method {
		
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			controller.GetHome(w, r)
		case "/refresh/token":
			controller.GetNewToken(w, r)
		case "/webhook/crud/items/en":
			controller.GetItemsEn(w, r)
		case "/webhook/crud/items/id":
			controller.GetItemsId(w, r)
		default:
			controller.NotFound(w, r)
		}
	case http.MethodPost:
		switch {
		case helper.URLParam(r.URL.Path, "/webhook/nomor/:nomorwa"):
			controller.PostInboxNomor(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/telebot/:nomorwa"):
			controller.TelebotWebhook(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/endpoint_user/user"):
			controller.Register(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/endpoint_user/login"):
			controller.Login(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/en"):
			controller.CreateItemEn(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/id"):
			controller.CreateItemId(w, r)
		default:
			controller.NotFound(w, r)
		}
	case http.MethodPut:
		switch {
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/en"):
			controller.UpdateItemEn(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/id"):
			controller.UpdateItemId(w, r)
		default:
			controller.NotFound(w, r)
		}
	case http.MethodDelete:
		switch {
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/en"):
			controller.DeleteItemEn(w, r)
		case helper.URLParam(r.URL.Path, "/webhook/crud/item/id"):
			controller.DeleteItemId(w, r)
		default:
			controller.NotFound(w, r)
		}
	default:
		controller.NotFound(w, r)
	}
}
