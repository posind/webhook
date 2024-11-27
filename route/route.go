package route

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper/at"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return // If it's a preflight request, return early.
	}
	config.SetEnv()

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "POST" && at.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	case method == "POST" && at.URLParam(path, "/webhook/telebot/:nomorwa"):
		controller.TelebotWebhook(w, r)
	case method == "POST" && path == "/register":
		controller.Register(w, r)
	case method == "POST" && path == "/login":
		controller.Login(w, r)

	//user data
	// case method == "GET" && path == "get/data/user":
	// 	controller.GetDataUser(w, r)
	//generate token linked device
	// case method == "PUT" && path == "put/data/user":
	// 	controller.PutTokenDataUser(w, r)
	// case method == "PUT" && path == "post/data/user":
	// 	controller.PostDataUser(w, r)
	// case method == "PUT" && path == "post/datawa/user":
	// 	controller.PostDataUserFromWA(w, r)

	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	//jalan setiap jam 3 pagi
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)

	// ProhibitedItem Routes (English)
	case method == "GET" && path == "/data/item":
		controller.GetProhibitedItem(w, r)
	case method == "PUT" && path == "/itemid":
		controller.EnsureIDItemExists(w, r)
	case method == "POST" && path == "/post/prohibited-items/en":
		controller.PostProhibitedItem(w, r)
	case method == "PUT" && path == "/data/item":
		controller.UpdateProhibitedItem(w, r)
	case method == "DELETE" && path == "/data/item":
		controller.DeleteProhibitedItem(w, r)

	// ProhibitedItem Routes (Indonesian)
	case method == "GET" && path == "/data/barang":
		controller.GetItemLarangan(w, r)
	case method == "POST" && path == "/data/barang":
		controller.PostItemLarangan(w, r)
	case method == "PUT" && path == "/data/barang":
		controller.UpdateItemLarangan(w, r)
	case method == "DELETE" && path == "/data/barang":
		controller.DeleteItemLarangan(w, r)
	default:
		controller.NotFound(w, r)
	}
}
