package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

// RegisterHandler menghandle permintaan registrasi admin.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var registrationData model.LoginRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&registrationData)
	if err != nil {
		http.Error(w, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = atdb.InsertOneDoc(config.DB, "users", registrationData)
	if err != nil {
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Registrasi berhasil Jang!",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}


// GetUser mengambil informasi user dari database berdasarkan email dan password.
func GetUser(respw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(respw, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var loginDetails model.User
	if err := json.NewDecoder(req.Body).Decode(&loginDetails); err != nil {
		http.Error(respw, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	var login model.User
	filter := bson.M{"email": loginDetails.Email, "password": loginDetails.Password}
	login, err := atdb.GetOneDoc[model.User](config.DB, "users", filter)
	if err != nil {
		http.Error(respw, "Email atau password salah Jang!", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Login berhasil!",
		"user":    login,
	}
	respw.Header().Set("Content-Type", "application/json")
	respw.WriteHeader(http.StatusOK)
	json.NewEncoder(respw).Encode(response)
}

