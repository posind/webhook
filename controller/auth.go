package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/model"
	"github.com/whatsauth/watoken"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Register(w http.ResponseWriter, r *http.Request) {
	resp := new(model.Credential)
	userdata := new(model.User)
	resp.Status = false

	err := json.NewDecoder(r.Body).Decode(&userdata)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
	} else {
		// Hash password user
		hash, err := passwordhash.HashPassword(userdata.Password)
		if err != nil {
			resp.Message = "Gagal Hash Password: " + err.Error()
		} else {
			// Cek apakah user sudah ada berdasarkan email
			existingUser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "users", primitive.M{"email": userdata.Email})
			if err == nil && existingUser.Email != "" {
				resp.Message = "User sudah terdaftar"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Set password yang sudah di-hash
			userdata.Password = hash

			// Insert data user baru
			_, err = atdb.InsertOneDoc(config.Mongoconn, "users", userdata)
			if err != nil {
				resp.Message = "Gagal menyimpan data user: " + err.Error()
			} else {
				resp.Status = true
				resp.Message = "Berhasil Input data"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := GetUserByEmail(loginRequest.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusNotFound)
		return
	}

	// if !cek.CheckPasswordHash(loginRequest.Password, user.PasswordHash) {
	// 	http.Error(w, "Invalid password", http.StatusBadRequest)
	// 	return
	// }

	token, err := watoken.Encode(user.ID.Hex(), user.Private)
	if err != nil {
		log.Println("Error generating token:", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// user.Token = token
	// err = UpdateUserToken(user.ID, token)
	// if err != nil {
	// 	log.Println("Error updating token:", err)
	// 	http.Error(w, "Error updating token", http.StatusInternalServerError)
	// 	return
	// }

	response := map[string]string{
		"message": "Login successful",
		"token":   token,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UpdateUserToken(userID primitive.ObjectID, token string) error {
	collection := ConnectDB("user_email")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"token": token}},
	)
	if err != nil {
		log.Printf("Error updating user token: %v", err)
	}
	return err
}
