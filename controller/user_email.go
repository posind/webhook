package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	pwd "github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SaveUserToDB(user *model.User) error {
	collection := config.Mongoconn.Collection("user_email")
	_, err := collection.InsertOne(context.Background(), user)
	return err
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response := map[string]string{"error": "Invalid request payload"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

// hashingggggg

    var hashedPassword, err = pwd.HashPass(user.Password)
    if err != nil {
		response := map[string]string{"error": "Gagal Hash password!"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(response)
		return
	}
    user.Password = hashedPassword

    user.ID = primitive.NewObjectID()
// saving to mongoDB
	err = SaveUserToDB(&user)
	if err != nil {
		response := map[string]string{"error": "Error inserting user"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"message": "Registration successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := map[string]string{"error": "Method not allowed"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		response := map[string]string{"error": "Invalid request payload"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

    hashcheck, err:= pwd.HashPass(loginRequest.Password)
    if err != nil {
		response := map[string]string{"error": "Gagal Hash password!"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(response)
		return
	}

    checkinghash :=pwd.CheckPasswordHash(loginRequest.Password, hashcheck)
    if !checkinghash{
        response := map[string]string{"error": "hash not valid"}
        w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
    }


	collection := config.GetCollection("user_email")
	var user model.User
	err = collection.FindOne(context.Background(), bson.M{"email": loginRequest.Email}).Decode(&user)
	if err != nil {
		response := map[string]string{"error": "Invalid email or password"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}
    valid := pwd.IsPasswordValid(config.Mongoconn, user)
    if !valid{
        response := map[string]string{"error": "Invalid password brok"}
        w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
    }

    token, err:= watoken.Encode(loginRequest.Email, config.PrivateKey)
    if err != nil {
		response := map[string]string{"error": "Invalid password bre"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

    // untuk mengembalikan token paseto
    user.Token = token
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"token": user.Token}},
	)
	if err != nil {
		response := map[string]string{"error": "Aduh error nih token nya"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"message": "Login sukses tunggu bentar ya!",
		"token":   token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}



	// if user.Password != loginRequest.Password {
	// 	response := map[string]string{"error": "Invalid email or password"}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(response)
	// 	return
	// }

// 	// Generate PASETO token
// 	now := time.Now()
// 	expiration := now.Add(24 * time.Hour)
// 	jsonToken := paseto.JSONToken{
// 		Expiration: expiration,
// 		Subject:    user.ID.Hex(),
// 	}
// 	footer := "some footer"
// 	token, err := paseto.V2().Encrypt(pasetoKey, jsonToken, footer)
// 	if err != nil {
// 		response := map[string]string{"error": "Error generating token"}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	// Update user token
// 	user.Token = token
// 	_, err = collection.UpdateOne(
// 		context.Background(),
// 		bson.M{"_id": user.ID},
// 		bson.M{"$set": bson.M{"token": user.Token}},
// 	)
// 	if err != nil {
// 		response := map[string]string{"error": "Error updating user token"}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	response := map[string]string{
// 		"message": "Login successful",
// 		"token":   token,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(response)
// }
