package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"github.com/o1egl/paseto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var pasetoKey = []byte("YELLOW SUBMARINE, BLACK WIZARDRY")

func SaveUserToDB(user *model.User) error {
    collection := config.DB.Collection("user_email")
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
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    user.ID = primitive.NewObjectID()

    err := SaveUserToDB(&user)
    if err != nil {
        http.Error(w, "Error inserting user", http.StatusInternalServerError)
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
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var loginRequest model.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    collection := config.GetCollection("user_email")
    var user model.User
    err := collection.FindOne(context.Background(), bson.M{"email": loginRequest.Email}).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    if user.Password != loginRequest.Password {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Generate PASETO token
    now := time.Now()
    expiration := now.Add(24 * time.Hour)
    jsonToken := paseto.JSONToken{
        Expiration: expiration,
        Subject:    user.ID.Hex(),
    }
    footer := "some footer"
    token, err := paseto.NewV2().Encrypt(pasetoKey, jsonToken, footer)
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
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
