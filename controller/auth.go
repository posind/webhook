package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/watoken"
	"go.mongodb.org/mongo-driver/bson"
)

func Register(w http.ResponseWriter, r *http.Request) {
	resp := new(model.Credential)
	userdata := new(model.User)
	resp.Status = false

	// Decode request body to userdata struct
	err := json.NewDecoder(r.Body).Decode(&userdata)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
	} else {
		// Check if the phone number is already in use
		existingUser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"phonenumber": userdata.PhoneNumber})
		if err == nil && existingUser.PhoneNumber != "" {
			resp.Message = "Phone number already registered"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Generate private and public keys
		privateKey, publicKey := watoken.GenerateKey()
		userdata.Private = privateKey
		userdata.Public = publicKey

		// Insert new user data without email and username
		_, err = atdb.InsertOneDoc(config.Mongoconn, "user_email", userdata)
		if err != nil {
			resp.Message = "Failed to save user data: " + err.Error()
		} else {
			resp.Status = true
			resp.Message = "User registered successfully"
		}
	}

	// Set response header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var resp model.Credential
	var loginReq model.LoginRequest

	// Parsing JSON request from body (extracting phoneNumber from QR code)
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		sendErrorResponse(w, "Error parsing application/json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate phoneNumber input
	if loginReq.PhoneNumber == "" {
		sendErrorResponse(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Get user data based on phoneNumber from the QR code
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber})
	if err != nil {
		sendErrorResponse(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if user data was found
	if userData.PhoneNumber == "" {
		sendErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate the token using the provided Encode function
	// Ensure that userData.Private and config.PrivateKey are correctly passed
	if userData.Private == "" || config.PrivateKey == "" {
		sendErrorResponse(w, "Invalid data for token generation", http.StatusInternalServerError)
		return
	}

	tokenString, err := watoken.Encode(userData.Private, config.PrivateKey)
	if err != nil {
		sendErrorResponse(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update token in the database
	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	_, err = config.Mongoconn.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"phonenumber": userData.PhoneNumber}, // Match the field correctly
		update,
	)
	if err != nil {
		sendErrorResponse(w, "Failed to update token in MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add token to the response header
	w.Header().Set("login", tokenString) // Set the token in the custom header 'Login'
	w.Header().Set("Content-Type", "application/json")

	// Provide a response after the token has been successfully updated
	resp.Status = true
	resp.Token = tokenString
	resp.Message = "Login successful via QR code"

	// Sending the response
	json.NewEncoder(w).Encode(resp)
}



// sendErrorResponse is a helper function to handle error responses
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

// punya teh fahira

// func Login(w http.ResponseWriter, r *http.Request) {
// 	var resp model.Credential
// 	var loginReq model.LoginRequest

// 	err := json.NewDecoder(r.Body).Decode(&loginReq)
// 	if err != nil {
// 		resp.Message = "Error parsing application/json: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Get user data based on phone number
// 	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber})
// 	if err != nil || userData.PhoneNumber == "" {
// 		resp.Message = "User not found"
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Generate token using user's private key
// 	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
// 	if err != nil {
// 		resp.Message = "Failed to encode token: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Update token in the database using phone number
// 	update := bson.M{
// 		"$set": bson.M{
// 			"token": tokenString,
// 		},
// 	}

// 	_, err = atdb.UpdateOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": userData.PhoneNumber}, update)
// 	if err != nil {
// 		resp.Message = "Failed to update token: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	resp.Status = true
// 	resp.Token = tokenString
// 	resp.Message = "Login successful"

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
