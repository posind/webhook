package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

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
			existingUser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", primitive.M{"email": userdata.Email})
			if err == nil && existingUser.Email != "" {
				resp.Message = "User sudah terdaftar"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Generate private and public keys
			privateKey, publicKey := watoken.GenerateKey()
			userdata.Private = privateKey
			userdata.Public = publicKey

			// Set password yang sudah di-hash
			userdata.Password = hash

			// Insert data user baru
			_, err = atdb.InsertOneDoc(config.Mongoconn, "user_email", userdata)
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
	var resp model.Credential
	var loginReq model.LoginRequest

	// Decode request body into loginReq
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
	} else {
		// Validasi password
		if passwordhash.PasswordValidator(loginReq) {
			// Mengambil data user berdasarkan email
			userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"email": loginReq.Email})
			if err != nil || userData.Email == "" {
				resp.Message = "User tidak ditemukan"
			} else {
				// Membuat token untuk user
				tokenString, err := passwordhash.EncodeToken(userData.Email, os.Getenv("PRIVATE_KEY"))
				if err != nil {
					resp.Message = "Gagal Encode Token: " + err.Error()
				} else {
					resp.Status = true
					resp.Token = tokenString
					resp.Message = "Login successful"
				}
			}
		} else {
			resp.Message = "Password Salah"
		}
	}

	// Mengirimkan response dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func UpdateUserToken(userID primitive.ObjectID, token string) error {
	// Koneksi ke koleksi "user_email"
	collection := config.Mongoconn.Collection("user_email")

	// Melakukan update pada token user berdasarkan userID
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"token": token}},
	)
	if err != nil {
		// Menambahkan logging yang lebih informatif jika terjadi kesalahan
		log.Printf("Error updating user token for userID %s: %v", userID.Hex(), err)
		return err
	}

	// Jika berhasil, kembalikan nil
	return nil
}
