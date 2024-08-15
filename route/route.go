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

	// Mendapatkan data pengguna
	case method == "GET" && path == "/get/data/user":
	controller.GetDataUser(w, r)

	// Mengupdate data pengguna berdasarkan scan QR code
	case method == "PUT" && path == "/put/data/user":
	controller.PutTokenDataUser(w, r)

	// Mengupdate data pengguna
	case method == "POST" && path == "/post/data/user":
	controller.PostDataUser(w, r)

	// Mengupdate data pengguna dari WhatsApp
	case method == "POST" && path == "/post/datawa/user":
	controller.PostDataUserFromWA(w, r)	

	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	//jalan setiap jam 3 pagi
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)

	// ProhibitedItem Routes (English)
	case method == "GET" && path == "/get/prohibited-items/en":
		controller.GetProhibitedItemByField(w, r)
	case method == "POST" && path == "/post/prohibited-items/en":
		controller.PostProhibitedItem(w, r)
	case method == "PUT" && path == "/put/prohibited-items/en":
		controller.UpdateProhibitedItem(w, r)
	case method == "DELETE" && path == "/delete/prohibited-items/en":
		controller.DeleteProhibitedItemByField(w, r)

	// ProhibitedItem Routes (Indonesian)
	case method == "GET" && path == "/get/item":
		controller.GetItemByField(w, r)
	case method == "POST" && path == "/post/item":
		controller.PostItem(w, r)
	case method == "PUT" && path == "/put/item":
		controller.UpdateItem(w, r)
	case method == "DELETE" && path == "/delete/item":
		controller.DeleteItemByField(w, r)
	default:
		controller.NotFound(w, r)
	}
}
