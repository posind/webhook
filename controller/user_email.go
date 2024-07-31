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
/// get paseto token
var pasetoKey = []byte("YELLOW SUBMARINE, BLACK WIZARDRY")

/// save ke database
func SaveUserToDB(user *model.User) error {
    collection := config.DB.Collection("user_email")
    _, err := collection.InsertOne(context.Background(), user)
    return err
}

/// func register
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

    user.ID = primitive.NewObjectID()

    err := SaveUserToDB(&user)
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

/// func login
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

    collection := config.GetCollection("user_email")
    var user model.User
    err := collection.FindOne(context.Background(), bson.M{"email": loginRequest.Email}).Decode(&user)
    if err != nil {
        response := map[string]string{"error": "Invalid email or password"}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(response)
        return
    }

    if user.Password != loginRequest.Password {
        response := map[string]string{"error": "Invalid email or password"}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(response)
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
        response := map[string]string{"error": "Error generating token"}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(response)
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

