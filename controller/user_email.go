package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocroot/config"
	pwd "github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, err := pwd.HashPass(user.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusBadGateway)
		return
	}
	user.Password = hashedPassword
	user.ID = primitive.NewObjectID()

	if err := SaveUserToDB(&user); err != nil {
		http.Error(w, "Error inserting user", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Registration successful"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
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

	if !pwd.CheckPasswordHash(loginRequest.Password, user.Password) {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	token, err := watoken.Encode(loginRequest.Email, config.PrivateKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	if err := UpdateUserToken(user.ID, token); err != nil {
		http.Error(w, "Error updating token", http.StatusInternalServerError)
		return
	}

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