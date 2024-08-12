package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleQRCodeScan handles the QR code scan request
func HandleQRCodeScan(w http.ResponseWriter, r *http.Request) {
	collection := config.Mongoconn.Collection("user_email")

	var reqBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := reqBody["email"]
	filter := bson.D{bson.E{Key: "email", Value: email}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			bson.E{Key: "is_logged_in", Value: true},
		}},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Unable to update login status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}


// GenerateQRCode generates a QR code for the user
func GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	collection := config.Mongoconn.Collection("user_email")

	// Decode request body
	var reqBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := reqBody["email"]

	// Check if user exists
	var user model.User_code
	err = collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error finding user", http.StatusInternalServerError)
		}
		return
	}

	// Generate QR code
	qrData := email // Data yang ingin Anda encode ke dalam QR code, misalnya email atau token unik
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Encode QR code to base64
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCode)

	// Update user document with QR code
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "qr_code", Value: qrCodeBase64},
		}},
	}

	_, err = collection.UpdateOne(context.TODO(), bson.M{"email": email}, update)
	if err != nil {
		http.Error(w, "Failed to update user QR code", http.StatusInternalServerError)
		return
	}

	// Send QR code via WhatsApp (pseudo-code, replace with actual API call)
	// Example: sendWhatsAppMessage(user.PhoneNumber, qrCodeBase64)
	// Implementasi tergantung pada layanan API WhatsApp yang Anda gunakan

	// Contoh sederhana: kirimkan QR code ke email (untuk pengujian)
	fmt.Fprintf(w, "QR code generated and sent successfully: %s", qrCodeBase64)
}